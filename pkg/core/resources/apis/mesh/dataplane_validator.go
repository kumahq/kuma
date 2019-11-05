package mesh

import (
	"fmt"
	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *DataplaneResource) Validate() error {
	var err validators.ValidationError
	err.AddError("networking", validateNetworking(d.Spec.GetNetworking()))
	return err.ToError()
}

func validateNetworking(networking *mesh_proto.Dataplane_Networking) validators.ValidationError {
	var err validators.ValidationError
	if len(networking.GetInbound()) == 0 {
		err.AddViolation("inbound", "has to contain at least one inbound interface")
	}
	for i, inbound := range networking.GetInbound() {
		result := validateInbound(inbound)
		err.AddError(fmt.Sprintf("inbound[%d]", i), result)
	}
	for i, outbound := range networking.GetOutbound() {
		result := validateOutbound(outbound)
		err.AddError(fmt.Sprintf("outbound[%d]", i), result)
	}
	return err
}

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound) validators.ValidationError {
	var result validators.ValidationError
	if _, err := mesh_proto.ParseInboundInterface(inbound.Interface); err != nil {
		result.AddViolation("interface", "invalid format: expected format is DATAPLANE_IP:DATAPLANE_PORT:WORKLOAD_PORT ex. 192.168.0.100:9090:8080")
	}
	if _, exist := inbound.Tags[mesh_proto.ServiceTag]; !exist {
		result.AddViolation(fmt.Sprintf(`tags["%s"]`, mesh_proto.ServiceTag), `tag has to exist`)
	}
	for name, value := range inbound.Tags {
		if value == "" {
			result.AddViolation(fmt.Sprintf(`tags["%s"]`, name), `tag value cannot be empty`)
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
