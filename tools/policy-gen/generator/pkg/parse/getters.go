package parse

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
)

type Getter struct {
	ReceiverVar  string
	ReceiverType string
	FieldName    string
	FieldType    string
	ZeroValue    string
}

func GettersAndImports(path string) ([]*Getter, []string, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, nil, err
	}

	if len(pkgs) != 1 {
		return nil, nil, errors.New("expected 1 package")
	}

	pkg := pkgs[0]
	info := pkg.TypesInfo

	imports := []string{}
	for _, t := range pkg.Syntax[0].Imports {
		if t.Name != nil {
			imports = append(imports, fmt.Sprintf("%s %s", t.Name.String(), t.Path.Value))
		} else {
			imports = append(imports, t.Path.Value)
		}
	}

	return processAST(pkgs[0].Syntax[0], info), imports, nil
}

func processAST(f *ast.File, info *types.Info) []*Getter {
	result := []*Getter{}

	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			// Skip unexported identifiers.
			if !ts.Name.IsExported() {
				continue
			}

			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range st.Fields.List {
				if len(field.Names) == 0 {
					continue
				}

				fieldName := field.Names[0]
				// Skip unexported identifiers.
				if !fieldName.IsExported() {
					continue
				}

				expr := field.Type
				result = append(result, &Getter{
					ReceiverVar:  strings.ToLower(ts.Name.String()[:1]),
					ReceiverType: ts.Name.String(),
					FieldName:    fieldName.String(),
					FieldType:    fieldType(expr),
					ZeroValue:    fieldZero(info.TypeOf(expr), ""),
				})
			}
		}
	}

	return result
}

func fieldType(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.String()
	case *ast.ArrayType:
		return "[]" + fieldType(x.Elt)
	case *ast.StarExpr:
		return "*" + fieldType(x.X)
	case *ast.SelectorExpr:
		return fieldType(x.X) + "." + x.Sel.String()
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", fieldType(x.Key), fieldType(x.Value))
	}
	panic(fmt.Sprintf("unknown type %v", reflect.TypeOf(expr)))
}

func fieldZero(t types.Type, name string) string {
	switch tt := t.(type) {
	case *types.Basic:
		switch {
		case (tt.Info() & types.IsNumeric) != 0:
			return "0"
		case (tt.Info() & types.IsString) != 0:
			return `""`
		case (tt.Info() & types.IsBoolean) != 0:
			return "false"
		}
	case *types.Struct:
		return name + "{}"
	case *types.Named:
		return fieldZero(tt.Underlying(), tt.Obj().Name())
	case *types.Pointer, *types.Slice, *types.Map:
		return "nil"
	}
	panic(fmt.Sprintf("unknown type %v", reflect.TypeOf(t)))
}
