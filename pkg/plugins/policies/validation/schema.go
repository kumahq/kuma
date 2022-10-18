package validation

import (
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func ValidateSchema(spec core_model.ResourceSpec, schema *spec.Schema) error {
	res := validate.NewSchemaValidator(schema, nil, "", strfmt.Default).Validate(spec)
	if res.IsValid() {
		return nil
	}

	var verr validators.ValidationError
	for _, err := range res.Errors {
		verr.AddViolation("spec", err.Error())
	}

	return &verr
}
