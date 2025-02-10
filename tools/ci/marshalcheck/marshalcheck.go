package marshalcheck

import (
    "fmt"
    "go/ast"
    "go/types"
    "golang.org/x/tools/go/analysis"
    "path/filepath"
    "strings"
)

var eql = func(a, b string) bool { return a == b }

var excludedPackages = map[string]func(string, string) bool{
    "testdata": eql,
    "_test": strings.HasSuffix,
}

var excludedFiles = map[string]func(string, string) bool{
    "zz_generated": strings.Contains,
    "validator": strings.Contains,
    "compare": strings.Contains,
}

var Analyzer = &analysis.Analyzer{
    Name: "marshalcheck",
    Doc:  "checks that struct fields follow proper serialization rules",
    Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
    for _, file := range pass.Files {
        fileName := pass.Fset.File(file.Pos()).Name()
        fileNameWithoutExtension := stripExtension(fileName)
        if shouldExcludeResource(pass.Pkg.Path(), excludedPackages) || shouldExcludeResource(fileName, excludedFiles) {
            continue
        }
        ast.Inspect(file, func(n ast.Node) bool {
            typeSpec, ok := n.(*ast.TypeSpec)
            if !ok {
                return true
            }

            if strings.ToLower(typeSpec.Name.String()) != fileNameWithoutExtension {
                fmt.Println("DEBUG: Skipping type", typeSpec.Name.String(), "in file", fileNameWithoutExtension)
                return true
            }

            structType, ok := typeSpec.Type.(*ast.StructType)
            if !ok {
                return true
            }

            analyzeStructFields(pass, structType, typeSpec.Name.Name, "", false)

            return false
        })
    }
    return nil, nil
}

func shouldExcludeResource(name string, rules map[string] func(string, string) bool) bool {
    for pattern, matchFunc := range rules {
        if matchFunc(name, pattern) {
            return true
        }
    }
    return false
}

func analyzeStructFields(pass *analysis.Pass, structType *ast.StructType, structName string, parentPath string, isMergeable bool) {
    for _, field := range structType.Fields.List {
        fieldPath := structName + parentPath
        fieldName := field.Names[0].Name
        if len(field.Names) > 0 {
            fieldPath = structName + parentPath + "." + fieldName
        }
        fmt.Println("DEBUG: Analyzing field", fieldPath)

        // ✅ Handle pointers to structs (*Struct)
        baseType := field.Type
        if ptrType, ok := field.Type.(*ast.StarExpr); ok {
            baseType = ptrType.X // Unwrap pointer
        }

        // ✅ Recursively analyze named nested structs
        if ident, ok := baseType.(*ast.Ident); ok {
            namedStruct := findStructByName(pass, ident.Name)
            if namedStruct != nil {
                analyzeStructFields(pass, namedStruct, ident.Name, fieldPath, isMergeable)
                continue
            }
        }

        // ✅ Handle pointers to slices (*[]T)
        if arrayType, ok := baseType.(*ast.ArrayType); ok {
            if elemIdent, ok := arrayType.Elt.(*ast.Ident); ok {
                namedStruct := findStructByName(pass, elemIdent.Name)
                if namedStruct != nil {
                    analyzeStructFields(pass, namedStruct, elemIdent.Name, fieldPath+"[]", isMergeable)
                }
            }
        }

        // Process the field normally
        if isMergeable {
            if !isPointer(field) {
                pass.Reportf(field.Pos(), "mergeable field %s must be a pointer", fieldPath)
            }
            if !hasOmitEmptyTag(field) {
                pass.Reportf(field.Pos(), "mergeable field %s must have 'omitempty' in JSON tag", fieldPath)
            }
            // todo: doesn't have default annotation
        } else {
            _, isValid := determineNonMergeableCategory(pass, field)
            //fmt.Println("DEBUG: ", fieldPath, " detected as ", category)  // Add this to check the extracted tag value

            if !isValid {
                pass.Reportf(field.Pos(), "field %s does not match any allowed non-mergeable category", fieldPath)
            }
        }
    }
}


func findStructByName(pass *analysis.Pass, structName string) *ast.StructType {
    var foundStruct *ast.StructType

    for _, file := range pass.Files {
        ast.Inspect(file, func(n ast.Node) bool {
            typeSpec, ok := n.(*ast.TypeSpec)
            if !ok {
                return true
            }
            if typeSpec.Name.Name != structName {
                return true
            }
            if structType, ok := typeSpec.Type.(*ast.StructType); ok {
                foundStruct = structType // Save the struct in the variable
                return false             // Stop further inspection
            }
            return true
        })
        if foundStruct != nil {
            return foundStruct
        }
    }
    return nil
}

func isPointerToSlice(field *ast.Field) bool {
    if ptr, ok := field.Type.(*ast.StarExpr); ok {
        _, isArray := ptr.X.(*ast.ArrayType)
        return isArray
    }
    return false
}

func determineNonMergeableCategory(pass *analysis.Pass, field *ast.Field) (string, bool) {
    hasDefault := hasDefaultAnnotation(pass, field)
    hasOmitEmpty := hasOmitEmptyTag(field)
    isPtr := isPointer(field)

    if isPtr && hasOmitEmpty && !hasDefault {
        return "optional_without_default", true
    }
    if !isPtr && hasDefault && !hasOmitEmpty {
        return "optional_with_default", true
    }
    if !isPtr && !hasDefault && !hasOmitEmpty {
        return "required", true
    }
    return "", false
}

func hasDefaultAnnotation(pass *analysis.Pass, field *ast.Field) bool {
    if pass.Files == nil {
        return false
    }

    fieldPos := field.Pos() // The position of the field in the source file
    for _, commentGroup := range pass.Files {
        for _, comment := range commentGroup.Comments {
            if comment.Pos() < fieldPos { // Ensure comment appears before the field
                for _, c := range comment.List {
                    if strings.Contains(c.Text, "+kubebuilder:default") {
                        return true
                    }
                }
            } else {
                // If we found a comment AFTER the field, stop checking (since we only want the last comment before it)
                break
            }
        }
    }
    return false
}

func isPointer(field *ast.Field) bool {
    _, ok := field.Type.(*ast.StarExpr)
    return ok
}

func isArrayOrSlice(field *ast.Field, pass *analysis.Pass) bool {
    switch t := field.Type.(type) {
    case *ast.ArrayType:
        return true
    case *ast.StarExpr: // Handle pointers to slices (e.g., *[]T)
        if _, ok := t.X.(*ast.ArrayType); ok {
            return true
        }
    case *ast.Ident:
        if pass.TypesInfo != nil {
            typeObj := pass.TypesInfo.ObjectOf(t)
            if typeObj != nil {
                _, ok := typeObj.Type().Underlying().(*types.Slice)
                return ok
            }
        }
    }
    return false
}

func hasOmitEmptyTag(field *ast.Field) bool {
    if field.Tag == nil {
        return false
    }
    tag := strings.Trim(field.Tag.Value, "`")
    return strings.Contains(tag, "json:") && strings.Contains(tag, "omitempty")
}

// Extracts the struct name from the filename
func stripExtension(filename string) string {
    base := filepath.Base(filename)
    name := strings.TrimSuffix(base, ".go")
    return name
}
