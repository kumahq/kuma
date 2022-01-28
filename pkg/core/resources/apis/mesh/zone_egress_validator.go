package mesh

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *ZoneEgressResource) Validate() error {
	var err validators.ValidationError
	err.Add(r.validateNetworking(validators.RootedAt("networking"), r.Spec.GetNetworking()))
	return err.OrNil()
}

func (r *ZoneEgressResource) validateNetworking(path validators.PathBuilder, networking *mesh_proto.ZoneEgress_Networking) validators.ValidationError {
	var err validators.ValidationError
	if admin := networking.GetAdmin(); admin != nil {
		if r.UsesInboundInterface(IPv4Loopback, admin.GetPort()) {
			err.AddViolationAt(path.Field("admin").Field("port"), "must differ from port")
		}
	}

	if networking.GetAddress() != "" {
		err.Add(validateAddress(path.Field("address"), networking.GetAddress()))
	}

	err.Add(ValidatePort(path.Field("port"), networking.GetPort()))

	return err
}
