package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHealthCheckType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshHealthCheckType]
	if !ok {
		return nil
	}

	clusters := policies_xds.GatherClusters(rs)

	if err := applyToOutbounds(policies.ToRules, clusters.Outbound, clusters.OutboundSplit, proxy.Dataplane, ctx.Mesh); err != nil {
		return err
	}

	if err := applyToGateways(policies.GatewayRules, clusters.Gateway, proxy); err != nil {
		return err
	}

	return nil
}

func applyToOutbounds(
	rules core_rules.ToRules,
	outboundClusters map[string]*envoy_cluster.Cluster,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	dataplane *core_mesh.DataplaneResource,
	meshCtx xds_context.MeshContext,
) error {
	targetedClusters := policies_xds.GatherTargetedClusters(
		dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter),
		outboundSplitClusters,
		outboundClusters,
	)

	for cluster, serviceName := range targetedClusters {
		if err := configure(dataplane, rules.Rules, core_rules.MeshServiceElement(serviceName), meshCtx.GetServiceProtocol(serviceName), cluster); err != nil {
			return err
		}
	}

	return nil
}

func applyToGateways(
	gatewayRules core_rules.GatewayRules,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	proxy *core_xds.Proxy,
) error {
	for _, listenerInfo := range gateway_plugin.ExtractGatewayListeners(proxy) {
		for _, listenerHostname := range listenerInfo.ListenerHostnames {
			inboundListener := rules.NewInboundListenerHostname(
				proxy.Dataplane.Spec.GetNetworking().Address,
				listenerInfo.Listener.Port,
				listenerHostname.Hostname,
			)
			rules, ok := gatewayRules.ToRules.ByListenerAndHostname[inboundListener]
			if !ok {
				continue
			}
			for _, hostInfo := range listenerHostname.HostInfos {
				destinations := gateway_plugin.RouteDestinationsMutable(hostInfo.Entries())
				for _, dest := range destinations {
					clusterName, err := dest.Destination.DestinationClusterName(hostInfo.Host.Tags)
					if err != nil {
						continue
					}
					cluster, ok := gatewayClusters[clusterName]
					if !ok {
						continue
					}

					serviceName := dest.Destination[mesh_proto.ServiceTag]

					if err := configure(
						proxy.Dataplane,
						rules,
						core_rules.MeshServiceElement(serviceName),
						toProtocol(listenerInfo.Listener.Protocol),
						cluster,
					); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func toProtocol(p mesh_proto.MeshGateway_Listener_Protocol) core_mesh.Protocol {
	switch p {
	case mesh_proto.MeshGateway_Listener_HTTP, mesh_proto.MeshGateway_Listener_HTTPS:
		return core_mesh.ProtocolHTTP
	case mesh_proto.MeshGateway_Listener_TCP, mesh_proto.MeshGateway_Listener_TLS:
		return core_mesh.ProtocolTCP
	}
	return core_mesh.ProtocolTCP
}

func configure(
	dataplane *core_mesh.DataplaneResource,
	rules core_rules.Rules,
	element core_rules.Element,
	protocol core_mesh.Protocol,
	cluster *envoy_cluster.Cluster,
) error {
	conf := core_rules.ComputeConf[api.Conf](rules, element)
	if conf == nil {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Conf:     *conf,
		Protocol: protocol,
		Tags:     dataplane.Spec.TagSet(),
	}

	if err := configurer.Configure(cluster); err != nil {
		return err
	}
	return nil
}
