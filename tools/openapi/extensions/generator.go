// Package extensions documents registered resource extensions in an OpenAPI spec.
//
// The extension points on Kuma resources take a freeform `config`, so the
// generated spec says nothing about what may go in it. This generator takes the
// Go types registered with pkg/core/resources/extensions, runs them through
// controller-gen the same way resource specs are generated, and patches the result
// into an existing OpenAPI document as a oneOf keyed on the extension's
// discriminator.
//
// It reads the *registry*, so the binary that drives it decides which extensions
// exist: Kuma's own generator registers none and leaves the document untouched,
// while a downstream build importing its extension packages documents them.
package extensions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"sigs.k8s.io/yaml"

	core_extensions "github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
	"github.com/kumahq/kuma/v3/tools/openapi/unions"
)

// group is the API group of the throwaway CRDs. Nothing installs them; the group
// only has to be a valid DNS name so controller-gen accepts it.
const group = "openapi.kuma.io"

const version = "v1alpha1"

// Options configures Generate.
type Options struct {
	// Spec is the OpenAPI document to patch in place.
	Spec string
	// WorkDir is the parent of the scratch directory the generator writes a
	// throwaway Go package into. It has to be inside the module being generated,
	// because controller-gen reads Go source. Nothing in it is touched other than
	// the one directory the generator creates and removes.
	WorkDir string
	// ControllerGenBin is the path to a controller-gen binary.
	ControllerGenBin string
	// YqBin is the path to a yq binary.
	YqBin string
	// Stderr receives the output of the binaries the generator shells out to.
	Stderr io.Writer
}

// wrapper is one registered extension together with the names it is generated under.
type wrapper struct {
	ext core_extensions.Extension
	// kind is the CRD kind, which is also the name of the generated Go wrapper type.
	kind string
	// schemaName is the name the config schema takes in `components.schemas`.
	schemaName string
}

// Generate documents every registered extension in opts.Spec. It is a no-op when
// nothing is registered, which is what keeps upstream Kuma's spec free of a
// downstream product's extensions.
func Generate(ctx context.Context, opts Options) error {
	registered := core_extensions.Registered()
	if len(registered) == 0 {
		return nil
	}

	wrappers, err := newWrappers(registered)
	if err != nil {
		return err
	}

	spec, err := loadSpec(opts.Spec)
	if err != nil {
		return err
	}
	for _, w := range wrappers {
		if err := checkExtensionPoint(spec, w.ext.Point); err != nil {
			return err
		}
	}

	if opts.WorkDir == "" {
		return fmt.Errorf("work dir must not be empty")
	}
	if err := os.MkdirAll(opts.WorkDir, 0o755); err != nil {
		return err
	}
	// Generate its own directory rather than taking one on trust: the cleanup
	// below is then a directory this process created under a name nothing else
	// holds, so no work dir a caller passes can turn it into a destructive one.
	scratch, err := os.MkdirTemp(opts.WorkDir, "gen-")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	pkgDir := filepath.Join(scratch, version)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return err
	}
	if err := writePackage(pkgDir, wrappers); err != nil {
		return err
	}

	crdDir := filepath.Join(scratch, "crd")
	if err := runControllerGen(ctx, opts, pkgDir, crdDir); err != nil {
		return err
	}

	schemas, err := configSchemas(crdDir, wrappers)
	if err != nil {
		return err
	}

	expression, err := patchExpression(wrappers, schemas)
	if err != nil {
		return err
	}

	return runYq(ctx, opts, expression)
}

func newWrappers(registered []core_extensions.Extension) ([]wrapper, error) {
	wrappers := make([]wrapper, 0, len(registered))
	byKind := map[string]core_extensions.Extension{}
	for _, ext := range registered {
		kind := identifier(string(ext.Point.ResourceType) + upperFirst(ext.Value))
		if kind == "" {
			return nil, fmt.Errorf("extension %q on %s has no character usable in a schema name",
				ext.Value, ext.Point.ResourceType)
		}
		if clash, ok := byKind[kind]; ok {
			return nil, fmt.Errorf("extensions %q and %q on %s both generate the schema name %q",
				clash.Value, ext.Value, ext.Point.ResourceType, kind)
		}
		byKind[kind] = ext
		wrappers = append(wrappers, wrapper{ext: ext, kind: kind, schemaName: kind + "ExtensionConfig"})
	}
	return wrappers, nil
}

func loadSpec(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var spec map[string]any
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("could not parse %s: %w", path, err)
	}
	return spec, nil
}

// checkExtensionPoint fails when the point does not describe a node that is really
// in the document. yq creates whatever a path names, so without this a stale
// SchemaPath would grow a plausible-looking branch nobody asked for instead of
// reporting that the resource moved.
func checkExtensionPoint(spec map[string]any, point core_extensions.Point) error {
	item := string(point.ResourceType) + "Item"
	node, ok := dig(spec, append([]string{"components", "schemas", item}, propertyPath(point.SchemaPath)...))
	if !ok {
		return fmt.Errorf("%s does not have a %s.%s schema to document",
			point.ResourceType, item, strings.Join(point.SchemaPath, "."))
	}
	obj, ok := node.(map[string]any)
	if !ok {
		return fmt.Errorf("%s.%s is not an object schema", item, strings.Join(point.SchemaPath, "."))
	}
	properties, _ := obj["properties"].(map[string]any)
	if _, ok := properties[point.Discriminator]; !ok {
		return fmt.Errorf("%s.%s has no %q property to discriminate on",
			item, strings.Join(point.SchemaPath, "."), point.Discriminator)
	}
	return nil
}

func dig(node any, path []string) (any, bool) {
	for _, key := range path {
		obj, ok := node.(map[string]any)
		if !ok {
			return nil, false
		}
		if node, ok = obj[key]; !ok {
			return nil, false
		}
	}
	return node, true
}

// propertyPath interleaves `properties` into a schema path, so {"spec", "extension"}
// addresses `properties.spec.properties.extension`.
func propertyPath(schemaPath []string) []string {
	path := make([]string, 0, len(schemaPath)*2)
	for _, segment := range schemaPath {
		path = append(path, "properties", segment)
	}
	return path
}

func writePackage(pkgDir string, wrappers []wrapper) error {
	doc := fmt.Sprintf(`// Code generated by tools/openapi/extensions. DO NOT EDIT.

// +groupName=%s
// +versionName=%s
package %s
`, group, version, version)
	if err := os.WriteFile(filepath.Join(pkgDir, "doc.go"), []byte(doc), 0o600); err != nil {
		return err
	}

	// controller-gen only walks types marked as a CRD root, so every config is
	// generated as the spec of a throwaway resource.
	aliases := map[string]string{}
	var imports []string
	for _, w := range wrappers {
		pkgPath := w.ext.ConfigType().PkgPath()
		if _, ok := aliases[pkgPath]; ok {
			continue
		}
		alias := fmt.Sprintf("ext%d", len(aliases))
		aliases[pkgPath] = alias
		imports = append(imports, fmt.Sprintf("\t%s %q\n", alias, pkgPath))
	}
	sort.Strings(imports)

	var sb strings.Builder
	sb.WriteString("// Code generated by tools/openapi/extensions. DO NOT EDIT.\n\n")
	fmt.Fprintf(&sb, "package %s\n\n", version)
	sb.WriteString("import (\n\tmetav1 \"k8s.io/apimachinery/pkg/apis/meta/v1\"\n\n")
	for _, imp := range imports {
		sb.WriteString(imp)
	}
	sb.WriteString(")\n")
	for _, w := range wrappers {
		configType := w.ext.ConfigType()
		fmt.Fprintf(&sb, `
// %s wraps the %q extension of %s.
// +kubebuilder:object:root=true
type %s struct {
	metav1.TypeMeta   `+"`json:\",inline\"`"+`
	metav1.ObjectMeta `+"`json:\"metadata,omitempty\"`"+`

	Spec *%s.%s `+"`json:\"spec,omitempty\"`"+`
}
`, w.kind, w.ext.Value, w.ext.Point.ResourceType, w.kind, aliases[configType.PkgPath()], configType.Name())
	}

	return os.WriteFile(filepath.Join(pkgDir, "types.go"), []byte(sb.String()), 0o600)
}

func runControllerGen(ctx context.Context, opts Options, pkgDir, crdDir string) error {
	// go/packages reads a directory pattern only when it is explicitly relative or
	// absolute, and WorkDir may be either.
	pattern := filepath.ToSlash(pkgDir)
	if !filepath.IsAbs(pattern) {
		pattern = "./" + pattern
	}
	cmd := exec.CommandContext(ctx, opts.ControllerGenBin, //nolint:gosec // the binary is a build tool path passed by the Makefile
		"crd:crdVersions=v1,ignoreUnexportedFields=true",
		"paths="+pattern+"/...",
		"output:crd:artifacts:config="+crdDir,
	)
	cmd.Stderr = opts.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("controller-gen failed on the generated extension wrappers: %w", err)
	}
	return nil
}

// configSchemas returns the schema of each extension config, keyed by CRD kind.
// The CRDs are indexed by the kind they declare rather than by filename, which
// controller-gen derives through pluralization we would rather not reimplement.
func configSchemas(crdDir string, wrappers []wrapper) (map[string]any, error) {
	entries, err := os.ReadDir(crdDir)
	if err != nil {
		return nil, err
	}

	byKind := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join(crdDir, entry.Name())
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var crd struct {
			Spec struct {
				Names struct {
					Kind string `json:"kind"`
				} `json:"names"`
			} `json:"spec"`
		}
		if err := yaml.Unmarshal(raw, &crd); err != nil {
			return nil, err
		}
		byKind[crd.Spec.Names.Kind] = path
	}

	schemas := map[string]any{}
	for _, w := range wrappers {
		path, ok := byKind[w.kind]
		if !ok {
			return nil, fmt.Errorf("controller-gen produced no schema for the %q extension of %s",
				w.ext.Value, w.ext.Point.ResourceType)
		}
		properties, err := unions.CRDProperties(path)
		if err != nil {
			return nil, err
		}
		schema, ok := properties["spec"]
		if !ok {
			return nil, fmt.Errorf("the schema for the %q extension of %s has no spec",
				w.ext.Value, w.ext.Point.ResourceType)
		}
		schemas[w.kind] = schema
	}
	return schemas, nil
}

// patchExpression renders the yq program that adds every config schema to
// `components.schemas` and points the extension nodes at them.
func patchExpression(wrappers []wrapper, schemas map[string]any) (string, error) {
	var assignments []string

	for _, w := range wrappers {
		encoded, err := json.Marshal(schemas[w.kind])
		if err != nil {
			return "", err
		}
		schemaPath := []string{"components", "schemas", w.schemaName}
		assignments = append(assignments, fmt.Sprintf("%s = %s", unions.YQPath(schemaPath), encoded))

		// The config is a resource spec like any other, so it can hold unions of
		// its own that controller-gen cannot express.
		unionAssignments, err := unions.Assignments(schemas[w.kind], schemaPath)
		if err != nil {
			return "", err
		}
		if unionAssignments != "" {
			assignments = append(assignments, unionAssignments)
		}
	}

	for _, point := range points(wrappers) {
		oneOf, err := json.Marshal(branches(point, wrappers))
		if err != nil {
			return "", err
		}
		path := append([]string{"components", "schemas", string(point.ResourceType) + "Item"},
			propertyPath(point.SchemaPath)...)
		assignments = append(assignments, fmt.Sprintf("%s.oneOf = %s", unions.YQPath(path), oneOf))
	}

	return strings.Join(assignments, "\n  | ") + "\n", nil
}

// points lists the distinct extension points in registration order.
func points(wrappers []wrapper) []core_extensions.Point {
	var ordered []core_extensions.Point
	seen := map[string]bool{}
	for _, w := range wrappers {
		key := string(w.ext.Point.ResourceType)
		if seen[key] {
			continue
		}
		seen[key] = true
		ordered = append(ordered, w.ext.Point)
	}
	return ordered
}

// branches describes one extension point: a branch per known extension, plus a
// catch-all so that an extension this build does not ship still validates. Without
// it the spec would reject configurations that the control plane accepts.
func branches(point core_extensions.Point, wrappers []wrapper) []any {
	var known []any
	var values []any
	for _, w := range wrappers {
		if w.ext.Point.ResourceType != point.ResourceType {
			continue
		}
		values = append(values, w.ext.Value)
		known = append(known, map[string]any{
			"title": w.ext.Value,
			"properties": map[string]any{
				point.Discriminator: map[string]any{"const": w.ext.Value},
				"config":            map[string]any{"$ref": "#/components/schemas/" + w.schemaName},
			},
		})
	}
	return append(known, map[string]any{
		"title":       "Other",
		"description": "An extension this control plane does not ship. Its configuration is not described here.",
		"properties": map[string]any{
			point.Discriminator: map[string]any{"not": map[string]any{"enum": values}},
		},
	})
}

func runYq(ctx context.Context, opts Options, expression string) error {
	file, err := os.CreateTemp("", "openapi-extensions-*.yq")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(file.Name()) }()
	if _, err := file.WriteString(expression); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}

	// --from-file, not -f: -f is --front-matter, which quietly reformats the
	// document instead of running the expression.
	cmd := exec.CommandContext(ctx, opts.YqBin, "e", "-i", "--from-file", file.Name(), opts.Spec) //nolint:gosec // the binary is a build tool path passed by the Makefile
	cmd.Stderr = opts.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not patch %s: %w", opts.Spec, err)
	}
	return nil
}

// identifier keeps the characters that are legal in a Go type name and a CRD kind.
func identifier(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || (unicode.IsDigit(r) && sb.Len() > 0) {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
