package marshalcheck

import (
    "go/ast"
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
                // Check if field type is a pointer or a slice/array
                switch field.Type.(type) {
                case *ast.StarExpr:
                    // Pointer type, ensure 'omitempty' is in JSON tag
                    if !hasOmitEmptyTag(field) {
                        if len(field.Names) > 0 {
                            pass.Reportf(field.Pos(), "field '%s' must have 'omitempty' in JSON tag", field.Names[0].Name)
                        } else {
                            pass.Reportf(field.Pos(), "embedded field must have 'omitempty' in JSON tag")
                        }
                    }
                case *ast.ArrayType:
                    // Allowed, as it's an array/slice
                    continue
                default:
                    // If not a pointer or slice, report an error
                    if len(field.Names) > 0 {
                        pass.Reportf(field.Pos(), "field '%s' must be a pointer with 'omitempty' JSON tag or a slice/array", field.Names[0].Name)
                    } else {
                        pass.Reportf(field.Pos(), "embedded field must be a pointer with 'omitempty' JSON tag or a slice/array")
                    }
                }
            }
            return false
        })
    }
    return nil, nil
}

func hasOmitEmptyTag(field *ast.Field) bool {
    if field.Tag == nil {
        return false
    }
    tag := field.Tag.Value
    // Strip backticks
    tag = strings.Trim(tag, "`")

    // Look for json:"...,omitempty"
    return strings.Contains(tag, "json:") && strings.Contains(tag, "omitempty")
}
