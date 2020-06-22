package mesh

import (
	"fmt"
	"net"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
)

func (d *DataplaneResource) Validate() error {
	var err validators.ValidationError
	if d.Spec.IsIngress() {
		err.Add(validateIngressNetworking(d.Spec.GetNetworking()))
	} else {
		err.Add(validateNetworking(d.Spec.GetNetworking()))
	}
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
	if net.ParseIP(networking.Address) == nil {
		err.AddViolationAt(path.Field("address"), "address has to be valid IP address")
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

func validateIngressNetworking(networking *mesh_proto.Dataplane_Networking) validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("networking")
	if networking.Gateway != nil {
		err.AddViolationAt(path, "gateway cannot be defined in the ingress mode")
	}
	if len(networking.GetOutbound()) != 0 {
		err.AddViolationAt(path, "dataplane cannot have outbounds in the ingress mode")
	}
	if len(networking.GetInbound()) != 1 {
		err.AddViolationAt(path, "dataplane must have one inbound interface")
	}
	for i, inbound := range networking.GetInbound() {
		p := path.Field("inbound").Index(i)
		if inbound.Port < 1 || inbound.Port > 65535 {
			err.AddViolationAt(p.Field("port"), `port has to be in range of [1, 65535]`)
		}
		if inbound.ServicePort != 0 {
			err.AddViolationAt(p.Field("servicePort"), `cannot be defined in the ingress mode`)
		}
		if inbound.Address != "" {
			err.AddViolationAt(p.Field("address"), `cannot be defined in the ingress mode`)
		}
		err.AddErrorAt(p.Field("address"), validateTags(inbound.Tags))
	}
	for i, ingressInterface := range networking.GetIngress().GetAvailableServices() {
		p := path.Field("ingress").Field("availableService").Index(i)
		if _, ok := ingressInterface.Tags[mesh_proto.ServiceTag]; !ok {
			err.AddViolationAt(p.Field("tags").Key(mesh_proto.ServiceTag), "cannot be empty")
		}
		err.AddErrorAt(p.Field("tags"), validateTags(ingressInterface.GetTags()))
	}
	return err
}

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound, dpAddress string) validators.ValidationError {
	var result validators.ValidationError
	if inbound.Port < 1 || inbound.Port > 65535 {
		result.AddViolationAt(validators.RootedAt("port"), `port has to be in range of [1, 65535]`)
	}
	if inbound.ServicePort > 65535 {
		result.AddViolationAt(validators.RootedAt("servicePort"), `servicePort has to be in range of [0, 65535]`)
	}
	if inbound.Address != "" && net.ParseIP(inbound.Address) == nil {
		result.AddViolationAt(validators.RootedAt("address"), `address has to be valid IP address`)
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
	if outbound.Port < 1 || outbound.Port > 65535 {
		result.AddViolation("port", "port has to be in range of [1, 65535]")
	}
	if outbound.Address != "" && net.ParseIP(outbound.Address) == nil {
		result.AddViolation("address", "address has to be valid IP address")
	}

	if len(outbound.Tags) == 0 {
		if outbound.Service == "" {
			result.AddViolation("service", "cannot be empty")
		}
	} else {
		if _, exist := outbound.Tags[mesh_proto.ServiceTag]; !exist {
			result.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ServiceTag), `tag has to exist`)
		}
		result.Add(validateTags(outbound.Tags))
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
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `tag value cannot be empty`)
		}
		if !tagNameCharacterSet.MatchString(name) {
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `tag name must consist of alphanumeric characters, dots, dashes, slashes and underscores`)
		}
		if !tagValueCharacterSet.MatchString(value) {
			result.AddViolationAt(validators.RootedAt("tags").Key(name), `tag value must consist of alphanumeric characters, dots, dashes and underscores`)
		}
	}
	return result
}
