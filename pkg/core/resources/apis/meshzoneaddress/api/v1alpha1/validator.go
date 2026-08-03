package v1alpha1

import (
	"github.com/asaskevich/govalidator"

	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

func (r *MeshZoneAddressResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	verr.Add(validateAddress(path.Field("address"), r.Spec.Address))
	verr.Add(validators.ValidatePort(path.Field("port"), uint32(r.Spec.Port)))

	return verr.OrNil()
}

func validateAddress(path validators.PathBuilder, address string) validators.ValidationError {
	var verr validators.ValidationError
	if address == "" {
		verr.AddViolationAt(path, validators.MustNotBeEmpty)
		return verr
	}
	if address == "0.0.0.0" || address == "::" {
		verr.AddViolationAt(path, "must not be 0.0.0.0 or ::")
		return verr
	}
	if !govalidator.IsIP(address) && !govalidator.IsDNSName(address) {
		verr.AddViolationAt(path, "must be a valid IP address or domain name")
	}
	return verr
}
