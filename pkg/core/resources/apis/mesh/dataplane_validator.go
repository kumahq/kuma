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

	if d.Spec.GetNetworking() == nil {
		err.AddViolationAt(net, "must be defined")
		return err.OrNil()
	}

	switch {
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
	err.Add(ValidatePort(path.Field("port"), probes.GetPort()))
	for i, endpoint := range probes.Endpoints {
		indexPath := path.Field("endpoints").Index(i)
		err.Add(ValidatePort(indexPath.Field("inboundPort"), endpoint.GetInboundPort()))
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
	if address == "0.0.0.0" || address == "::" {
		err.AddViolationAt(path.Field("address"), "must not be 0.0.0.0 or ::")
	}
	if !govalidator.IsIP(address) && !govalidator.IsDNSName(address) {
		err.AddViolationAt(path.Field("address"), "address has to be valid IP address or domain name")
	}
	return err
}

func validateInbound(inbound *mesh_proto.Dataplane_Networking_Inbound, dpAddress string) validators.ValidationError {
	var result validators.ValidationError
	result.Add(ValidatePort(validators.RootedAt("port"), inbound.GetPort()))
	if inbound.GetServicePort() != 0 {
		result.Add(ValidatePort(validators.RootedAt("servicePort"), inbound.GetServicePort()))
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

	validateProtocol := func(path validators.PathBuilder, selector map[string]string) validators.ValidationError {
		var result validators.ValidationError
		if value, exist := selector[mesh_proto.ProtocolTag]; exist {
			if ParseProtocol(value) == ProtocolUnknown {
				result.AddViolationAt(
					path.Key(mesh_proto.ProtocolTag), fmt.Sprintf("tag %q has an invalid value %q. %s", mesh_proto.ProtocolTag, value, AllowedValuesHint(SupportedProtocols.Strings()...)),
				)
			}
		}
		return result
	}
	result.Add(ValidateTags(validators.RootedAt("tags"), inbound.Tags, ValidateTagsOpts{
		ExtraTagsValidators: []TagsValidatorFunc{validateProtocol},
	}))

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

	result.Add(ValidatePort(validators.RootedAt("port"), outbound.GetPort()))

	if outbound.Address != "" && net.ParseIP(outbound.Address) == nil {
		result.AddViolation("address", "address has to be valid IP address")
	}

	if len(outbound.Tags) == 0 {
		// nolint:staticcheck
		if outbound.Service == "" {
			result.AddViolationAt(validators.RootedAt("tags"), `mandatory tag "kuma.io/service" is missing`)
		}
	} else {
		result.Add(ValidateTags(validators.RootedAt("tags"), outbound.Tags, ValidateTagsOpts{
			RequireService: true,
		}))
	}

	return result
}

func validateGateway(gateway *mesh_proto.Dataplane_Networking_Gateway) validators.ValidationError {
	var result validators.ValidationError

	validateProtocol := func(path validators.PathBuilder, selector map[string]string) validators.ValidationError {
		var result validators.ValidationError
		if protocol, exist := selector[mesh_proto.ProtocolTag]; exist {
			if protocol != ProtocolTCP {
				result.AddViolationAt(path.Key(mesh_proto.ProtocolTag), `other values than TCP are not allowed`)
			}
		}
		return result
	}
	result.Add(ValidateTags(validators.RootedAt("tags"), gateway.Tags, ValidateTagsOpts{
		RequireService:      true,
		ExtraTagsValidators: []TagsValidatorFunc{validateProtocol}}),
	)

	return result
}
