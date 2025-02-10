package marshalcheck

import (
    "go/ast"
    "go/types"
    "golang.org/x/tools/go/analysis"
    "strings"
)

var eql = func(a, b string) bool { return a == b }

var excludedPackages = map[string]func(string, string) bool{
    "testdata": eql,
    "_test": strings.HasSuffix,
}

var excludedFiles = map[string]func(string, string) bool{
    "zz_generated": strings.Contains,
    "validtor": strings.Contains,
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
        if shouldExcludeResource(pass.Pkg.Path(), excludedPackages) || shouldExcludeResource(fileName, excludedFiles) {
            continue
        }
        ast.Inspect(file, func(n ast.Node) bool {
            typeSpec, ok := n.(*ast.TypeSpec)
            if !ok {
                return true
            }

            structType, ok := typeSpec.Type.(*ast.StructType)
            if !ok {
                return true
            }

            if typeSpec.Name.Name == "Conf" {
                analyzeStructFields(pass, structType, typeSpec.Name.Name, "", true)
            } else {
                analyzeStructFields(pass, structType, typeSpec.Name.Name, "", false)
            }

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
        if len(field.Names) > 0 {
            fieldPath = structName + parentPath + "." + field.Names[0].Name
        }

        if isMergeable && isArrayOrSlice(field, pass) {
            if !isPointer(field) {
                pass.Reportf(field.Pos(), "field %s (mergeable list) must be a pointer to a slice (e.g., *[]T)", fieldPath)
            }
            if !hasOmitEmptyTag(field) {
                pass.Reportf(field.Pos(), "field %s (mergeable list) must have 'omitempty' in JSON tag", fieldPath)
            }
            continue
        }

        if !isMergeable {
            nonMergeableCategory := determineNonMergeableCategory(field)

            switch nonMergeableCategory {
            case "optional_without_default":
                if !isPointer(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, optional, without default) must be a pointer", fieldPath)
                }
                if hasDefaultAnnotation(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, optional, without default) must not have a 'default' annotation", fieldPath)
                }
                if !hasOmitEmptyTag(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, optional, without default) must have 'omitempty' in JSON tag", fieldPath)
                }

            case "optional_with_default":
                if isPointer(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, optional, with default) must not be a pointer", fieldPath)
                }
                if !hasDefaultAnnotation(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, optional, with default) must have a 'default' annotation", fieldPath)
                }
                if hasOmitEmptyTag(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, optional, with default) must not have 'omitempty' in JSON tag", fieldPath)
                }

            case "required":
                if isPointer(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, required) must not be a pointer", fieldPath)
                }
                if hasDefaultAnnotation(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, required) must not have a 'default' annotation", fieldPath)
                }
                if hasOmitEmptyTag(field) {
                    pass.Reportf(field.Pos(), "field %s (non-mergeable, required) must not have 'omitempty' in JSON tag", fieldPath)
                }
            }
            continue
        }

        pass.Reportf(field.Pos(), "field %s must be a pointer with 'omitempty' JSON tag or be inside a slice/array", fieldPath)
    }
}

func determineNonMergeableCategory(field *ast.Field) string {
    hasDefault := hasDefaultAnnotation(field)
    hasOmitEmpty := hasOmitEmptyTag(field)

    if hasOmitEmpty && !hasDefault {
        return "optional_without_default"
    }
    if hasDefault && !hasOmitEmpty {
        return "optional_with_default"
    }
    return "required"
}

func hasDefaultAnnotation(field *ast.Field) bool {
    if field.Tag == nil {
        return false
    }
    tag := strings.Trim(field.Tag.Value, "`")
    return strings.Contains(tag, "default:")
}

func analyzeNestedStruct(pass *analysis.Pass, structIdent *ast.Ident, parentPath string) {
    if obj := pass.TypesInfo.ObjectOf(structIdent); obj != nil {
        if named, ok := obj.Type().(*types.Named); ok {
            if structType, ok := named.Underlying().(*types.Struct); ok {
                for i := 0; i < structType.NumFields(); i++ {
                    field := structType.Field(i)
                    fieldPath := parentPath + "." + field.Name()
                    if !field.Anonymous() {
                        pass.Reportf(obj.Pos(), "field %s must be a pointer with 'omitempty' JSON tag or be inside a slice/array", fieldPath)
                    }
                }
            }
        }
    }
}

func isPointer(field *ast.Field) bool {
    _, ok := field.Type.(*ast.StarExpr)
    return ok
}

func isArrayOrSlice(field *ast.Field, pass *analysis.Pass) bool {
    switch t := field.Type.(type) {
    case *ast.ArrayType:
        return true
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

func reportFieldError(pass *analysis.Pass, field *ast.Field, fieldPath, msg string) {
    pass.Reportf(field.Pos(), "field %s %s", fieldPath, msg)
}
