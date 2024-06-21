package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshServiceResource) validate() error {
	var verr validators.ValidationError

	verr.Add(validators.ValidateLength(validators.RootedAt("name"), 63, r.GetMeta().GetName()))

	return verr.OrNil()
}
