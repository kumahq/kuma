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
	"github.com/kumahq/kuma/pkg/plugins/policies/utils"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/clusters"
)

const MeshTraceOrigin = "meshTrace"

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

	listeners := utils.GatherListeners(rs)
	if err := applyToInbounds(policies.SingleItemRules, listeners.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.SingleItemRules, listeners.Outbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToClusters(policies.SingleItemRules, rs, proxy); err != nil {
		return err
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
	conf := rawConf.(*api.MeshTrace_Conf)

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

	conf := rawConf.(*api.MeshTrace_Conf)

	backend := conf.GetBackends()[0]
	if backend == nil {
		return nil
	}

	var endpoint *xds.Endpoint
	var provider string

	if backend.GetZipkin() != nil {
		endpoint = endpointForZipkin(backend.GetZipkin())
		provider = "zipkin"
	} else {
		endpoint = endpointForDatadog(backend.GetDatadog())
		provider = "datadog"
	}

	res, err := clusters.NewClusterBuilder(proxy.APIVersion).
		Configure(clusters.ProvidedEndpointCluster(plugin_xds.GetTracingClusterName(provider), proxy.Dataplane.IsIPv6(), *endpoint)).
		Configure(clusters.ClientSideTLS([]xds.Endpoint{*endpoint})).
		Configure(clusters.DefaultTimeout()).
		Build()
	if err != nil {
		return err
	}

	rs.Add(&xds.Resource{Name: plugin_xds.GetTracingClusterName(provider), Origin: MeshTraceOrigin, Resource: res})

	return nil
}

func endpointForZipkin(cfg *api.MeshTrace_ZipkinBackend) *xds.Endpoint {
	url, _ := net_url.ParseRequestURI(cfg.Url)
	port, _ := strconv.Atoi(url.Port())
	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
		ExternalService: &xds.ExternalService{
			TLSEnabled:         url.Scheme == "https",
			AllowRenegotiation: true,
		},
	}
}

func endpointForDatadog(cfg *api.MeshTrace_DatadogBackend) *xds.Endpoint {
	url, _ := net_url.ParseRequestURI(cfg.Url)
	port, _ := strconv.Atoi(url.Port())

	return &xds.Endpoint{
		Target: url.Hostname(),
		Port:   uint32(port),
	}
}
