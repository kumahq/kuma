package mesh

import (
	"fmt"
	"net"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *DataplaneResource) Validate() error {
	var err validators.ValidationError
	err.Add(validateNetworking(d.Spec.GetNetworking()))
	return err.OrNil()
}

// For networking section validation we need to take into account our legacy model.
// Legacy model is detected by having interface defined on inbound listeners.
// We do not allow networking.address with the old format. Instead, we recommend switching to the new format.
// When we've got dataplane in the new format, we require networking.address field to be defined.
func validateNetworking(networking *mesh_proto.Dataplane_Networking) validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("networking")
	if len(networking.GetInbound()) == 0 && networking.Gateway == nil {
		err.AddViolationAt(path, "has to contain at least one inbound interface or gateway")
	}
	if len(networking.GetInbound()) > 0 && networking.Gateway != nil {
		err.AddViolationAt(path, "inbound cannot be defined both with gateway")
	}
	// backwards compatibility validate networking.address only when all inbounds are in new format
	if networking.Address != "" || !hasLegacyInbound(networking) {
		if net.ParseIP(networking.Address) == nil {
			err.AddViolationAt(path.Field("address"), "address has to be valid IP address")
		}
	}
	if networking.Gateway != nil {
		result := validateGateway(networking.Gateway)
		err.AddErrorAt(path.Field("gateway"), result)
	}
	for i, inbound := range networking.GetInbound() {
		result := validateInbound(inbound, networking.Address)
		err.AddErrorAt(path.Field("inbound").Index(i), result)
	}
	for i, outbound := range networking.GetOutbound() {
		result := validateOutbound(outbound)
		err.AddErrorAt(path.Field("outbound").Index(i), result)
	}
	return err
}

func hasLegacyInbound(networking *mesh_proto.Dataplane_Networking) bool {
	for _, inbound := range networking.GetInbound() {
		if inbound.Interface != "" && inbound.Port == 0 && inbound.ServicePort == 0 && inbound.Address == "" {
			return true
		}
	}
	return false
}

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound, dpAddress string) validators.ValidationError {
	var result validators.ValidationError
	if inbound.Interface != "" { // for backwards compatibility
		if inbound.Port != 0 {
			result.AddViolationAt(validators.RootedAt("interface"), `interface cannot be defined with port. Replace it with port, servicePort and networking.address`)
		}
		if inbound.ServicePort != 0 {
			result.AddViolationAt(validators.RootedAt("interface"), `interface cannot be defined with servicePort. Replace it with port, servicePort and networking.address`)
		}
		if inbound.Address != "" {
			result.AddViolationAt(validators.RootedAt("interface"), `interface cannot be defined with address. Replace it with port, servicePort and networking.address`)
		}
		if dpAddress != "" {
			result.AddViolationAt(validators.RootedAt("interface"), `interface cannot be defined with networking.address. Replace it with port, servicePort and networking.address`)
		}
		if _, err := mesh_proto.ParseInboundInterface(inbound.Interface); err != nil {
			result.AddViolation("interface", "invalid format: expected format is DATAPLANE_IP:DATAPLANE_PORT:WORKLOAD_PORT , e.g. 192.168.0.100:9090:8080 or [2001:db8::1]:7070:6060")
		}
	} else {
		if inbound.Port < 1 || inbound.Port > 65535 {
			result.AddViolationAt(validators.RootedAt("port"), `port has to be in range of [1, 65535]`)
		}
		if inbound.ServicePort > 65535 {
			result.AddViolationAt(validators.RootedAt("servicePort"), `servicePort has to be in range of [0, 65535]`)
		}
		if inbound.Address != "" && net.ParseIP(inbound.Address) == nil {
			result.AddViolationAt(validators.RootedAt("address"), `address has to be valid IP address`)
		}
	}
	if _, exist := inbound.Tags[mesh_proto.ServiceTag]; !exist {
		result.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ServiceTag), `tag has to exist`)
	}
	if value, exist := inbound.Tags[mesh_proto.ProtocolTag]; exist {
		if ParseProtocol(value) == ProtocolUnknown {
			result.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ProtocolTag), fmt.Sprintf("tag %q has an invalid value %q. %s", mesh_proto.ProtocolTag, value, AllowedValuesHint(SupportedProtocols.Strings()...)))
		}
	}
	result.Add(validateTags(inbound.Tags))
	return result
}

func validateOutbound(outbound *mesh_proto.Dataplane_Networking_Outbound) validators.ValidationError {
	var result validators.ValidationError
	if outbound.Interface != "" { // for backwards compatibility
		if outbound.Port != 0 {
			result.AddViolation("interface", "interface cannot be defined with port. Replace it with port and address")
		}
		if outbound.Address != "" {
			result.AddViolation("interface", "interface cannot be defined with address. Replace it with port and address")
		}
		if _, err := mesh_proto.ParseOutboundInterface(outbound.Interface); err != nil {
			result.AddViolation("interface", "invalid format: expected format is DATAPLANE_IP:DATAPLANE_PORT where DATAPLANE_IP is optional. E.g. 127.0.0.1:9090, :9090, [::1]:8080")
		}
	} else {
		if outbound.Port < 1 || outbound.Port > 65535 {
			result.AddViolation("port", "port has to be in range of [1, 65535]")
		}
		if outbound.Address != "" && net.ParseIP(outbound.Address) == nil {
			result.AddViolation("address", "address has to be valid IP address")
		}
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
	result.Add(validateTags(gateway.Tags))
	return result
}

func validateTags(tags map[string]string) validators.ValidationError {
	var result validators.ValidationError
	for name, value := range tags {
		if value == "" {
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `value cannot be empty`)
		}
		if !nameCharacterSet.MatchString(name) {
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `key must consist of alphanumeric characters, dots, dashes and underscores`)
		}
		if !nameCharacterSet.MatchString(value) {
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `value must consist of alphanumeric characters, dots, dashes and underscores`)
		}
	}
	return result
}
