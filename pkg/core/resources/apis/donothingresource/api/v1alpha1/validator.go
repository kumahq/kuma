package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

func (r *DoNothingResourceResource) validate() error {
	var verr validators.ValidationError
	_ = validators.RootedAt("spec")
	return verr.OrNil()
}
