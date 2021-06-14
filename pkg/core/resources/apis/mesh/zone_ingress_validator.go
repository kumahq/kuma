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
	if networking.GetPort() == 0 {
		err.AddViolationAt(path.Field("port"), `port has to be defined`)
	}
	if networking.GetPort() > 65535 {
		err.AddViolationAt(path.Field("port"), `port has to be in range of [1, 65535]`)
	}
	if networking.GetAdvertisedPort() > 65535 {
		err.AddViolationAt(path.Field("advertisedPort"), `port has to be in range of [1, 65535]`)
	}
	return err
}

func (r *ZoneIngressResource) validateAvailableServices(path validators.PathBuilder, availableServices []*mesh_proto.ZoneIngress_AvailableService) validators.ValidationError {
	var err validators.ValidationError
	for i, availableService := range availableServices {
		p := path.Index(i)
		if _, ok := availableService.Tags[mesh_proto.ServiceTag]; !ok {
			err.AddViolationAt(p.Field("tags").Key(mesh_proto.ServiceTag), "cannot be empty")
		}
		err.AddErrorAt(p.Field("tags"), validateTags(availableService.GetTags()))
	}
	return err
}
