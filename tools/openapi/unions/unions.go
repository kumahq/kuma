// Package unions finds the discriminated unions in a controller-gen schema and
// renders them as yq assignments.
//
// Kuma models a union as a `type` discriminator plus one optional property per
// variant. controller-gen emits those variants as unrelated siblings, so nothing
// in the spec says which property a given `type` selects and consumers have to
// hardcode the mapping. Describing them with a oneOf removes the guesswork.
package unions

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"sigs.k8s.io/yaml"
)

// Site is a schema node that models a discriminated union: an object with a
// `type` enum where every value has a matching sibling property holding that
// variant's configuration.
type Site struct {
	// Path is the sequence of map keys leading to the node, for a yq assignment.
	Path []string
	// OneOf pairs each discriminator value with the property it selects.
	OneOf []any
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
	if oneOf, ok := unionOneOf(obj); ok {
		sites = append(sites, Site{Path: append([]string{}, base...), OneOf: oneOf})
	}
	for key, child := range obj {
		sites = append(sites, Find(child, append(base, key))...)
	}
	return sites
}

func unionOneOf(obj map[string]any) ([]any, bool) {
	properties, ok := obj["properties"].(map[string]any)
	if !ok {
		return nil, false
	}
	discriminator, ok := properties["type"].(map[string]any)
	if !ok {
		return nil, false
	}
	values, ok := discriminator["enum"].([]any)
	if !ok || len(values) < 2 {
		return nil, false
	}

	oneOf := make([]any, 0, len(values))
	for _, value := range values {
		name, ok := value.(string)
		if !ok {
			return nil, false
		}
		variant, ok := variantProperty(properties, name)
		if !ok {
			return nil, false
		}
		// The variant is matched as an unconstrained schema: the branch records
		// which property the value selects without making it required, so a
		// variant that carries no configuration can still be written as just
		// `type: <value>`.
		oneOf = append(oneOf, map[string]any{
			"properties": map[string]any{
				"type":  map[string]any{"enum": []any{name}},
				variant: map[string]any{},
			},
		})
	}
	return oneOf, true
}

// variantProperty resolves a discriminator value to the sibling property holding
// that variant, e.g. `RoundRobin` to `roundRobin` and `URLRewrite` to `urlRewrite`.
func variantProperty(properties map[string]any, value string) (string, bool) {
	for _, candidate := range []string{lowerFirst(value), lowerAcronym(value)} {
		if candidate == "" || candidate == "type" {
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

// Assignments renders a yq expression adding a oneOf to every discriminated union
// in the schema, with each path prefixed by base.
//
// The unions are found in the parsed schema rather than in the file being edited
// so the assignments can be appended to the yq call that writes it: that keeps the
// generated file's key order intact, which marshaling the whole document through
// Go would not.
func Assignments(schema any, base []string) (string, error) {
	sites := Find(schema, base)
	if len(sites) == 0 {
		return "", nil
	}
	sort.Slice(sites, func(i, j int) bool {
		return strings.Join(sites[i].Path, ".") < strings.Join(sites[j].Path, ".")
	})

	assignments := make([]string, 0, len(sites))
	for _, site := range sites {
		encoded, err := json.Marshal(site.OneOf)
		if err != nil {
			return "", err
		}
		assignments = append(assignments, fmt.Sprintf("%s.oneOf = %s", YQPath(site.Path), encoded))
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
