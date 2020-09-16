package mesh

import (
	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (d *ExternalServiceResource) Validate() error {
	var err validators.ValidationError
	err.Add(validateExternalServiceNetworking(d.Spec.GetNetworking()))

	return err.OrNil()
}

func validateExternalServiceNetworking(networking *mesh_proto.ExternalService_Networking) validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("networking")
	err.Add(validateExtarnalServiceAddress(path, networking.Address))

	return err
}

func validateExtarnalServiceAddress(path validators.PathBuilder, address string) validators.ValidationError {
	var err validators.ValidationError
	if address == "" {
		err.AddViolationAt(path.Field("address"), "address can't be empty")
		return err
	}
	if !govalidator.IsDNSName(address) {
		err.AddViolationAt(path.Field("address"), "address has to be valid IP address or domain name")
	}
	return err
}
