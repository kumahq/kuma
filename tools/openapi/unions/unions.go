// Package unions finds the discriminated unions in a controller-gen schema and
// renders them as yq assignments.
//
// Kuma models a union as a `type` discriminator plus one optional property per
// variant. controller-gen emits those variants as unrelated siblings, so nothing
// in the spec says which property a given `type` selects and consumers have to
// hardcode the mapping. Describing them as an OpenAPI discriminated union, a oneOf
// of named member schemas plus a `discriminator` mapping each value to its
// member, removes the guesswork and gives code generators a name for every
// variant instead of a position.
package unions

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"sort"
	"strings"
	"unicode"

	"sigs.k8s.io/yaml"
)

// discriminatorProperty is the property every Kuma union is discriminated on.
const discriminatorProperty = "type"

const schemaRefPrefix = "#/components/schemas/"

// Site is a schema node that models a discriminated union: an object with a
// `type` enum where every value has a matching sibling property holding that
// variant's configuration.
type Site struct {
	// Path is the sequence of map keys leading to the node, for a yq assignment.
	Path []string
	// Values are the discriminator values, in the order of the enum.
	Values []string
	// Variants maps each discriminator value to the property it selects.
	Variants map[string]string
}

// Find walks a schema and reports every discriminated union, prefixing each
// reported path with base.
//
// A node qualifies only when *every* enum value resolves to a sibling property,
// which is a tight enough fingerprint to avoid dragging in plain enums that happen
// to sit next to similarly named fields.
func Find(node any, base []string) []Site {
	obj, ok := node.(map[string]any)
	if !ok {
		return nil
	}

	var sites []Site
	if values, variants, ok := union(obj); ok {
		sites = append(sites, Site{Path: append([]string{}, base...), Values: values, Variants: variants})
	}
	for key, child := range obj {
		sites = append(sites, Find(child, append(base, key))...)
	}
	return sites
}

func union(obj map[string]any) ([]string, map[string]string, bool) {
	properties, ok := obj["properties"].(map[string]any)
	if !ok {
		return nil, nil, false
	}
	discriminator, ok := properties[discriminatorProperty].(map[string]any)
	if !ok {
		return nil, nil, false
	}
	enum, ok := discriminator["enum"].([]any)
	if !ok || len(enum) < 2 {
		return nil, nil, false
	}

	values := make([]string, 0, len(enum))
	variants := map[string]string{}
	for _, value := range enum {
		name, ok := value.(string)
		if !ok {
			return nil, nil, false
		}
		variant, ok := variantProperty(properties, name)
		if !ok {
			return nil, nil, false
		}
		values = append(values, name)
		variants[name] = variant
	}
	return values, variants, true
}

// variantProperty resolves a discriminator value to the sibling property holding
// that variant, e.g. `RoundRobin` to `roundRobin` and `URLRewrite` to `urlRewrite`.
func variantProperty(properties map[string]any, value string) (string, bool) {
	for _, candidate := range []string{lowerFirst(value), lowerAcronym(value)} {
		if candidate == "" || candidate == discriminatorProperty {
			continue
		}
		if _, ok := properties[candidate]; ok {
			return candidate, true
		}
	}
	return "", false
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// lowerAcronym lowercases a leading run of capitals, so `URLRewrite` becomes
// `urlRewrite` the way encoding/json names it.
func lowerAcronym(s string) string {
	r := []rune(s)
	i := 0
	for i < len(r) && unicode.IsUpper(r[i]) {
		i++
	}
	if i <= 1 {
		return lowerFirst(s)
	}
	if i < len(r) {
		i-- // the last capital starts the next word
	}
	return strings.ToLower(string(r[:i])) + string(r[i:])
}

// CRDProperties reads the top-level `properties` of a controller-gen CRD.
func CRDProperties(crdPath string) (map[string]any, error) {
	raw, err := os.ReadFile(crdPath)
	if err != nil {
		return nil, err
	}
	var crd struct {
		Spec struct {
			Versions []struct {
				Schema struct {
					OpenAPIV3Schema struct {
						Properties map[string]any `json:"properties"`
					} `json:"openAPIV3Schema"`
				} `json:"schema"`
			} `json:"versions"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(raw, &crd); err != nil {
		return nil, err
	}
	if len(crd.Spec.Versions) == 0 {
		return nil, nil
	}
	return crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties, nil
}

// Member is the named schema of one variant of a union. It allows a single
// discriminator value and holds the property that value selects, together with
// whatever the union's variants have in common. Only the discriminator is
// required, so a variant that carries no configuration can still be written as
// just `type: <value>`.
type Member struct {
	Title      string         `json:"title"`
	Type       string         `json:"type"`
	Required   []any          `json:"required"`
	Properties map[string]any `json:"properties"`
}

// Rewrite is what Assignments applies to a document.
type Rewrite struct {
	// Members are the member schemas to add to `components.schemas`, by name.
	Members map[string]Member
	// Unions are the union nodes to point at their members, by path. Unions
	// nested inside another union end up in that union's members, so only the
	// outermost ones are listed.
	Unions map[string]Union
}

// Union is the oneOf and discriminator replacing a union node's properties.
type Union struct {
	Path          []string
	OneOf         []any
	Discriminator map[string]any
}

// Plan works out the member schemas of every discriminated union in the schema
// and how each union node refers to them. The schema itself is not modified.
//
// Members are named prefix + the name of the union property + the discriminator
// value, e.g. `MeshLoadBalancingStrategy` + `LoadBalancer` + `RingHash`, with the
// property of an array item singularized (`backends` gives `Backend`). Two unions
// with the same property name share their members when the members are equal,
// which is the common case of one Go type used in two places. When they are not,
// the names of the enclosing properties are prepended until they differ.
func Plan(schema any, base []string, prefix string) (Rewrite, error) {
	sites := Find(schema, base)
	// Deepest first, so a union nested in another one is rewritten before the
	// outer union's members copy it.
	sort.Slice(sites, func(i, j int) bool {
		if len(sites[i].Path) != len(sites[j].Path) {
			return len(sites[i].Path) > len(sites[j].Path)
		}
		return strings.Join(sites[i].Path, ".") < strings.Join(sites[j].Path, ".")
	})

	depths := make([]int, len(sites))
	for i := range depths {
		depths[i] = 1
	}
	for {
		rewrite, conflicts, err := plan(schema, base, prefix, sites, depths)
		if err != nil {
			return Rewrite{}, err
		}
		if len(conflicts) == 0 {
			return rewrite, nil
		}
		for i := range conflicts {
			depths[i]++
		}
	}
}

// plan renders the rewrite with the given name depths, and reports the sites
// whose member names clash with a different member.
func plan(schema any, base []string, prefix string, sites []Site, depths []int) (Rewrite, map[int]bool, error) {
	// The schema is rewritten as it goes, so that the members of an outer union
	// hold the rewritten form of the unions nested in them.
	var doc any
	encoded, err := json.Marshal(schema)
	if err != nil {
		return Rewrite{}, nil, err
	}
	if err := json.Unmarshal(encoded, &doc); err != nil {
		return Rewrite{}, nil, err
	}

	rewrite := Rewrite{Members: map[string]Member{}, Unions: map[string]Union{}}
	owners := map[string][]int{}
	contents := map[string]string{}
	conflicts := map[int]bool{}
	for i, site := range sites {
		node, ok := dig(doc, site.Path[len(base):]).(map[string]any)
		if !ok {
			return Rewrite{}, nil, fmt.Errorf("union at %s is not an object", strings.Join(site.Path, "."))
		}
		names := propertyNames(site.Path)
		if depths[i] > len(names) {
			return Rewrite{}, nil, fmt.Errorf("could not give the members of the union at %s a unique name", strings.Join(site.Path, "."))
		}
		qualifier := prefix + strings.Join(names[len(names)-depths[i]:], "")

		oneOf := make([]any, 0, len(site.Values))
		mapping := map[string]any{}
		for _, value := range site.Values {
			name := qualifier + value
			member := newMember(node, site, value)
			content, err := json.Marshal(member)
			if err != nil {
				return Rewrite{}, nil, err
			}
			if existing, ok := contents[name]; ok && existing != string(content) {
				conflicts[i] = true
				for _, owner := range owners[name] {
					conflicts[owner] = true
				}
			}
			contents[name] = string(content)
			owners[name] = append(owners[name], i)
			rewrite.Members[name] = member

			ref := schemaRefPrefix + name
			oneOf = append(oneOf, map[string]any{"$ref": ref})
			mapping[value] = ref
		}
		discriminator := map[string]any{"propertyName": discriminatorProperty, "mapping": mapping}

		delete(node, "properties")
		delete(node, "required")
		delete(node, "type")
		node["oneOf"] = oneOf
		node["discriminator"] = discriminator

		rewrite.Unions[strings.Join(site.Path, ".")] = Union{Path: site.Path, OneOf: oneOf, Discriminator: discriminator}
	}

	// A union inside another one is gone from the document once the outer one
	// drops its properties, and lives on in the outer union's members.
	for key, u := range rewrite.Unions {
		for _, outer := range rewrite.Unions {
			if len(outer.Path) < len(u.Path) && isPrefix(outer.Path, u.Path) {
				delete(rewrite.Unions, key)
				break
			}
		}
	}
	return rewrite, conflicts, nil
}

func newMember(node map[string]any, site Site, value string) Member {
	properties, _ := node["properties"].(map[string]any)
	variants := map[string]bool{}
	for _, variant := range site.Variants {
		variants[variant] = true
	}

	memberProperties := map[string]any{}
	for name, property := range properties {
		switch {
		case name == discriminatorProperty:
			discriminator := map[string]any{}
			maps.Copy(discriminator, property.(map[string]any))
			discriminator["enum"] = []any{value}
			memberProperties[name] = discriminator
		case variants[name] && name != site.Variants[value]:
			// another variant's configuration
		default:
			memberProperties[name] = property
		}
	}

	required, _ := node["required"].([]any)
	required = append([]any{}, required...)
	hasDiscriminator := false
	for _, r := range required {
		if r == discriminatorProperty {
			hasDiscriminator = true
		}
	}
	// OpenAPI requires the discriminator of a member to be required.
	if !hasDiscriminator {
		required = append(required, discriminatorProperty)
	}

	return Member{Title: value, Type: "object", Required: required, Properties: memberProperties}
}

// propertyNames lists the names of the properties along a schema path, upper
// cased to be joined into a schema name. The property of an array is named after
// its items, so `backends` gives `Backend`.
func propertyNames(path []string) []string {
	var names []string
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case "properties":
			if i+1 < len(path) {
				names = append(names, upperFirst(path[i+1]))
				i++
			}
		case "items":
			if len(names) > 0 {
				names[len(names)-1] = singular(names[len(names)-1])
			}
		}
	}
	if len(names) == 0 {
		return []string{""}
	}
	return names
}

func singular(s string) string {
	switch {
	case strings.HasSuffix(s, "ies"):
		return strings.TrimSuffix(s, "ies") + "y"
	case strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss"):
		return strings.TrimSuffix(s, "s")
	}
	return s
}

func dig(node any, path []string) any {
	for _, key := range path {
		obj, ok := node.(map[string]any)
		if !ok {
			return nil
		}
		node = obj[key]
	}
	return node
}

func isPrefix(prefix, path []string) bool {
	if len(prefix) > len(path) {
		return false
	}
	for i := range prefix {
		if prefix[i] != path[i] {
			return false
		}
	}
	return true
}

// Assignments renders a yq expression describing every discriminated union in the
// schema, with each path prefixed by base: the member schemas are added to
// `components.schemas` of the document, under names starting with prefix, and
// each union node has its properties replaced by a oneOf of its members and a
// discriminator mapping each value to its member.
//
// The unions are found in the parsed schema rather than in the file being edited
// so the assignments can be appended to the yq call that writes it: that keeps the
// generated file's key order intact, which marshaling the whole document through
// Go would not.
func Assignments(schema any, base []string, prefix string) (string, error) {
	rewrite, err := Plan(schema, base, prefix)
	if err != nil {
		return "", err
	}

	names := make([]string, 0, len(rewrite.Members))
	for name := range rewrite.Members {
		names = append(names, name)
	}
	sort.Strings(names)
	keys := make([]string, 0, len(rewrite.Unions))
	for key := range rewrite.Unions {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	assignments := make([]string, 0, len(names)+len(keys))
	for _, name := range names {
		encoded, err := json.Marshal(rewrite.Members[name])
		if err != nil {
			return "", err
		}
		assignments = append(assignments, fmt.Sprintf("%s = %s", YQPath([]string{"components", "schemas", name}), encoded))
	}
	for _, key := range keys {
		u := rewrite.Unions[key]
		oneOf, err := json.Marshal(u.OneOf)
		if err != nil {
			return "", err
		}
		discriminator, err := json.Marshal(u.Discriminator)
		if err != nil {
			return "", err
		}
		assignments = append(assignments, fmt.Sprintf("%s |= (del(.properties, .required, .type) | .oneOf = %s | .discriminator = %s)",
			YQPath(u.Path), oneOf, discriminator))
	}
	return strings.Join(assignments, "\n  | "), nil
}

// YQPath renders map keys as a yq path expression.
func YQPath(path []string) string {
	var sb strings.Builder
	for _, key := range path {
		fmt.Fprintf(&sb, ".%q", key)
	}
	return sb.String()
}
