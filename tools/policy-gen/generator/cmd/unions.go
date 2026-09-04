package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"sigs.k8s.io/yaml"
)

// unionSite is a schema node that models a discriminated union: an object with a
// `type` enum where every value has a matching sibling property holding that
// variant's configuration.
type unionSite struct {
	// path is the sequence of map keys leading to the node, for a yq assignment.
	path []string
	// oneOf pairs each discriminator value with the property it selects.
	oneOf []any
}

// findUnionSites walks a generated schema and reports every discriminated union.
//
// Kuma models unions as a `type` discriminator plus one optional property per
// variant. controller-gen emits those variants as unrelated siblings, so nothing
// in the spec says which property a given `type` selects and consumers have to
// hardcode the mapping. A node qualifies only when *every* enum value resolves to
// a sibling property, which is a tight enough fingerprint to avoid dragging in
// plain enums that happen to sit next to similarly named fields.
func findUnionSites(node any, path []string) []unionSite {
	obj, ok := node.(map[string]any)
	if !ok {
		return nil
	}

	var sites []unionSite
	if oneOf, ok := unionOneOf(obj); ok {
		sites = append(sites, unionSite{path: append([]string{}, path...), oneOf: oneOf})
	}
	for key, child := range obj {
		sites = append(sites, findUnionSites(child, append(path, key))...)
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

// unionAssignments reads the CRD and renders a yq expression adding a oneOf to
// every discriminated union it declares.
//
// The unions are found in the CRD rather than in the enriched schema so the
// assignments can be appended to the yq call that does the enrichment: that keeps
// the generated file's key order intact, which marshaling the whole document
// through Go would not.
func unionAssignments(crdPath string) (string, error) {
	raw, err := os.ReadFile(crdPath)
	if err != nil {
		return "", err
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
		return "", err
	}
	if len(crd.Spec.Versions) == 0 {
		return "", nil
	}

	// The enrichment merges the CRD properties into `.properties`, so a union at
	// `spec.foo` in the CRD lands at `.properties.spec.foo` in the schema.
	sites := findUnionSites(crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties, []string{"properties"})
	if len(sites) == 0 {
		return "", nil
	}
	sort.Slice(sites, func(i, j int) bool {
		return strings.Join(sites[i].path, ".") < strings.Join(sites[j].path, ".")
	})

	assignments := make([]string, 0, len(sites))
	for _, site := range sites {
		encoded, err := json.Marshal(site.oneOf)
		if err != nil {
			return "", err
		}
		assignments = append(assignments, fmt.Sprintf("%s.oneOf = %s", yqPath(site.path), encoded))
	}
	return strings.Join(assignments, "\n  | "), nil
}

func yqPath(path []string) string {
	var sb strings.Builder
	for _, key := range path {
		fmt.Fprintf(&sb, ".%q", key)
	}
	return sb.String()
}
