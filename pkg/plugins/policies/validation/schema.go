package validation

import (
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func ValidateSchema(rspec core_model.ResourceSpec, schema *spec.Schema) error {
	var rootSchema *spec.Schema = nil
	var root = ""
	validator := validate.NewSchemaValidator(schema, rootSchema, root, strfmt.Default)

	res := validator.Validate(rspec)
	if res.IsValid() {
		return nil
	}

	var verr validators.ValidationError
	for _, err := range res.Errors {
		verr.AddViolation("spec", err.Error())
	}

	return &verr
}
