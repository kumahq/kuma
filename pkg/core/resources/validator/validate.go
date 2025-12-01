package validator

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

type ResourceValidator interface {
	Validate() error
}

func Validate(resource model.Resource) error {
	var verr validators.ValidationError
	if rv, ok := resource.(ResourceValidator); ok {
		if err := rv.Validate(); err != nil {
			if validationErr, ok := err.(*validators.ValidationError); ok {
				verr.Add(*validationErr)
			} else {
				verr.AddViolationAt(validators.Root(), err.Error())
			}
		}
	}
	vals := registry.Global().GetValidators(resource.Descriptor().Name)
	for _, validator := range vals {
		if err := validator.Validate(resource); err != nil {
			if validationErr, ok := err.(*validators.ValidationError); ok {
				verr.Add(*validationErr)
			} else {
				verr.AddViolationAt(validators.Root(), err.Error())
			}
		}
	}
	return nil
}
