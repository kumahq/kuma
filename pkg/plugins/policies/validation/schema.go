package validation

import (
	"github.com/ghodss/yaml"
	"github.com/xeipuuv/gojsonschema"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func ValidateSchema(message proto.Message, schema *gojsonschema.JSONLoader) error {
	json, err := protojson.Marshal(message)
	if err != nil {
		return err
	}

	documentLoader := gojsonschema.NewBytesLoader(json)
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

func YamlToJsonSchemaLoader(rawSchema []byte) (*gojsonschema.JSONLoader, error) {
	json, err := yaml.YAMLToJSON(rawSchema)
	if err != nil {
		return nil, err
	}
	loader := gojsonschema.NewStringLoader(string(json))

	return &loader, nil
}
