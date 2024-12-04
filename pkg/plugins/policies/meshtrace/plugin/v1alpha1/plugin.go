package v1alpha1

import (
	net_url "net/url"
	"strconv"
	"strings"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshmultizoneservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

const OriginMeshTrace = "mesh-trace"

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTraceType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *xds.ResourceSet, ctx xds_context.Context, proxy *xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshTraceType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	if err := applyToInbounds(policies.SingleItemRules, listeners.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.SingleItemRules, listeners.Outbound, proxy.Outbounds, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToClusters(policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}
	if err := applyToGateway(policies.SingleItemRules, listeners.Gateway, ctx.Mesh.Resources.MeshLocalResources, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToRealResources(ctx, policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}

	return nil
}

func applyToGateway(
	rules core_rules.SingleItemRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	resources xds_context.ResourceMap,
	dataplane *core_mesh.DataplaneResource,
) error {
	var gateways *core_mesh.MeshGatewayResourceList
	if rawList := resources[core_mesh.MeshGatewayType]; rawList != nil {
		gateways = rawList.(*core_mesh.MeshGatewayResourceList)
	} else {
		return nil
	}

	gateway := xds_topology.SelectGateway(gateways.Items, dataplane.Spec.Matches)
	if gateway == nil {
		return nil
	}

	for _, listener := range gateway.Spec.GetConf().GetListeners() {
		address := dataplane.Spec.GetNetworking().Address
		port := listener.GetPort()
		inboundListener := core_rules.InboundListener{
			Address: address,
			Port:    port,
		}
		listener, ok := gatewayListeners[inboundListener]
		if !ok {
			continue
		}

		if err := configureListener(
			rules,
			dataplane,
			listener,
			"",
		); err != nil {
			return err
		}
	}

	return nil
}

func applyToInbounds(rules core_rules.SingleItemRules, inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource) error {
	for _, inboundListener := range inboundListeners {
		if err := configureListener(rules, dataplane, inboundListener, ""); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(
	rules core_rules.SingleItemRules,
	outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener,
	outbounds xds_types.Outbounds,
	dataplane *core_mesh.DataplaneResource,
) error {
	for _, outbound := range outbounds.Filter(xds_types.NonBackendRefFilter) {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound.LegacyOutbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.LegacyOutbound.GetService()

		if err := configureListener(rules, dataplane, listener, serviceName); err != nil {
			return err
		}
	}

	return nil
}

func applyToRealResources(
	ctx xds_context.Context,
	rules core_rules.SingleItemRules,
	rs *xds.ResourceSet,
	proxy *xds.Proxy,
) error {
	for uri, resType := range rs.IndexByOrigin(xds.NonMeshExternalService) {
		destination, err := getDestinationName(uri, ctx)
		if err != nil {
			return err
		}
		for typ, resources := range resType {
			switch typ {
			case envoy_resource.ListenerType:
				for _, listener := range resources {
					if err := configureListener(rules, proxy.Dataplane, listener.Resource.(*envoy_listener.Listener), destination); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func getDestinationName(uri model.TypedResourceIdentifier, ctx xds_context.Context) (string, error) {
	var destination string
	switch uri.ResourceType {
	case meshservice_api.MeshServiceType:
		ms := ctx.Mesh.MeshServiceByIdentifier[uri.ResourceIdentifier]
		port, ok := ms.FindPortByName(uri.SectionName)
		if !ok {
			return "", errors.Errorf("cannot find port for MeshService %s", ms.Meta.GetName())
		}
		destination = ms.DestinationName(port.Port)
	case meshmultizoneservice_api.MeshMultiZoneServiceType:
		mmzs := ctx.Mesh.MeshMultiZoneServiceByIdentifier[uri.ResourceIdentifier]
		port, ok := mmzs.FindPortByName(uri.SectionName)
		if !ok {
			return "", errors.Errorf("cannot find port for MeshMultiZoneService %s", mmzs.Meta.GetName())
		}
		destination = mmzs.DestinationName(port.Port)
	}
	return destination, nil
}

func configureListener(rules core_rules.SingleItemRules, dataplane *core_mesh.DataplaneResource, listener *envoy_listener.Listener, destination string) error {
	serviceName := dataplane.Spec.GetIdentifyingService()
	if len(rules.Rules) == 0 {
		return nil
	}
	rawConf := rules.Rules[0].Conf
	conf := rawConf.(api.Conf)

	configurer := plugin_xds.Configurer{
		Conf:             conf,
		Service:          serviceName,
		TrafficDirection: listener.TrafficDirection,
		Destination:      destination,
		IsGateway:        dataplane.Spec.IsBuiltinGateway(),
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.Configure(chain); err != nil {
			return err
		}
	}

	return nil
}

func applyToClusters(rules core_rules.SingleItemRules, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	if len(rules.Rules) == 0 {
		return nil
	}
	rawConf := rules.Rules[0].Conf

	conf := rawConf.(api.Conf)

	var backend api.Backend
	if backends := pointer.Deref(conf.Backends); len(backends) == 0 {
		return nil
	} else {
		backend = backends[0]
	}

	var endpoint *xds.Endpoint
	var provider string

	switch {
	case backend.Zipkin != nil:
		endpoint = endpointForZipkin(backend.Zipkin)
		provider = plugin_xds.ZipkinProviderName
	case backend.Datadog != nil:
		endpoint = endpointForDatadog(backend.Datadog)
		provider = plugin_xds.DatadogProviderName
	case backend.OpenTelemetry != nil:
		endpoint = endpointForOpenTelemetry(backend.OpenTelemetry)
		provider = plugin_xds.OpenTelemetryProviderName
	}
	builder := clusters.NewClusterBuilder(proxy.APIVersion, plugin_xds.GetTracingClusterName(provider))

	if backend.OpenTelemetry != nil {
		builder.Configure(clusters.Http2())
	}

	res, err := builder.Configure(clusters.ProvidedEndpointCluster(proxy.Dataplane.IsIPv6(), *endpoint)).
		Configure(clusters.ClientSideTLS([]xds.Endpoint{*endpoint})).
		Configure(clusters.DefaultTimeout()).Build()
	if err != nil {
		return err
	}

	rs.Add(&xds.Resource{Name: plugin_xds.GetTracingClusterName(provider), Origin: OriginMeshTrace, Resource: res})

	return nil
}

func endpointForZipkin(cfg *api.ZipkinBackend) *xds.Endpoint {
	url, _ := net_url.ParseRequestURI(cfg.Url)
	port, _ := strconv.ParseInt(url.Port(), 10, 32)
	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
		ExternalService: &xds.ExternalService{
			TLSEnabled:         url.Scheme == "https",
			AllowRenegotiation: true,
		},
	}
}

func endpointForOpenTelemetry(cfg *api.OpenTelemetryBackend) *xds.Endpoint {
	target := strings.Split(cfg.Endpoint, ":")
	port := uint32(4317) // default gRPC port
	if len(target) > 1 {
		val, _ := strconv.ParseInt(target[1], 10, 32)
		port = uint32(val)
	}
	return &xds.Endpoint{
		Target: target[0],
		Port:   port,
	}
}

func endpointForDatadog(cfg *api.DatadogBackend) *xds.Endpoint {
	url, _ := net_url.ParseRequestURI(cfg.Url)
	port, _ := strconv.ParseInt(url.Port(), 10, 32)

	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
	}
}
