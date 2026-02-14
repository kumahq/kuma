package v1alpha1

import (
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

const maxNameLength = 63

func (r *WorkloadResource) validate() error {
	var verr validators.ValidationError
	name := model.GetDisplayName(r.GetMeta())
	verr.Add(validators.ValidateLength(validators.RootedAt("name"), maxNameLength, name))

	return verr.OrNil()
}
