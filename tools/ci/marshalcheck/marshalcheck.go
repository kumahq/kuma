package marshalcheck

import (
    "go/ast"
    "go/types"
    "golang.org/x/tools/go/analysis"
    "strings"
)

var Analyzer = &analysis.Analyzer{
    Name: "marshalcheck",
    Doc:  "checks that struct fields follow proper serialization rules",
    Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
    for _, file := range pass.Files {
        // Exclude testdata packages
        if pass.Pkg.Path() != "" && strings.Contains(pass.Pkg.Path(), "testdata") {
            continue
        }
        ast.Inspect(file, func(n ast.Node) bool {
            // Look for struct type declarations
            typeSpec, ok := n.(*ast.TypeSpec)
            if !ok || typeSpec.Name.Name != "Conf" {
                return true
            }

            structType, ok := typeSpec.Type.(*ast.StructType)
            if !ok {
                return true
            }

            analyzeStructFields(pass, structType, typeSpec.Name.Name, "")
            return false
        })
    }
    return nil, nil
}

func analyzeStructFields(pass *analysis.Pass, structType *ast.StructType, structName string, parentPath string) {
    for _, field := range structType.Fields.List {
        fieldPath := parentPath
        if len(field.Names) > 0 {
            fieldPath = parentPath + "." + field.Names[0].Name
        }

        if isPointer(field) {
            if !hasOmitEmptyTag(field) {
                reportFieldError(pass, field, fieldPath, "field must have 'omitempty' in JSON tag")
            }
        } else if isArrayOrSlice(field, pass) {
            if hasOmitEmptyTag(field) {
                reportFieldError(pass, field, fieldPath, "field inside a slice/array should not have 'omitempty'")
            }
            if elemType, ok := field.Type.(*ast.ArrayType); ok {
                if structIdent, ok := elemType.Elt.(*ast.Ident); ok {
                    analyzeNestedStruct(pass, structIdent, fieldPath+"[]")
                }
            }
        } else {
            pass.Reportf(field.Pos(), "field %s must be a pointer with 'omitempty' JSON tag or be inside a slice/array", fieldPath)
        }
    }
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
    tag := field.Tag.Value
    tag = strings.Trim(tag, "`")
    return strings.Contains(tag, "json:") && strings.Contains(tag, "omitempty")
}

func reportFieldError(pass *analysis.Pass, field *ast.Field, fieldPath, msg string) {
    pass.Reportf(field.Pos(), "field %s %s", fieldPath, msg)
}
