package mesh

import (
	"fmt"
	"net"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (d *ExternalServiceResource) Validate() error {
	var err validators.ValidationError
	err.Add(validateExternalServiceNetworking(d.Spec.GetNetworking()))

	return err.OrNil()
}

// For networking section validation we need to take into account our legacy model.
// Legacy model is detected by having interface defined on inbound listeners.
// We do not allow networking.address with the old format. Instead, we recommend switching to the new format.
// When we've got dataplane in the new format, we require networking.address field to be defined.
func validateExternalServiceNetworking(networking *mesh_proto.ExternalService_Networking) validators.ValidationError {
	var err validators.ValidationError
	path := validators.RootedAt("networking")
	if len(networking.GetInbound()) == 0 {
		err.AddViolationAt(path, "has to contain at least one inbound interface")
	}
	err.Add(validateExtarnalServiceAddress(path, networking.Address))

	for i, inbound := range networking.GetInbound() {
		result := validateExternalServiceInbound(inbound, networking.Address)
		err.AddErrorAt(path.Field("inbound").Index(i), result)
	}
	return err
}

func validateExtarnalServiceAddress(path validators.PathBuilder, address string) validators.ValidationError {
	var err validators.ValidationError
	if address == "" {
		err.AddViolationAt(path.Field("address"), "address can't be empty")
		return err
	}
	if !DNSRegex.MatchString(address) {
		err.AddViolationAt(path.Field("address"), "address has to be valid IP address or domain name")
	}
	return err
}

func validateExternalServiceInbound(inbound *mesh_proto.ExternalService_Networking_Inbound, dpAddress string) validators.ValidationError {
	var result validators.ValidationError
	if inbound.Port < 1 || inbound.Port > 65535 {
		result.AddViolationAt(validators.RootedAt("port"), `port has to be in range of [1, 65535]`)
	}
	if inbound.ServicePort > 65535 {
		result.AddViolationAt(validators.RootedAt("servicePort"), `servicePort has to be in range of [0, 65535]`)
	}
	if inbound.ServiceAddress != "" {
		if net.ParseIP(inbound.ServiceAddress) == nil {
			result.AddViolationAt(validators.RootedAt("serviceAddress"), `serviceAddress has to be valid IP address`)
		}
		if inbound.ServiceAddress == dpAddress {
			if inbound.ServicePort == 0 || inbound.ServicePort == inbound.Port {
				result.AddViolationAt(validators.RootedAt("serviceAddress"), `serviceAddress and servicePort has to differ from address and port`)
			}
		}
	}
	if inbound.Address != "" {
		if net.ParseIP(inbound.Address) == nil {
			result.AddViolationAt(validators.RootedAt("address"), `address has to be valid IP address`)
		}
		if inbound.Address == inbound.ServiceAddress {
			if inbound.ServicePort == 0 || inbound.ServicePort == inbound.Port {
				result.AddViolationAt(validators.RootedAt("serviceAddress"), `serviceAddress and servicePort has to differ from address and port`)
			}
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
