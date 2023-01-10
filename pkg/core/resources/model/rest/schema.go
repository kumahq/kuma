package rest

import (
	"strings"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (u *unmarshaler) rawSchemaValidation(bytes []byte, schema *spec.Schema) error {
	rawObj := map[string]interface{}{}

	if err := u.unmarshalFn(bytes, &rawObj); err != nil {
		return err
	}

	var rootSchema *spec.Schema = nil
	var root = ""
	validator := validate.NewSchemaValidator(schema, rootSchema, root, strfmt.Default)

	res := validator.Validate(rawObj)
	if res.IsValid() {
		return nil
	}

	return toValidationError(res)
}

func toValidationError(res *validate.Result) *validators.ValidationError {
	verr := &validators.ValidationError{}
	for _, e := range res.Errors {
		parts := strings.Split(e.Error(), " ")
		if len(parts) > 1 && strings.HasPrefix(parts[0], "spec.") {
			verr.AddViolation(parts[0], strings.Join(parts[1:], " "))
		} else {
			verr.AddViolation("", e.Error())
		}
	}
	return verr
}
