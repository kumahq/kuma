package mesh

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/maps"
)

var allowedKinds = map[string]struct{}{
	string(common_api.MeshService):          {},
	string(common_api.MeshExternalService):  {},
	string(common_api.MeshMultiZoneService): {},
}

func (d *DataplaneResource) Validate() error {
	var err validators.ValidationError

	net := validators.RootedAt("networking")

	if d.Spec.GetNetworking() == nil {
		err.AddViolationAt(net, "must be defined")
		return err.OrNil()
	}

	if admin := d.Spec.GetNetworking().GetAdmin(); admin != nil {
		adminPort := net.Field("admin").Field("port")

		if d.UsesInboundInterface(IPv4Loopback, admin.GetPort()) {
			err.AddViolationAt(adminPort, "must differ from inbound")
		}
		if d.UsesOutboundInterface(IPv4Loopback, admin.GetPort()) {
			err.AddViolationAt(adminPort, "must differ from outbound")
		}
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
		if d.Spec.GetMetrics() != nil {
			err.Add(validateMetricsBackend(d.Spec.GetMetrics()))
		}
	}

	return err.OrNil()
}

// For networking section validation we need to take into account our legacy model.
// Sotw model is detected by having interface defined on inbound listeners.
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
	err.AddErrorAt(path.Field("transparentProxing"), validateTransparentProxying(networking.GetTransparentProxying()))
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

func validateMetricsBackend(metrics *mesh_proto.MetricsBackend) validators.ValidationError {
	var verr validators.ValidationError
	if metrics.GetType() != mesh_proto.MetricsPrometheusType {
		verr.AddViolationAt(validators.RootedAt("metrics").Field("type"), fmt.Sprintf("unknown backend type. Available backends: %q", mesh_proto.MetricsPrometheusType))
	} else {
		verr.AddErrorAt(validators.RootedAt("metrics").Field("conf"), validatePrometheusConfig(metrics.GetConf()))
	}
	return verr
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

func validateServiceProbe(serviceProbe *mesh_proto.Dataplane_Networking_Inbound_ServiceProbe) validators.ValidationError {
	var err validators.ValidationError
	if serviceProbe == nil {
		return err
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
	return err
}

func validateOutbound(outbound *mesh_proto.Dataplane_Networking_Outbound) validators.ValidationError {
	var result validators.ValidationError

	result.Add(ValidatePort(validators.RootedAt("port"), outbound.GetPort()))

	if outbound.Address != "" && net.ParseIP(outbound.Address) == nil {
		result.AddViolation("address", "address has to be valid IP address")
	}

	switch {
	case outbound.BackendRef != nil:
		if _, allowed := allowedKinds[outbound.BackendRef.Kind]; !allowed {
			result.AddViolation("backendRef.kind", fmt.Sprintf("invalid value. Available values are: %s", strings.Join(maps.SortedKeys(allowedKinds), ",")))
		}
		if outbound.BackendRef.Name == "" && len(outbound.BackendRef.Labels) == 0 {
			result.AddViolation("backendRef", "either 'name' or 'labels' should be specified")
		}
		// for MeshExternalService the port does not matter because it's taken from endpoints
		if outbound.BackendRef.Kind != string(common_api.MeshExternalService) {
			result.Add(ValidatePort(validators.RootedAt("backendRef").Field("port"), outbound.BackendRef.Port))
		}
	case len(outbound.Tags) == 0:
		// nolint:staticcheck
		if outbound.GetService() == "" {
			result.AddViolationAt(validators.RootedAt("tags"), `mandatory tag "kuma.io/service" is missing`)
		}
	default:
		result.Add(ValidateTags(validators.RootedAt("tags"), outbound.Tags, ValidateTagsOpts{
			RequireService: true,
		}))
	}

	if outbound.BackendRef != nil && (len(outbound.Tags) != 0 || outbound.GetService() != "") {
		result.AddViolationAt(validators.RootedAt("backendRef"), "both backendRef and tags/service cannot be defined")
	}

	return result
}

func validateTransparentProxying(tp *mesh_proto.Dataplane_Networking_TransparentProxying) validators.ValidationError {
	var result validators.ValidationError
	path := validators.RootedAt("reachableBackends.refs")
	if tp != nil && tp.ReachableBackends != nil {
		for i, backendRef := range tp.ReachableBackends.Refs {
			switch backendRef.Kind {
			case string(common_api.MeshMultiZoneService), string(common_api.MeshService), string(common_api.MeshExternalService):
			default:
				result.AddViolationAt(path.Index(i).Field("kind"), fmt.Sprintf("invalid value. Available values are: %s", strings.Join(maps.SortedKeys(allowedKinds), ",")))
			}
			if backendRef.Name != "" {
				result.AddErrorAt(path.Index(i).Field("name"), validateIdentifier(backendRef.Name, identifierRegexp, identifierErrMsg))
			}
			if backendRef.Name == "" && backendRef.Namespace == "" && len(backendRef.Labels) == 0 {
				result.AddViolationAt(path.Index(i).Field("name"), "name or labels are required")
			}
			if backendRef.Name == "" && backendRef.Namespace != "" {
				result.AddViolationAt(path.Index(i).Field("name"), "name is required, when namespace is defined")
			}
			if (backendRef.Name != "" || backendRef.Namespace != "") && len(backendRef.Labels) > 0 {
				result.AddViolationAt(path.Index(i).Field("labels"), "labels cannot be defined when name is specified")
			}
		}
	}
	return result
}

func validateGateway(gateway *mesh_proto.Dataplane_Networking_Gateway) validators.ValidationError {
	var result validators.ValidationError

	validateProtocol := func(path validators.PathBuilder, selector map[string]string) validators.ValidationError {
		var result validators.ValidationError
		if protocol, exist := selector[mesh_proto.ProtocolTag]; exist {
			if protocol != ProtocolTCP {
				result.AddViolationAt(path.Key(mesh_proto.ProtocolTag), fmt.Sprintf(`other values than tcp are not allowed, provided value "%s"`, protocol))
			}
		}
		return result
	}
	result.Add(ValidateTags(validators.RootedAt("tags"), gateway.Tags, ValidateTagsOpts{
		RequireService:      true,
		ExtraTagsValidators: []TagsValidatorFunc{validateProtocol},
	}),
	)

	return result
}
