package model

import (
	"fmt"
	"maps"
	"strings"
)

const schemaRefPrefix = "#/components/schemas/"

// FlattenDiscriminatedUnions rewrites, in place, every discriminated union in an
// OpenAPI schema back into the single object Kubernetes structural schemas need.
//
// The published spec describes a union as a oneOf of named member schemas in
// `components.schemas`, each carrying the discriminator with one allowed value.
// A structural schema rejects `$ref` and prunes anything its `properties` do not
// list, so each union node is turned back into one object holding the properties
// of all of its members, with the discriminator allowing every value. That is the
// schema controller-gen produced before the union was described.
func FlattenDiscriminatedUnions(node any, schemas map[string]any) error {
	switch n := node.(type) {
	case map[string]any:
		if _, ok := n["discriminator"]; ok {
			if err := flattenUnion(n, schemas); err != nil {
				return err
			}
		}
		for _, child := range n {
			if err := FlattenDiscriminatedUnions(child, schemas); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range n {
			if err := FlattenDiscriminatedUnions(child, schemas); err != nil {
				return err
			}
		}
	}
	return nil
}

func flattenUnion(node map[string]any, schemas map[string]any) error {
	discriminator, _ := node["discriminator"].(map[string]any)
	propertyName, _ := discriminator["propertyName"].(string)
	oneOf, _ := node["oneOf"].([]any)
	if propertyName == "" || len(oneOf) == 0 {
		return fmt.Errorf("discriminated union must have a propertyName and a oneOf")
	}

	properties := map[string]any{}
	var values []any
	var required any
	var discriminatorProperty map[string]any
	for _, entry := range oneOf {
		member, err := resolveMember(entry, schemas)
		if err != nil {
			return err
		}
		memberProperties, _ := member["properties"].(map[string]any)
		for name, property := range memberProperties {
			if name != propertyName {
				properties[name] = property
				continue
			}
			property, ok := property.(map[string]any)
			if !ok {
				return fmt.Errorf("discriminator %q of a union member is not a schema", propertyName)
			}
			enum, _ := property["enum"].([]any)
			values = append(values, enum...)
			if discriminatorProperty == nil {
				discriminatorProperty = map[string]any{}
				maps.Copy(discriminatorProperty, property)
			}
		}
		if required == nil {
			required = member["required"]
		}
	}
	if discriminatorProperty == nil {
		return fmt.Errorf("no member of the union has the discriminator %q", propertyName)
	}
	discriminatorProperty["enum"] = values
	properties[propertyName] = discriminatorProperty

	delete(node, "discriminator")
	delete(node, "oneOf")
	node["type"] = "object"
	node["properties"] = properties
	if required != nil {
		node["required"] = required
	}
	return nil
}

func resolveMember(entry any, schemas map[string]any) (map[string]any, error) {
	ref, _ := entry.(map[string]any)["$ref"].(string)
	name, ok := strings.CutPrefix(ref, schemaRefPrefix)
	if !ok {
		return nil, fmt.Errorf("union member %v is not a reference to %s", entry, schemaRefPrefix)
	}
	member, ok := schemas[name].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("union member %q not found in components.schemas", name)
	}
	return member, nil
}
