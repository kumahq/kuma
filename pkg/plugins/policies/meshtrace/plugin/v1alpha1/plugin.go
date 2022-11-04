package v1alpha1

import (
	net_url "net/url"
	"strconv"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	xds_topology "github.com/kumahq/kuma/pkg/xds/topology"
)

const OriginMeshTrace = "mesh-trace"

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTraceType, dataplane, resources)
}

func (p plugin) Apply(rs *xds.ResourceSet, ctx xds_context.Context, proxy *xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshTraceType]
	if !ok {
		return nil
	}
	if len(policies.SingleItemRules.Rules) == 0 {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	if err := applyToInbounds(policies.SingleItemRules, listeners.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.SingleItemRules, listeners.Outbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToClusters(policies.SingleItemRules, rs, proxy); err != nil {
		return err
	}
	if err := applyToGateway(policies.SingleItemRules, listeners.Gateway, ctx.Mesh.Resources.MeshLocalResources, proxy.Dataplane); err != nil {
		return err
	}

	return nil
}

func applyToGateway(
	rules xds.SingleItemRules, gatewayListeners map[xds.InboundListener]*envoy_listener.Listener, resources xds_context.ResourceMap, dataplane *core_mesh.DataplaneResource,
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
		listener, ok := gatewayListeners[xds.InboundListener{
			Address: address,
			Port:    port,
		}]
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

func applyToInbounds(rules xds.SingleItemRules, inboundListeners map[xds.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource) error {
	for _, inboundListener := range inboundListeners {
		if err := configureListener(rules, dataplane, inboundListener, ""); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(rules xds.SingleItemRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]

		if err := configureListener(rules, dataplane, listener, serviceName); err != nil {
			return err
		}
	}

	return nil
}

func configureListener(rules xds.SingleItemRules, dataplane *core_mesh.DataplaneResource, listener *envoy_listener.Listener, destination string) error {
	serviceName := dataplane.Spec.GetIdentifyingService()
	rawConf := rules.Rules[0].Conf
	conf := rawConf.(api.Conf)

	configurer := plugin_xds.Configurer{
		Conf:             conf,
		Service:          serviceName,
		TrafficDirection: listener.TrafficDirection,
		Destination:      destination,
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.Configure(chain); err != nil {
			return err
		}
	}

	return nil
}

func applyToClusters(rules xds.SingleItemRules, rs *xds.ResourceSet, proxy *xds.Proxy) error {
	rawConf := rules.Rules[0].Conf

	conf := rawConf.(api.Conf)

	backend := conf.Backends[0]

	var endpoint *xds.Endpoint
	var provider string

	if backend.Zipkin != nil {
		endpoint = endpointForZipkin(backend.Zipkin)
		provider = plugin_xds.ZipkinProviderName
	} else {
		endpoint = endpointForDatadog(backend.Datadog)
		provider = plugin_xds.DatadogProviderName
	}

	res, err := clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(clusters.ProvidedEndpointCluster(plugin_xds.GetTracingClusterName(provider), proxy.Dataplane.IsIPv6(), *endpoint)).
		Configure(clusters.ClientSideTLS([]xds.Endpoint{*endpoint})).
		Configure(clusters.DefaultTimeout()).
		Build()
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

func endpointForDatadog(cfg *api.DatadogBackend) *xds.Endpoint {
	url, _ := net_url.ParseRequestURI(cfg.Url)
	port, _ := strconv.ParseInt(url.Port(), 10, 32)

	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
	}
}
