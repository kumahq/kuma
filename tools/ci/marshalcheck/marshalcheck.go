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

            for _, field := range structType.Fields.List {
                if isPointer(field) {
                    if !hasOmitEmptyTag(field) {
                        reportFieldError(pass, field, "field must have 'omitempty' in JSON tag")
                    }
                } else if isArrayOrSlice(field, pass) {
                    if hasOmitEmptyTag(field) {
                        reportFieldError(pass, field, "field inside a slice/array should not have 'omitempty'")
                    }
                } else {
                    pass.Reportf(field.Pos(), "field must be a pointer with 'omitempty' JSON tag or be inside a slice/array")
                }
            }
            return false
        })
    }
    return nil, nil
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

func reportFieldError(pass *analysis.Pass, field *ast.Field, msg string) {
    if len(field.Names) > 0 {
        pass.Reportf(field.Pos(), "field '%s' %s", field.Names[0].Name, msg)
    } else {
        pass.Reportf(field.Pos(), "embedded field %s", msg)
    }
}
