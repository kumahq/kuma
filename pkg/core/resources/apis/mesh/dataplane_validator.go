package mesh

import (
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *DataplaneResource) Validate() error {
	var err validators.ValidationError
	err.Add(validateNetworking(d.Spec.GetNetworking()))
	return err.OrNil()
}

func validateNetworking(networking *mesh_proto.Dataplane_Networking) validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("networking")
	if len(networking.GetInbound()) == 0 && networking.Gateway == nil {
		err.AddViolationAt(path, "has to contain at least one inbound interface or gateway")
	}
	if len(networking.GetInbound()) > 0 && networking.Gateway != nil {
		err.AddViolationAt(path, "inbound cannot be defined both with gateway")
	}
	if networking.Gateway != nil {
		result := validateGateway(networking.Gateway)
		err.AddErrorAt(path.Field("gateway"), result)
	}
	for i, inbound := range networking.GetInbound() {
		result := validateInbound(inbound)
		err.AddErrorAt(path.Field("inbound").Index(i), result)
	}
	for i, outbound := range networking.GetOutbound() {
		result := validateOutbound(outbound)
		err.AddErrorAt(path.Field("outbound").Index(i), result)
	}
	return err
}

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound) validators.ValidationError {
	var result validators.ValidationError
	if _, err := mesh_proto.ParseInboundInterface(inbound.Interface); err != nil {
		result.AddViolation("interface", "invalid format: expected format is DATAPLANE_IP:DATAPLANE_PORT:WORKLOAD_PORT ex. 192.168.0.100:9090:8080")
	}
	if _, exist := inbound.Tags[mesh_proto.ServiceTag]; !exist {
		result.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ServiceTag), `tag has to exist`)
	}
	for name, value := range inbound.Tags {
		if value == "" {
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `tag value cannot be empty`)
		}
	}
	return result
}

func validateOutbound(outbound *mesh_proto.Dataplane_Networking_Outbound) validators.ValidationError {
	var result validators.ValidationError
	if _, err := mesh_proto.ParseOutboundInterface(outbound.Interface); err != nil {
		result.AddViolation("interface", "invalid format: expected format is IP_ADDRESS:PORT where IP_ADDRESS is optional. ex. 192.168.0.100:9090 or :9090")
	}
	if outbound.Service == "" {
		result.AddViolation("service", "cannot be empty")
	}
	return result
}

func validateGateway(gateway *mesh_proto.Dataplane_Networking_Gateway) validators.ValidationError {
	var result validators.ValidationError
	if _, exist := gateway.Tags[mesh_proto.ServiceTag]; !exist {
		result.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ServiceTag), `tag has to exist`)
	}
	for name, value := range gateway.Tags {
		if value == "" {
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `tag value cannot be empty`)
		}
	}
	return result
}
