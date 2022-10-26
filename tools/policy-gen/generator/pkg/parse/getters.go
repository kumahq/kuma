package parse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// inspired by https://github.com/google/go-github/blob/2e864650397abd972ca7092ae8fe5b18c4b3718c/github/gen-accessors.go

type Getter struct {
	sortVal      string
	ReceiverVar  string
	ReceiverType string
	FieldName    string
	FieldType    string
	ZeroValue    string
	PtrField     bool
}

func GettersAndImports(path string) ([]*Getter, map[string]string, error) {
	fset := token.NewFileSet()

	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	getters, imports := processAST(f)
	return getters, imports, nil
}

func processAST(f *ast.File) ([]*Getter, map[string]string) {
	result := []*Getter{}
	imports := map[string]string{}

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
				if se, ok := expr.(*ast.StarExpr); ok {
					expr = se.X
				}

				switch x := expr.(type) {
				case *ast.ArrayType:
					getter := addArrayType(x, ts.Name.String(), fieldName.String())
					if getter != nil {
						result = append(result, getter)
					}
				case *ast.Ident:
					getter := addIdent(x, ts.Name.String(), fieldName.String())
					if getter != nil {
						result = append(result, getter)
					}
				case *ast.MapType:
					getter := addMapType(x, ts.Name.String(), fieldName.String())
					if getter != nil {
						result = append(result, getter)
					}
				case *ast.SelectorExpr:
					getter, imprts := addSelectorExpr(x, ts.Name.String(), fieldName.String())
					if getter != nil {
						result = append(result, getter)
					}
					for k, v := range imprts {
						imports[k] = v
					}
				}
			}
		}
	}

	return result, imports
}

func addArrayType(x *ast.ArrayType, receiverType, fieldName string) *Getter {
	return newGetter(receiverType, fieldName, fieldType(x), "nil", false)
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
		return x.Sel.String() + "." + fieldType(x.X)
	}
	return ""
}

func addIdent(x *ast.Ident, receiverType, fieldName string) *Getter {
	expr := x
	if x.Obj != nil {
		if ts, ok := x.Obj.Decl.(*ast.TypeSpec); ok {
			switch e := ts.Type.(type) {
			case *ast.StructType:
				return newGetter(receiverType, fieldName, x.String(), "nil", true)
			case *ast.Ident:
				expr = e
			}
		}
	}
	var zeroValue string
	ptrField := false
	switch expr.String() {
	case "int", "int32", "int64":
		zeroValue = "0"
	case "string":
		zeroValue = `""`
	case "bool":
		zeroValue = "false"
	default:
		zeroValue = "nil"
		ptrField = true
	}
	return newGetter(receiverType, fieldName, x.String(), zeroValue, ptrField)
}

func addMapType(x *ast.MapType, receiverType, fieldName string) *Getter {
	var keyType string
	switch key := x.Key.(type) {
	case *ast.Ident:
		keyType = key.String()
	default:
		return nil
	}

	var valueType string
	switch value := x.Value.(type) {
	case *ast.Ident:
		valueType = value.String()
	default:
		return nil
	}

	fieldType := fmt.Sprintf("map[%v]%v", keyType, valueType)
	zeroValue := fmt.Sprintf("map[%v]%v{}", keyType, valueType)
	return newGetter(receiverType, fieldName, fieldType, zeroValue, false)
}

func addSelectorExpr(x *ast.SelectorExpr, receiverType, fieldName string) (*Getter, map[string]string) {
	if strings.ToLower(fieldName[:1]) == fieldName[:1] { // Non-exported field.
		return nil, nil
	}

	var xX string
	if xx, ok := x.X.(*ast.Ident); ok {
		xX = xx.String()
	}

	imprts := map[string]string{}
	switch xX {
	case "time", "json":
		if xX == "json" {
			imprts["encoding/json"] = "encoding/json"
		} else {
			imprts[xX] = xX
		}
		fieldType := fmt.Sprintf("%v.%v", xX, x.Sel.Name)
		zeroValue := fmt.Sprintf("%v.%v{}", xX, x.Sel.Name)
		if xX == "time" && x.Sel.Name == "Duration" {
			zeroValue = "0"
		}
		return newGetter(receiverType, fieldName, fieldType, zeroValue, false), imprts
	case "common_api":
		imprts["common_api"] = "github.com/kumahq/kuma/api/common/v1alpha1"
		fieldType := fmt.Sprintf("%v.%v", xX, x.Sel.Name)
		zeroValue := fmt.Sprintf("%v.%v{}", xX, x.Sel.Name)
		zeroValue = "nil"
		return newGetter(receiverType, fieldName, fieldType, zeroValue, true), imprts
	default:
		return nil, nil
	}
}

func newGetter(receiverType, fieldName, fieldType, zeroValue string, ptrField bool) *Getter {
	return &Getter{
		sortVal:      strings.ToLower(receiverType) + "." + strings.ToLower(fieldName),
		ReceiverVar:  strings.ToLower(receiverType[:1]),
		ReceiverType: receiverType,
		FieldName:    fieldName,
		FieldType:    fieldType,
		ZeroValue:    zeroValue,
		PtrField:     ptrField,
	}
}
