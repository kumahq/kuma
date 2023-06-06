package v1alpha1

import (
	"context"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
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

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshHealthCheckType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshHealthCheckType]
	if !ok {
		return nil
	}

	clusters := policies_xds.GatherClusters(rs)

	if err := applyToOutbounds(policies.ToRules, clusters.Outbound, clusters.OutboundSplit, proxy.Dataplane, proxy.Routing); err != nil {
		return err
	}

	if err := applyToGateways(ctx, policies.ToRules, clusters.Gateway, proxy); err != nil {
		return err
	}

	return nil
}

func applyToOutbounds(rules core_rules.ToRules, outboundClusters map[string]*envoy_cluster.Cluster, outboundSplitClusters map[string][]*envoy_cluster.Cluster, dataplane *core_mesh.DataplaneResource, routing core_xds.Routing) error {
	targetedClusters := policies_xds.GatherTargetedClusters(dataplane.Spec.Networking.GetOutbound(), outboundSplitClusters, outboundClusters)

	for cluster, serviceName := range targetedClusters {
		protocol := policies_xds.InferProtocol(routing, serviceName)

		if err := configure(dataplane, rules.Rules, core_rules.MeshService(serviceName), protocol, cluster); err != nil {
			return err
		}
	}

	return nil
}

func applyToGateways(
	ctx xds_context.Context,
	rules core_rules.ToRules,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	proxy *core_xds.Proxy,
) error {
	if !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}
	gatewayListenerInfos, err := gateway_plugin.GatewayListenerInfoFromProxy(context.TODO(), ctx.Mesh, proxy, ctx.ControlPlane.Zone)
	if err != nil {
		return err
	}

	for _, listenerInfo := range gatewayListenerInfos {
		for _, hostInfo := range listenerInfo.HostInfos {
			destinations := gateway_plugin.RouteDestinationsMutable(hostInfo.Entries)
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
					rules.Rules,
					core_rules.MeshService(serviceName),
					toProtocol(listenerInfo.Listener.Protocol),
					cluster,
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func toProtocol(p mesh_proto.MeshGateway_Listener_Protocol) core_mesh.Protocol {
	return core_mesh.ParseProtocol(p.String())
}

func configure(
	dataplane *core_mesh.DataplaneResource,
	rules core_rules.Rules,
	subset core_rules.Subset,
	protocol core_mesh.Protocol,
	cluster *envoy_cluster.Cluster,
) error {
	var conf api.Conf
	if computed := rules.Compute(subset); computed != nil {
		conf = computed.Conf.(api.Conf)
	} else {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Conf:     conf,
		Protocol: protocol,
		Tags:     dataplane.Spec.TagSet(),
	}

	if err := configurer.Configure(cluster); err != nil {
		return err
	}
	return nil
}
