package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/generator"
)

var _ core_plugins.PolicyPlugin = &plugin{}
var log = core.Log.WithName("MeshHealthCheck")

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHealthCheckType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshHealthCheckType]
	if !ok {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)

	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, clusters.Outbound, proxy.Dataplane, proxy.Routing); err != nil {
		return err
	}

	log.Info("apply is not implemented")
	return nil
}

func applyToOutbounds(rules core_xds.ToRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, outboundClusters map[string]*envoy_cluster.Cluster, dataplane *core_mesh.DataplaneResource, routing core_xds.Routing) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
		cluster, ok := outboundClusters[serviceName]
		if !ok {
			continue
		}

		protocol := inferProtocol(routing, serviceName)
		if err := configure(rules.Rules, core_xds.MeshService(serviceName), protocol, listener, cluster, nil); err != nil {
			return err
		}
	}

	return nil
}

func inferProtocol(routing core_xds.Routing, serviceName string) core_mesh.Protocol {
	var allEndpoints []core_xds.Endpoint
	outboundEndpoints := core_xds.EndpointList(routing.OutboundTargets[serviceName])
	allEndpoints = append(allEndpoints, outboundEndpoints...)
	externalEndpoints := routing.ExternalServiceOutboundTargets[serviceName]
	allEndpoints = append(allEndpoints, externalEndpoints...)

	return envoy_common.InferServiceProtocol(allEndpoints)
}

func configure(rules core_xds.Rules, subset core_xds.Subset, protocol core_mesh.Protocol, listener *envoy_listener.Listener, cluster *envoy_cluster.Cluster, routeActions []*envoy_route.RouteAction) error {
	var conf api.Conf
	if computed := rules.Compute(subset); computed != nil {
		conf = computed.Conf.(api.Conf)
	} else {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Conf:     conf,
		Protocol: protocol,
	}

	if err := configurer.Configure(cluster); err != nil {
		return err
	}
	return nil
}