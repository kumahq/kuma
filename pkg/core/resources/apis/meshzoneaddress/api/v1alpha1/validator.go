package v1alpha1

import (
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

func (r *MeshZoneAddressResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	if r.Spec.Address == "" {
		verr.AddViolationAt(path.Field("address"), validators.MustNotBeEmpty)
	}
	verr.Add(validators.ValidatePort(path.Field("port"), uint32(r.Spec.Port)))

	return verr.OrNil()
}
