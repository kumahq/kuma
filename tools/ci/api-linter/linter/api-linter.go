package linter

import (
	"flag"
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
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
	// opaqueAnnotation marks raw JSON that is the user's own data rather than the
	// configuration of an extension, so it has no schema to document.
	opaqueAnnotation = "+kuma:opaque-payload"
)

const (
	rawJSONPackage     = "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	rawJSONType        = "JSON"
	extensionsPackage  = "github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
	extensionPointType = "Point"
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

func run(pass *analysis.Pass) (any, error) {
	points := collectExtensionPoints(pass)
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

			// "spec" roots the JSON path because an extension point is addressed
			// from the resource's item schema, where the API sits under spec.
			analyzeStructFields(pass, points, structType, typeSpec.Name.Name, []string{"spec"}, false)
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

// analyzeStructFields walks a resource's fields. jsonPath is where the current
// struct sits in the item schema, or nil once the walk passes through something an
// extension point cannot address, such as a slice.
func analyzeStructFields(pass *analysis.Pass, points []extensionPoint, structType *ast.StructType, parentPath string, jsonPath []string, isMergeable bool) {
	for _, field := range structType.Fields.List {
		fieldName := ""
		if len(field.Names) != 1 {
			var names []string
			for _, name := range field.Names {
				names = append(names, name.Name)
			}
			if !hasAnnotations(field, nolintAnnotation) {
				pass.Reportf(field.Pos(), "field in struct %s must have exactly one name, got '%s'", parentPath, strings.Join(names, ","))
			}
			continue
		}
		fieldName = field.Names[0].Name
		if fieldName == "Default" {
			isMergeable = true
		}
		fieldPath := parentPath + "." + fieldName

		var fieldJSONPath []string
		if jsonPath != nil {
			if name := jsonName(field); name != "" {
				fieldJSONPath = append(append([]string{}, jsonPath...), name)
			}
		}

		if *debugLog {
			fmt.Println("DEBUG: Analyzing field", fieldPath)
		}

		checkRawJSON(pass, points, field, fieldPath, jsonPath)

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
					isMergeable = false
				}
				analyzeStructFields(pass, points, namedStruct, fieldPath, fieldJSONPath, isMergeable)
			}
		}

		// Handle slices ([]T)
		if arrayType, ok := baseType.(*ast.ArrayType); ok {
			if elemIdent, ok := arrayType.Elt.(*ast.Ident); ok {
				namedStruct := findStructByName(pass, elemIdent.Name)
				if namedStruct != nil {
					// nil: an extension point addresses properties, not elements.
					analyzeStructFields(pass, points, namedStruct, fieldPath+"[]", nil, false)
				}
			}

			// Resolve the type of the slice element using type information
			elemType := pass.TypesInfo.TypeOf(arrayType.Elt)

			// Dereference named types to get to the struct
			if named, ok := elemType.(*types.Named); ok {
				namedStruct := findStructByName(pass, named.String())
				if namedStruct != nil {
					analyzeStructFields(pass, points, namedStruct, fieldPath+"[]", nil, false)
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

// extensionPoint is an extensions.Point declared in the package under analysis.
type extensionPoint struct {
	schemaPath     []string
	configProperty string
}

// checkRawJSON requires every raw JSON field to say what it is. Left undeclared it
// reaches the OpenAPI spec as "anything at all": either an extension point
// describes the configurations that may go in it, or it is the user's own opaque
// data and says so. Forgetting is what leaves a documented API with a hole in it.
func checkRawJSON(pass *analysis.Pass, points []extensionPoint, field *ast.Field, fieldPath string, parentJSONPath []string) {
	// hasAnnotations requires all of them, so ask separately: either marker is
	// enough on its own.
	if !isRawJSON(pass, field) || hasAnnotations(field, opaqueAnnotation) || hasAnnotations(field, nolintAnnotation) {
		return
	}

	property := jsonName(field)
	if parentJSONPath == nil || property == "" {
		pass.Reportf(field.Pos(), "raw JSON field %s cannot be reached by an extension point, so it must be marked '%s'",
			fieldPath, opaqueAnnotation)
		return
	}

	for _, point := range points {
		if point.configProperty == property && slices.Equal(point.schemaPath, parentJSONPath) {
			return
		}
	}
	pass.Reportf(field.Pos(),
		"raw JSON field %s is undocumented: declare an extensions.Point with SchemaPath []string{%s} and ConfigProperty %q, or mark the field '%s'",
		fieldPath, `"`+strings.Join(parentJSONPath, `", "`)+`"`, property, opaqueAnnotation)
}

func isRawJSON(pass *analysis.Pass, field *ast.Field) bool {
	fieldType := pass.TypesInfo.TypeOf(field.Type)
	if fieldType == nil {
		return false
	}
	if ptr, ok := fieldType.(*types.Pointer); ok {
		fieldType = ptr.Elem()
	}
	named, ok := fieldType.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Name() == rawJSONType && named.Obj().Pkg().Path() == rawJSONPackage
}

// collectExtensionPoints reads the extensions.Point values the package declares.
// They are read from the syntax rather than evaluated, which is enough because a
// point is a literal sitting next to the resource it describes.
func collectExtensionPoints(pass *analysis.Pass) []extensionPoint {
	var points []extensionPoint
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok || !isExtensionPointType(pass, lit) {
				return true
			}
			point := extensionPoint{}
			for _, elt := range lit.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}
				switch key.Name {
				case "SchemaPath":
					point.schemaPath = stringLiterals(kv.Value)
				case "ConfigProperty":
					if values := stringLiterals(kv.Value); len(values) == 1 {
						point.configProperty = values[0]
					}
				}
			}
			points = append(points, point)
			return true
		})
	}
	return points
}

func isExtensionPointType(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	litType := pass.TypesInfo.TypeOf(lit.Type)
	named, ok := litType.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Name() == extensionPointType && named.Obj().Pkg().Path() == extensionsPackage
}

// stringLiterals reads a string literal or a slice of them, and reports nothing
// for anything computed, which a point declaration has no reason to be.
func stringLiterals(expr ast.Expr) []string {
	if basic, ok := expr.(*ast.BasicLit); ok {
		if value, err := strconv.Unquote(basic.Value); err == nil {
			return []string{value}
		}
		return nil
	}
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(lit.Elts))
	for _, elt := range lit.Elts {
		basic, ok := elt.(*ast.BasicLit)
		if !ok {
			return nil
		}
		value, err := strconv.Unquote(basic.Value)
		if err != nil {
			return nil
		}
		values = append(values, value)
	}
	return values
}

// jsonName is the name the field takes in the schema.
func jsonName(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	jsonTag, ok := reflect.StructTag(strings.Trim(field.Tag.Value, "`")).Lookup("json")
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(jsonTag, ",")
	if name == "-" {
		return ""
	}
	return name
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

	var comments strings.Builder
	for _, line := range field.Doc.List {
		comments.WriteString(line.Text + "\n")
	}

	for _, requiredAnnotation := range requiredAnnotations {
		if !strings.Contains(comments.String(), requiredAnnotation) {
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
	return slices.Contains(tagParts, "omitempty")
}

// Extracts the struct name from the filename
func stripExtension(filename string) string {
	base := filepath.Base(filename)
	name := strings.TrimSuffix(base, ".go")
	return name
}
