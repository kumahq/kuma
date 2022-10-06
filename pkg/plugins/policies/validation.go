package policies

import (
	"embed"
	"path"

	"github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

//go:embed */api/v1alpha1/schema.yaml
var content embed.FS

var resourceToSchema map[model.ResourceType]*gojsonschema.JSONLoader

func ValidateSchema(document string, type_ string) error {
	schema := resourceToSchema[model.ResourceType(type_)]
	if schema == nil {
		// we only validate new policies so for old or unknown ones we return no errors
		return nil
	}

	documentLoader := gojsonschema.NewStringLoader(document)
	result, err := gojsonschema.Validate(*schema, documentLoader)

	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	} else {
		return mapSchemaToValidatorErrors(result.Errors())
	}
}

func mapSchemaToValidatorErrors(errors []gojsonschema.ResultError) error {
	var verr validators.ValidationError

	for _, err := range errors {
		verr.AddViolation(err.Field(), err.Description())
	}

	return &verr
}

func init() {
	resourceToSchema = map[model.ResourceType]*gojsonschema.JSONLoader{}
	dirs, err := content.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, dir := range dirs {
		rawSchema, err := content.ReadFile(path.Join(".", dir.Name(), "api", "v1alpha1", "schema.yaml"))
		if err != nil {
			panic(err)
		}
		unmarshalled, err := unmarshal(rawSchema)
		if err != nil {
			panic(err)
		}

		name, err := getName(unmarshalled)
		if err != nil {
			panic(err)
		}

		schema, err := yamlToJsonSchemaLoader(rawSchema)
		if err != nil {
			panic(err)
		}

		resourceToSchema[model.ResourceType(name)] = schema
	}
}

func getName(unmarshalled map[string]interface{}) (string, error) {
	rawProperties := unmarshalled["properties"]
	properties := rawProperties.(map[string]interface{})
	rawType := properties["type"]
	type_ := rawType.(map[string]interface{})
	enums := type_["enum"]
	names := enums.([]interface{})
	name := names[0].(string)

	return name, nil
}

func yamlToJsonSchemaLoader(rawSchema []byte) (*gojsonschema.JSONLoader, error) {
	json, err := yaml.YAMLToJSON(rawSchema)
	if err != nil {
		return nil, err
	}
	loader := gojsonschema.NewStringLoader(string(json))

	return &loader, nil
}

func unmarshal(rawSchema []byte) (map[string]interface{}, error) {
	var schemaYaml map[string]interface{}
	err := yaml.Unmarshal(rawSchema, &schemaYaml)
	if err != nil {
		return nil, err
	}
	return schemaYaml, nil
}
