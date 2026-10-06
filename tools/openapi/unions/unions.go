// Package unions finds the discriminated unions in a controller-gen schema and
// renders them as yq assignments.
//
// Kuma models a union as a `type` discriminator plus one optional property per
// variant. controller-gen emits those variants as unrelated siblings, so nothing
// in the spec says which property a given `type` selects and consumers have to
// hardcode the mapping. Describing them with a oneOf of closed variants removes
// the guesswork.
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
	// Branches pairs each discriminator value with the property it selects, in
	// the order of the discriminator's enum.
	Branches []Branch
	// Shared lists the sibling properties that belong to no variant, which every
	// branch keeps.
	Shared []string
	// Required lists the node's required properties, which every branch keeps.
	Required []string
	// Description is the description of the `type` property, if any.
	Description string
}

// Branch is one variant of a discriminated union.
type Branch struct {
	// Value is the discriminator value selecting the variant.
	Value string
	// Property is the sibling property holding the variant's configuration.
	Property string
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
	if site, ok := union(obj); ok {
		site.Path = append([]string{}, base...)
		sites = append(sites, site)
	}
	for key, child := range obj {
		sites = append(sites, Find(child, append(base, key))...)
	}
	return sites
}

func union(obj map[string]any) (Site, bool) {
	properties, ok := obj["properties"].(map[string]any)
	if !ok {
		return Site{}, false
	}
	discriminator, ok := properties["type"].(map[string]any)
	if !ok {
		return Site{}, false
	}
	values, ok := discriminator["enum"].([]any)
	if !ok || len(values) < 2 {
		return Site{}, false
	}

	site := Site{Branches: make([]Branch, 0, len(values))}
	variants := map[string]bool{"type": true}
	for _, value := range values {
		name, ok := value.(string)
		if !ok {
			return Site{}, false
		}
		variant, ok := variantProperty(properties, name)
		if !ok {
			return Site{}, false
		}
		site.Branches = append(site.Branches, Branch{Value: name, Property: variant})
		variants[variant] = true
	}
	for key := range properties {
		if !variants[key] {
			site.Shared = append(site.Shared, key)
		}
	}
	sort.Strings(site.Shared)
	if required, ok := obj["required"].([]any); ok {
		for _, r := range required {
			if name, ok := r.(string); ok && name != "type" {
				site.Required = append(site.Required, name)
			}
		}
	}
	site.Description, _ = discriminator["description"].(string)
	return site, true
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

// Assignments renders a yq expression replacing every discriminated union in the
// schema with a oneOf, with each path prefixed by base.
//
// Each branch is a closed variant: `type` pinned with `const` plus only the
// property that value selects (and any properties shared by all variants).
// Generators that name inline oneOf members by title (Speakeasy) get a name per
// branch, and the const lets them tell the branches apart when decoding. `type`
// also carries a single-value `enum`, since the control plane validates against
// rest.yaml with kube-openapi, which ignores `const`. A variant that carries no
// configuration can still be written as just `type: <value>`, since only `type`
// is required.
//
// The union node keeps its `properties`, which the control plane needs for
// defaulting and pruning. Generators merge them into every branch, so the
// published document drops them (see `docs/generated/openapi.yaml` in
// mk/docs.mk).
//
// The unions are found in the parsed schema rather than in the file being edited
// so the assignments can be appended to the yq call that writes it: that keeps the
// generated file's key order intact, which marshaling the whole document through
// Go would not. The variant schemas are copied by path for the same reason, so
// the deepest unions are rewritten first and an enclosing union copies the
// rewritten form.
func Assignments(schema any, base []string) (string, error) {
	sites := Find(schema, base)
	if len(sites) == 0 {
		return "", nil
	}
	sort.Slice(sites, func(i, j int) bool {
		if len(sites[i].Path) != len(sites[j].Path) {
			return len(sites[i].Path) > len(sites[j].Path)
		}
		return strings.Join(sites[i].Path, ".") < strings.Join(sites[j].Path, ".")
	})

	assignments := make([]string, 0, len(sites))
	for _, site := range sites {
		oneOf, err := site.oneOf()
		if err != nil {
			return "", err
		}
		assignments = append(assignments, fmt.Sprintf("%s.oneOf = %s", YQPath(site.Path), oneOf))
	}
	return strings.Join(assignments, "\n  | "), nil
}

// oneOf renders the site's branches as a yq array expression.
func (s Site) oneOf() (string, error) {
	required, err := json.Marshal(append([]string{"type"}, s.Required...))
	if err != nil {
		return "", err
	}
	properties := YQPath(append(append([]string{}, s.Path...), "properties"))
	branches := make([]string, 0, len(s.Branches))
	for _, b := range s.Branches {
		value, err := json.Marshal(b.Value)
		if err != nil {
			return "", err
		}
		discriminator := fmt.Sprintf(`{"type": "string", "enum": [%s], "const": %s}`, value, value)
		if s.Description != "" {
			description, err := json.Marshal(s.Description)
			if err != nil {
				return "", err
			}
			discriminator = fmt.Sprintf(`{"description": %s, "type": "string", "enum": [%s], "const": %s}`, description, value, value)
		}
		fields := []string{fmt.Sprintf(`"type": %s`, discriminator)}
		for _, name := range append([]string{b.Property}, s.Shared...) {
			key, err := json.Marshal(name)
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("%s: %s%s", key, properties, YQPath([]string{name})))
		}
		title, err := json.Marshal(upperFirst(s.field()) + b.Value)
		if err != nil {
			return "", err
		}
		branches = append(branches, fmt.Sprintf(`{"title": %s, "type": "object", "required": %s, "properties": {%s}}`,
			title, required, strings.Join(fields, ", ")))
	}
	return "[" + strings.Join(branches, ", ") + "]", nil
}

// field names the property holding the union, which prefixes the branch titles.
// Generators that name a oneOf member by its title (Speakeasy) nest the variant
// under it, so a bare `LeastRequest` title gives
// `least_request = { least_request = {...} }`, while `LoadBalancerLeastRequest`
// gives `load_balancer_least_request = { least_request = {...} }`. It is the
// last key following a `properties` map in the path, so array items resolve to
// the array, e.g. `filters` for `...properties.filters.items`. A union that is a
// schema of its own falls back to the schema name.
func (s Site) field() string {
	var field string
	for i := 0; i+1 < len(s.Path); i++ {
		if s.Path[i] == "properties" {
			i++
			field = s.Path[i]
		}
	}
	if field == "" && len(s.Path) > 0 {
		field = s.Path[len(s.Path)-1]
	}
	return lowerFirst(field)
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// YQPath renders map keys as a yq path expression.
func YQPath(path []string) string {
	var sb strings.Builder
	for _, key := range path {
		fmt.Fprintf(&sb, ".%q", key)
	}
	return sb.String()
}
