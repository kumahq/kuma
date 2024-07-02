package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshServiceResource) validate() error {
	var verr validators.ValidationError

	name := model.GetDisplayName(r.GetMeta())
	verr.Add(validators.ValidateLength(validators.RootedAt("name"), maxNameLength, name))

	return verr.OrNil()
}
