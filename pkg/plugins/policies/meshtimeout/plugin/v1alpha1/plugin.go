package v1alpha1

import (
	"context"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTimeoutType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshTimeoutType]
	if !ok {
		return nil
	}
	if len(policies.ToRules.Rules) == 0 && len(policies.FromRules.Rules) == 0 {
		return nil
	}

	listeners := policies_xds.GatherListeners(rs)
	clusters := policies_xds.GatherClusters(rs)
	routes := policies_xds.GatherRoutes(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, clusters.Outbound, clusters.OutboundSplit, proxy.Dataplane, proxy.Routing); err != nil {
		return err
	}
	if err := applyToGateway(policies.ToRules, clusters.Gateway, routes.Gateway, ctx, proxy); err != nil {
		return err
	}
	return nil
}

func applyToInbounds(fromRules core_xds.FromRules, inboundListeners map[core_xds.InboundListener]*envoy_listener.Listener, inboundClusters map[string]*envoy_cluster.Cluster, dataplane *core_mesh.DataplaneResource) error {
	for _, inbound := range dataplane.Spec.Networking.GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_xds.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}

		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		rules := fromRules.Rules[listenerKey]

		cluster, ok := inboundClusters[createInboundClusterName(inbound.ServicePort, listenerKey.Port)]
		if !ok {
			continue
		}
		protocol := core_mesh.ParseProtocol(inbound.GetProtocol())
		if err := configure(rules, core_xds.MeshSubset(), protocol, listener, cluster, nil); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(
	rules core_xds.ToRules,
	outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener,
	outboundClusters map[string]*envoy_cluster.Cluster,
	outboundSplitClusters map[string][]*envoy_cluster.Cluster,
	dataplane *core_mesh.DataplaneResource,
	routing core_xds.Routing,
) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		clustersToConfigure := []*envoy_cluster.Cluster{}
		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
		splitCluster, ok := outboundSplitClusters[serviceName]
		if ok {
			clustersToConfigure = append(clustersToConfigure, splitCluster...)
		}
		cluster, ok := outboundClusters[serviceName]
		if ok {
			clustersToConfigure = append(clustersToConfigure, cluster)
		}

		if len(clustersToConfigure) == 0 {
			continue
		}

		protocol := policies_xds.InferProtocol(routing, serviceName)
		for _, cluster := range clustersToConfigure {
			if err := configure(rules.Rules, core_xds.MeshService(serviceName), protocol, listener, cluster, nil); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyToGateway(
	toRules core_xds.ToRules,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	ctx xds_context.Context,
	proxy *core_xds.Proxy,
) error {
	gatewayListerInfos, err := gateway_plugin.GatewayListenerInfoFromProxy(context.TODO(), ctx.Mesh, proxy, ctx.ControlPlane.Zone)
	if err != nil {
		return err
	}
	for _, listenerInfo := range gatewayListerInfos {
		route, ok := gatewayRoutes[listenerInfo.Listener.ResourceName]
		if !ok {
			continue
		}
		routeActionsPerCluster := routeActionPerCluster(route)

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

				routeActions, ok := routeActionsPerCluster[clusterName]
				if !ok {
					continue
				}

				serviceName := dest.Destination[mesh_proto.ServiceTag]

				if err := configure(
					toRules.Rules,
					core_xds.MeshService(serviceName),
					toProtocol(listenerInfo.Listener.Protocol),
					nil,
					cluster,
					routeActions,
				); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func configure(rules core_xds.Rules, subset core_xds.Subset, protocol core_mesh.Protocol, listener *envoy_listener.Listener, cluster *envoy_cluster.Cluster, routeActions []*envoy_route.RouteAction) error {
	var conf api.Conf
	if rules == nil {
		conf = api.Conf{}
	} else {
		if computed := rules.Compute(subset); computed != nil {
			conf = computed.Conf.(api.Conf)
		} else {
			return nil
		}
	}

	configurer := plugin_xds.Configurer{
		Conf:     conf,
		Protocol: protocol,
	}

	if err := configurer.ConfigureListener(listener); err != nil {
		return err
	}

	for _, routeAction := range routeActions {
		configurer.ConfigureRouteAction(routeAction)
	}

	if err := configurer.ConfigureCluster(cluster); err != nil {
		return err
	}
	return nil
}

func toProtocol(p mesh_proto.MeshGateway_Listener_Protocol) core_mesh.Protocol {
	return core_mesh.ParseProtocol(mesh_proto.MeshGateway_Listener_Protocol_name[int32(p.Number())])
}

func routeActionPerCluster(route *envoy_route.RouteConfiguration) map[string][]*envoy_route.RouteAction {
	actions := map[string][]*envoy_route.RouteAction{}
	for _, vh := range route.VirtualHosts {
		for _, r := range vh.Routes {
			routeAction := r.GetRoute()
			if routeAction == nil {
				continue
			}
			cluster := routeAction.GetWeightedClusters().GetClusters()[0].Name
			if actions[cluster] == nil {
				actions[cluster] = []*envoy_route.RouteAction{routeAction}
			} else {
				actions[cluster] = append(actions[cluster], routeAction)
			}
		}
	}
	return actions
}

func createInboundClusterName(servicePort uint32, listenerPort uint32) string {
	if servicePort != 0 {
		return envoy_names.GetLocalClusterName(servicePort)
	} else {
		return envoy_names.GetLocalClusterName(listenerPort)
	}
}
