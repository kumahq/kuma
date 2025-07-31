package linter

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"reflect"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	defaultAnnotation = "+kubebuilder:default"
	// we need both annotations otherwise kubebuilder creates OAPI schema with both "required" and "default" which
	// according to the spec https://swagger.io/docs/specification/v3_0/describing-parameters/#default-parameter-values
	// is invalid: "There are two common mistakes when using the default keyword: Using default with required parameters or properties"
	optionalAnnotation      = "+kubebuilder:validation:Optional"
	nonMergableAnotation    = "+kuma:non-mergeable-struct"
	discriminatorAnnotation = "+kuma:discriminator"
	nolintAnnotation        = "+kuma:nolint"
)

var eql = func(a, b string) bool { return a == b }

var excludedPackages = map[string]func(string, string) bool{
	"testdata": eql,
	"_test":    strings.HasSuffix,
}

var excludedFiles = map[string]func(string, string) bool{
	"zz_generated": strings.Contains,
	"validator":    strings.Contains,
	"compare":      strings.Contains,
}

var debugLog *bool

var Analyzer = &analysis.Analyzer{
	Name:  "apilinter",
	Doc:   "checks that struct fields follow proper serialization rules",
	Run:   run,
	Flags: flags(),
}

func flags() flag.FlagSet {
	set := flag.NewFlagSet("", flag.ExitOnError)
	debugLog = set.Bool("debugLog", false, "print debug logs")
	return *set
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		fileName := pass.Fset.File(file.Pos()).Name()
		fileNameWithoutExtension := stripExtension(fileName)
		if shouldExcludeResource(pass.Pkg.Path(), excludedPackages) || shouldExcludeResource(fileName, excludedFiles) {
			if *debugLog {
				fmt.Println("DEBUG: Skipping file", fileName, "in package", pass.Pkg.Path())
			}
			continue
		}
		hasRunForFile := false
		ast.Inspect(file, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			if strings.ToLower(typeSpec.Name.String()) != fileNameWithoutExtension {
				if *debugLog {
					fmt.Println("DEBUG: Skipping type", typeSpec.Name.String(), "in file", fileNameWithoutExtension)
				}
				return true
			}

			analyzeStructFields(pass, structType, typeSpec.Name.Name, false)
			hasRunForFile = true

			return false
		})
		if !hasRunForFile {
			if *debugLog {
				fmt.Println("DEBUG: No struct with the same name as filename found in file", fileNameWithoutExtension)
			}
		}
	}
	return nil, nil
}

func shouldExcludeResource(name string, rules map[string]func(string, string) bool) bool {
	for pattern, matchFunc := range rules {
		if matchFunc(name, pattern) {
			return true
		}
	}
	return false
}

func analyzeStructFields(pass *analysis.Pass, structType *ast.StructType, parentPath string, isMergeable bool) {
	for _, field := range structType.Fields.List {
		fieldName := ""
		if len(field.Names) != 1 {
			var names []string
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			if !hasAnnotations(field, nolintAnnotation) {
				pass.Reportf(field.Pos(), "field in struct %s must have exactly one name, got [%s]", parentPath, strings.Join(names, ","))
			}
			continue
		}
		fieldName = field.Names[0].Name
		if fieldName == "Default" {
			isMergeable = true
		}
		fieldPath := parentPath + "." + fieldName

		if *debugLog {
			fmt.Println("DEBUG: Analyzing field", fieldPath)
		}

		// Handle pointers to structs (*Struct)
		baseType := field.Type
		if ptrType, ok := field.Type.(*ast.StarExpr); ok {
			baseType = ptrType.X // Unwrap pointer
		}

		// Recursively analyze named nested structs
		if ident, ok := baseType.(*ast.Ident); ok {
			namedStruct := findStructByName(pass, ident.Name)
			if namedStruct != nil {
				if hasAnnotations(field, nonMergableAnotation) {
					analyzeStructFields(pass, namedStruct, fieldPath, false)
				} else {
					analyzeStructFields(pass, namedStruct, fieldPath, isMergeable)
				}
			}
		}

		// Handle slices ([]T)
		if arrayType, ok := baseType.(*ast.ArrayType); ok {
			if elemIdent, ok := arrayType.Elt.(*ast.Ident); ok {
				namedStruct := findStructByName(pass, elemIdent.Name)
				if namedStruct != nil {
					analyzeStructFields(pass, namedStruct, fieldPath+"[]", false)
				}
			}

			// Resolve the type of the slice element using type information
			elemType := pass.TypesInfo.TypeOf(arrayType.Elt)

			// Dereference named types to get to the struct
			if named, ok := elemType.(*types.Named); ok {
				namedStruct := findStructByName(pass, named.String())
				if namedStruct != nil {
					analyzeStructFields(pass, namedStruct, fieldPath+"[]", false)
				}
			}
		}

		// we do not lint default field
		if fieldName == "Default" {
			continue
		}

		if hasAnnotations(field, nolintAnnotation) {
			if *debugLog {
				fmt.Println("DEBUG: skipping field do to "+nolintAnnotation, fieldPath)
			}
			continue
		}

		// Process the field normally
		if isMergeable {
			if isKumaDiscriminator(field) {
				continue
			}
			if !isPointer(field) {
				pass.Reportf(field.Pos(), "mergeable field %s must be a pointer", fieldPath)
			}
			if !hasOmitEmptyTag(field) {
				pass.Reportf(field.Pos(), "mergeable field %s must have 'omitempty' in JSON tag", fieldPath)
			}
			if hasAnnotations(field, defaultAnnotation, optionalAnnotation) {
				pass.Reportf(field.Pos(), "mergeable field %s must not have '%s' annotation(s)", fieldPath, defaultAnnotation+", "+optionalAnnotation)
			}
		} else {
			category, isValid := determineNonMergeableCategory(field)

			if *debugLog {
				fmt.Println("DEBUG: Field", fieldPath, "is in non-mergeable category:", category)
			}

			if !isValid {
				pass.Reportf(field.Pos(), "field %s does not match any allowed non-mergeable category", fieldPath)
			}
		}
	}
}

var CommonTypes = map[string]*ast.StructType{}

func findStructByName(pass *analysis.Pass, structName string) *ast.StructType {
	var foundStruct *ast.StructType

	if CommonTypes[structName] != nil {
		if *debugLog {
			fmt.Println("DEBUG: Found struct in CommonTypes:", structName)
		}
		return CommonTypes[structName]
	}

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

func isKumaDiscriminator(field *ast.Field) bool {
	hasKumaDiscriminator := hasAnnotations(field, discriminatorAnnotation)
	hasDefaultAndOptional := hasAnnotations(field, defaultAnnotation, optionalAnnotation)
	hasOmitEmpty := hasOmitEmptyTag(field)
	isPtr := isPointer(field)
	return hasKumaDiscriminator && !isPtr && !hasDefaultAndOptional && !hasOmitEmpty
}

func determineNonMergeableCategory(field *ast.Field) (string, bool) {
	hasDefault := hasAnnotations(field, defaultAnnotation)
	hasOptional := hasAnnotations(field, optionalAnnotation)
	hasDefaultAndOptional := hasDefault && hasOptional
	hasOmitEmpty := hasOmitEmptyTag(field)
	isPtr := isPointer(field)

	if hasDefault && !hasOptional {
		return "missing_optional_annotation", false
	}
	if isPtr && hasOmitEmpty && !hasDefaultAndOptional {
		return "optional_without_default", true
	}
	if !isPtr && hasDefaultAndOptional && !hasOmitEmpty {
		return "optional_with_default", true
	}
	if !isPtr && !hasDefaultAndOptional && !hasOmitEmpty {
		return "required", true
	}
	return "", false
}

func hasAnnotations(field *ast.Field, requiredAnnotations ...string) bool {
	if field.Doc == nil {
		return false
	}

	comments := ""
	for _, line := range field.Doc.List {
		comments += line.Text + "\n"
	}

	for _, requiredAnnotation := range requiredAnnotations {
		if !strings.Contains(comments, requiredAnnotation) {
			return false
		}
	}
	return true
}

func isPointer(field *ast.Field) bool {
	_, ok := field.Type.(*ast.StarExpr)
	return ok
}

func hasOmitEmptyTag(field *ast.Field) bool {
	if field.Tag == nil {
		return false
	}

	// Extract the struct tag (removes surrounding backticks)
	tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))

	// Parse JSON tag
	jsonTag, ok := tag.Lookup("json")
	if !ok {
		return false
	}

	// Check if "omitempty" is in the tag
	tagParts := strings.Split(jsonTag, ",")
	for _, part := range tagParts {
		if part == "omitempty" {
			return true
		}
	}

	return false
}

// Extracts the struct name from the filename
func stripExtension(filename string) string {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, ".go")
	return name
}
