package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/util/pointer"
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
	if len(policies.ToRules.Rules) == 0 && len(policies.FromRules.Rules) == 0 && len(policies.GatewayRules.ToRules.ByListener) == 0 {
		return nil
	}

	listeners := xds.GatherListeners(rs)
	clusters := xds.GatherClusters(rs)
	routes := xds.GatherRoutes(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Dataplane, ctx.Mesh); err != nil {
		return err
	}
	if err := applyToGateway(policies.GatewayRules, listeners.Gateway, clusters.Gateway, routes.Gateway, proxy, ctx.Mesh); err != nil {
		return err
	}

	for serviceName, cluster := range clusters.Outbound {
		if err := applyToClusters(policies.ToRules.Rules, serviceName, ctx.Mesh.GetServiceProtocol(serviceName), cluster); err != nil {
			return err
		}
	}
	for serviceName, clusters := range clusters.OutboundSplit {
		if err := applyToClusters(policies.ToRules.Rules, serviceName, ctx.Mesh.GetServiceProtocol(serviceName), clusters...); err != nil {
			return err
		}
	}
	return nil
}

func applyToInbounds(fromRules core_rules.FromRules, inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener, inboundClusters map[string]*envoy_cluster.Cluster, dataplane *core_mesh.DataplaneResource) error {
	for _, inbound := range dataplane.Spec.Networking.GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}

		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		protocol := core_mesh.ParseProtocol(inbound.GetProtocol())
		configurer := plugin_xds.ListenerConfigurer{
			Rules:    fromRules.Rules[listenerKey],
			Subset:   core_rules.MeshSubset(),
			Protocol: protocol,
		}

		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}

		cluster, ok := inboundClusters[createInboundClusterName(inbound.ServicePort, listenerKey.Port)]
		if !ok {
			continue
		}

		conf := getConf(fromRules.Rules[listenerKey], core_rules.MeshSubset())
		if conf == nil {
			continue
		}

		clusterConfigurer := plugin_xds.ClusterConfigurerFromConf(*conf, protocol)
		if err := clusterConfigurer.Configure(cluster); err != nil {
			return err
		}
	}

	return nil
}

func applyToOutbounds(
	rules core_rules.ToRules,
	outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener,
	dataplane *core_mesh.DataplaneResource,
	meshCtx xds_context.MeshContext,
) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter) {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetService()
		configurer := plugin_xds.ListenerConfigurer{
			Rules:    rules.Rules,
			Protocol: meshCtx.GetServiceProtocol(serviceName),
			Subset:   core_rules.MeshService(serviceName),
		}

		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}
	}

	return nil
}

func applyToClusters(
	rules core_rules.Rules,
	serviceName string,
	protocol core_mesh.Protocol,
	clusters ...*envoy_cluster.Cluster,
) error {
	conf := getConf(rules, core_rules.MeshService(serviceName))
	if conf == nil {
		return nil
	}

	configurer := plugin_xds.ClusterConfigurerFromConf(*conf, protocol)
	for _, cluster := range clusters {
		if err := configurer.Configure(cluster); err != nil {
			return err
		}
	}
	return nil
}

func applyToGateway(
	gatewayRules core_rules.GatewayRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	gatewayClusters map[string]*envoy_cluster.Cluster,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	proxy *core_xds.Proxy,
	meshCtx xds_context.MeshContext,
) error {
	for _, listenerInfo := range gateway_plugin.ExtractGatewayListeners(proxy) {
		key := core_rules.InboundListener{
			Address: proxy.Dataplane.Spec.GetNetworking().Address,
			Port:    listenerInfo.Listener.Port,
		}

		conf := getConf(gatewayRules.FromRules[key], core_rules.MeshSubset())
		if err := plugin_xds.ConfigureGatewayListener(
			conf,
			listenerInfo.Listener.Protocol,
			gatewayListeners[key],
		); err != nil {
			return err
		}

		toRules, ok := gatewayRules.ToRules.ByListener[key]
		if !ok {
			continue
		}

		conf = getConf(toRules, core_rules.MeshSubset())
		for _, listenerHostname := range listenerInfo.ListenerHostnames {
			route, ok := gatewayRoutes[listenerHostname.EnvoyRouteName(listenerInfo.Listener.EnvoyListenerName)]

			if ok {
				for _, vh := range route.VirtualHosts {
					for _, r := range vh.Routes {
						routeConf := getConf(toRules, core_rules.MeshSubset().WithTag(core_rules.RuleMatchesHashTag, r.Name, false))
						if routeConf == nil {
							if conf == nil {
								continue
							}
							// use the common configuration for all routes
							routeConf = conf
						}
						plugin_xds.ConfigureRouteAction(
							r.GetRoute(),
							pointer.Deref(routeConf.Http).RequestTimeout,
							pointer.Deref(routeConf.Http).StreamIdleTimeout,
						)
					}
				}
			}
		}

		for _, listenerHostnames := range listenerInfo.ListenerHostnames {
			for _, hostInfo := range listenerHostnames.HostInfos {
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

					conf := getConf(toRules, core_rules.MeshService(serviceName))
					if conf == nil {
						continue
					}

					if err := applyToClusters(
						toRules,
						serviceName,
						meshCtx.GetServiceProtocol(serviceName),
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

func getConf(
	rules core_rules.Rules,
	subset core_rules.Subset,
) *api.Conf {
	if rules == nil {
		return &api.Conf{}
	} else {
		if computed := rules.Compute(subset); computed != nil {
			return pointer.To(computed.Conf.(api.Conf))
		} else {
			return nil
		}
	}
}

func createInboundClusterName(servicePort uint32, listenerPort uint32) string {
	if servicePort != 0 {
		return envoy_names.GetLocalClusterName(servicePort)
	} else {
		return envoy_names.GetLocalClusterName(listenerPort)
	}
}
