package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *ZoneIngressResource) Validate() error {
	var err validators.ValidationError
	err.Add(r.validateNetworking(validators.RootedAt("networking"), r.Spec.GetNetworking()))
	err.Add(r.validateAvailableServices(validators.RootedAt("availableService"), r.Spec.GetAvailableServices()))
	return err.OrNil()
}

func (r *ZoneIngressResource) validateNetworking(path validators.PathBuilder, networking *mesh_proto.ZoneIngress_Networking) validators.ValidationError {
	var err validators.ValidationError
	if admin := networking.GetAdmin(); admin != nil {
		if r.UsesInboundInterface(IPv4Loopback, admin.GetPort()) {
			err.AddViolationAt(path.Field("admin").Field("port"), "must differ from port")
		}
	}
	if networking.GetAdvertisedAddress() == "" && networking.GetAdvertisedPort() != 0 {
		err.AddViolationAt(path.Field("advertisedAddress"), `has to be defined with advertisedPort`)
	}
	if networking.GetAdvertisedPort() == 0 && networking.GetAdvertisedAddress() != "" {
		err.AddViolationAt(path.Field("advertisedPort"), `has to be defined with advertisedAddress`)
	}
	if networking.GetAddress() != "" {
		err.Add(validateAddress(path.Field("address"), networking.GetAddress()))
	}
	if networking.GetAdvertisedAddress() != "" {
		err.Add(validateAddress(path.Field("advertisedAddress"), networking.GetAdvertisedAddress()))
	}

	err.Add(ValidatePort(path.Field("port"), networking.GetPort()))

	if networking.GetAdvertisedPort() != 0 {
		err.Add(ValidatePort(path.Field("advertisedPort"), networking.GetAdvertisedPort()))
	}

	return err
}

func (r *ZoneIngressResource) validateAvailableServices(path validators.PathBuilder, availableServices []*mesh_proto.ZoneIngress_AvailableService) validators.ValidationError {
	var err validators.ValidationError
	for i, availableService := range availableServices {
		p := path.Index(i)
		err.Add(ValidateTags(p.Field("tags"), availableService.Tags, ValidateTagsOpts{
			RequireService: true,
		}))
	}
	return err
}
