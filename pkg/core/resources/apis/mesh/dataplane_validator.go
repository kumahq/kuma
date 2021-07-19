package mesh

import (
	"fmt"
	"net"
	"net/url"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
)

// allowBuiltinGateways specifies whether the dataplanes of builtin
// gateway type are allowed. This is controlled by the +gateway build
// conditional.
var allowBuiltinGateways = false

func (d *DataplaneResource) Validate() error {
	var err validators.ValidationError

	net := validators.RootedAt("networking")

	switch {
	case d.Spec.IsIngress():
		err.Add(validateIngressNetworking(d.Spec.GetNetworking()))
		err.Add(validateIngress(net.Field("ingress"), d.Spec.GetNetworking().GetIngress()))

	case d.Spec.IsDelegatedGateway():
		if len(d.Spec.GetNetworking().GetInbound()) > 0 {
			err.AddViolationAt(net.Field("inbound"),
				"inbound cannot be defined for delegated gateways")
		}

		err.AddErrorAt(net.Field("gateway"), validateGateway(d.Spec.GetNetworking().GetGateway()))
		err.Add(validateNetworking(d.Spec.GetNetworking()))
		err.Add(validateProbes(d.Spec.GetProbes()))

	case d.Spec.IsBuiltinGateway():
		if !allowBuiltinGateways {
			err.AddViolationAt(net.Field("gateway"), "unsupported gateway type")
			return err.OrNil()
		}

		if len(d.Spec.GetNetworking().GetInbound()) > 0 {
			err.AddViolationAt(net.Field("inbound"), "inbound cannot be defined for builtin gateways")
		}

		if len(d.Spec.GetNetworking().GetOutbound()) > 0 {
			err.AddViolationAt(net.Field("outbound"), "outbound cannot be defined for builtin gateways")
		}

		if d.Spec.GetProbes() != nil {
			err.AddViolationAt(net.Field("probes"), "probes cannot be defined for builtin gateways")
		}

		err.AddErrorAt(net.Field("gateway"), validateGateway(d.Spec.GetNetworking().GetGateway()))
		err.Add(validateNetworking(d.Spec.GetNetworking()))

	default:
		if len(d.Spec.GetNetworking().GetInbound()) == 0 {
			err.AddViolationAt(net, "has to contain at least one inbound interface or gateway")
		}
		err.Add(validateNetworking(d.Spec.GetNetworking()))
		err.Add(validateProbes(d.Spec.GetProbes()))
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
	err.Add(validateAddress(path, networking.Address))
	for i, inbound := range networking.GetInbound() {
		field := path.Field("inbound").Index(i)
		result := validateInbound(inbound, networking.Address)
		err.AddErrorAt(field, result)
		if _, exist := inbound.Tags[mesh_proto.ServiceTag]; !exist {
			err.AddViolationAt(field.Field("tags").Key(mesh_proto.ServiceTag), `tag has to exist`)
		}
	}
	for i, outbound := range networking.GetOutbound() {
		result := validateOutbound(outbound)
		err.AddErrorAt(path.Field("outbound").Index(i), result)
	}
	return err
}

func validateProbes(probes *mesh_proto.Dataplane_Probes) validators.ValidationError {
	if probes == nil {
		return validators.ValidationError{}
	}
	var err validators.ValidationError
	path := validators.RootedAt("probes")
	if probes.Port < 1 || probes.Port > 65535 {
		err.AddViolationAt(path.Field("port"), `port has to be in range of [1, 65535]`)
	}
	for i, endpoint := range probes.Endpoints {
		indexPath := path.Field("endpoints").Index(i)
		if endpoint.InboundPort < 1 || endpoint.InboundPort > 65535 {
			err.AddViolationAt(indexPath.Field("inboundPort"), `port has to be in range of [1, 65535]`)
		}
		if _, URIErr := url.ParseRequestURI(endpoint.InboundPath); URIErr != nil {
			err.AddViolationAt(indexPath.Field("inboundPath"), `should be a valid URL Path`)
		}
		if _, URIErr := url.ParseRequestURI(endpoint.Path); URIErr != nil {
			err.AddViolationAt(indexPath.Field("path"), `should be a valid URL Path`)
		}
	}
	return err
}

func validateAddress(path validators.PathBuilder, address string) validators.ValidationError {
	var err validators.ValidationError
	if address == "" {
		err.AddViolationAt(path.Field("address"), "address can't be empty")
		return err
	}
	if !govalidator.IsIP(address) && !govalidator.IsDNSName(address) {
		err.AddViolationAt(path.Field("address"), "address has to be valid IP address or domain name")
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
		if inbound.Port > 65535 {
			err.AddViolationAt(p.Field("port"), `port has to be in range of [1, 65535]`)
		}
		if inbound.ServicePort != 0 {
			err.AddViolationAt(p.Field("servicePort"), `cannot be defined in the ingress mode`)
		}
		if inbound.ServiceAddress != "" {
			err.AddViolationAt(p.Field("serviceAddress"), `cannot be defined in the ingress mode`)
		}
		if inbound.Address != "" {
			err.AddViolationAt(p.Field("address"), `cannot be defined in the ingress mode`)
		}
		err.AddErrorAt(p.Field("tags"), validateTags(inbound.Tags))
		if protocol, exist := inbound.Tags[mesh_proto.ProtocolTag]; exist {
			if protocol != ProtocolTCP {
				err.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ProtocolTag), `other values than TCP are not allowed`)
			}
		}
	}
	return err
}

func validateIngress(path validators.PathBuilder, ingress *mesh_proto.Dataplane_Networking_Ingress) validators.ValidationError {
	if ingress == nil {
		return validators.ValidationError{}
	}
	var err validators.ValidationError
	if ingress.GetPublicAddress() == "" && ingress.GetPublicPort() != 0 {
		err.AddViolationAt(path.Field("publicAddress"), `has to be defined with publicPort`)
	}
	if ingress.GetPublicPort() == 0 && ingress.GetPublicAddress() != "" {
		err.AddViolationAt(path.Field("publicPort"), `has to be defined with publicAddress`)
	}
	if ingress.GetPublicAddress() != "" {
		err.Add(validateAddress(path.Field("publicAddress"), ingress.GetPublicAddress()))
	}
	if ingress.GetPublicPort() > 65535 {
		err.AddViolationAt(path.Field("publicPort"), `port has to be in range of [1, 65535]`)
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

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound, dpAddress string) validators.ValidationError {
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
	if value, exist := inbound.Tags[mesh_proto.ProtocolTag]; exist {
		if ParseProtocol(value) == ProtocolUnknown {
			result.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ProtocolTag), fmt.Sprintf("tag %q has an invalid value %q. %s", mesh_proto.ProtocolTag, value, AllowedValuesHint(SupportedProtocols.Strings()...)))
		}
	}
	result.Add(validateTags(inbound.Tags))
	result.Add(validateServiceProbe(inbound.ServiceProbe))
	return result
}

func validateServiceProbe(serviceProbe *mesh_proto.Dataplane_Networking_Inbound_ServiceProbe) (err validators.ValidationError) {
	if serviceProbe == nil {
		return
	}
	path := validators.RootedAt("serviceProbe")
	if serviceProbe.Interval != nil {
		err.Add(ValidateDuration(path.Field("interval"), serviceProbe.Interval))
	}
	if serviceProbe.Timeout != nil {
		err.Add(ValidateDuration(path.Field("timeout"), serviceProbe.Timeout))
	}
	if serviceProbe.UnhealthyThreshold != nil {
		err.Add(ValidateThreshold(path.Field("unhealthyThreshold"), serviceProbe.UnhealthyThreshold.GetValue()))
	}
	if serviceProbe.HealthyThreshold != nil {
		err.Add(ValidateThreshold(path.Field("healthyThreshold"), serviceProbe.HealthyThreshold.GetValue()))
	}
	return
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
		// nolint:staticcheck
		if outbound.Service == "" {
			result.AddViolation("kuma.io/service", "cannot be empty")
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
	if protocol, exist := gateway.Tags[mesh_proto.ProtocolTag]; exist {
		if protocol != ProtocolTCP {
			result.AddViolationAt(validators.RootedAt("tags").Key(mesh_proto.ProtocolTag), `other values than TCP are not allowed`)
		}
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
