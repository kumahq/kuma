package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *ZoneIngressResource) Validate() error {
	var err validators.ValidationError
	err.Add(validateZoneIngress(validators.RootedAt("networking"), r.Spec))
	return err.OrNil()
}

func validateZoneIngress(path validators.PathBuilder, ingress *mesh_proto.ZoneIngress) validators.ValidationError {
	if ingress == nil {
		return validators.ValidationError{}
	}
	var err validators.ValidationError
	if ingress.GetNetworking().GetAdvertisedAddress() == "" && ingress.GetNetworking().GetAdvertisedPort() != 0 {
		err.AddViolationAt(path.Field("advertisedAddress"), `has to be defined with advertisedPort`)
	}
	if ingress.GetNetworking().GetAdvertisedPort() == 0 && ingress.GetNetworking().GetAdvertisedAddress() != "" {
		err.AddViolationAt(path.Field("advertisedPort"), `has to be defined with advertisedAddress`)
	}
	if ingress.GetNetworking().GetAddress() != "" {
		err.Add(validateAddress(path.Field("address"), ingress.GetNetworking().GetAddress()))
	}
	if ingress.GetNetworking().GetAdvertisedAddress() != "" {
		err.Add(validateAddress(path.Field("advertisedAddress"), ingress.GetNetworking().GetAdvertisedAddress()))
	}
	if ingress.GetNetworking().GetPort() == 0 {
		err.AddViolationAt(path.Field("port"), `port has to be defined`)
	}
	if ingress.GetNetworking().GetPort() > 65535 {
		err.AddViolationAt(path.Field("port"), `port has to be in range of [1, 65535]`)
	}
	if ingress.GetNetworking().GetAdvertisedPort() > 65535 {
		err.AddViolationAt(path.Field("advertisedPort"), `port has to be in range of [1, 65535]`)
	}
	for i, ingressInterface := range ingress.GetAvailableServices() {
		p := path.Field("availableService").Index(i)
		if _, ok := ingressInterface.Tags[mesh_proto.ServiceTag]; !ok {
			err.AddViolationAt(p.Field("tags").Key(mesh_proto.ServiceTag), "cannot be empty")
		}
		err.AddErrorAt(p.Field("tags"), validateTags(ingressInterface.GetTags()))
	}
	return err
}
