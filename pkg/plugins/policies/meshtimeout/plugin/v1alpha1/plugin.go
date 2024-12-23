package v1alpha1

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
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

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshTimeoutType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshTimeoutType]
	if !ok {
		return nil
	}
	if len(policies.ToRules.Rules) == 0 && len(policies.FromRules.Rules) == 0 && len(policies.GatewayRules.ToRules.ByListener) == 0 && len(policies.ToRules.ResourceRules) == 0 {
		return nil
	}

	listeners := xds.GatherListeners(rs)
	clusters := xds.GatherClusters(rs)
	routes := xds.GatherRoutes(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Outbounds, proxy.Dataplane, ctx.Mesh); err != nil {
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

	err := applyToRealResources(rs, policies.ToRules.ResourceRules, ctx.Mesh)
	if err != nil {
		return err
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
		conf := getConf(fromRules.Rules[listenerKey], core_rules.MeshElement())
		if conf == nil {
			continue
		}
		configurer := plugin_xds.ListenerConfigurer{
			Conf:     *conf,
			Protocol: protocol,
		}

		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}

		cluster, ok := inboundClusters[createInboundClusterName(inbound.ServicePort, listenerKey.Port)]
		if !ok {
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
	outbounds xds_types.Outbounds,
	dataplane *core_mesh.DataplaneResource,
	meshCtx xds_context.MeshContext,
) error {
	for _, outbound := range outbounds.Filter(xds_types.NonBackendRefFilter) {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound.LegacyOutbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.LegacyOutbound.GetService()
		configurer := plugin_xds.DeprecatedListenerConfigurer{
			Rules:    rules.Rules,
			Protocol: meshCtx.GetServiceProtocol(serviceName),
			Subset:   core_rules.MeshService(serviceName),
			Element:  core_rules.MeshServiceElement(serviceName),
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
	conf := getConf(rules, core_rules.MeshServiceElement(serviceName))
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

		conf := getConf(gatewayRules.FromRules[key], core_rules.MeshElement())
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

		conf = getConf(toRules.Rules, core_rules.MeshElement())
		for _, listenerHostname := range listenerInfo.ListenerHostnames {
			route, ok := gatewayRoutes[listenerHostname.EnvoyRouteName(listenerInfo.Listener.EnvoyListenerName)]

			if ok {
				for _, vh := range route.VirtualHosts {
					for _, r := range vh.Routes {
						routeConf := getConf(toRules.Rules, core_rules.MeshElement().WithKeyValue(core_rules.RuleMatchesHashTag, r.Name))
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

					conf := getConf(toRules.Rules, core_rules.MeshServiceElement(serviceName))
					if conf == nil {
						continue
					}

					if err := applyToClusters(
						toRules.Rules,
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
	element core_rules.Element,
) *api.Conf {
	if rules == nil {
		return &api.Conf{}
	} else {
		if computed := rules.Compute(element); computed != nil {
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

func applyToRealResources(rs *core_xds.ResourceSet, rules core_rules.ResourceRules, meshCtx xds_context.MeshContext) error {
	for uri, resType := range rs.IndexByOrigin() {
		conf := rules.Compute(uri, meshCtx.Resources)
		if conf == nil {
			conf = &core_rules.ResourceRule{Conf: []interface{}{plugin_xds.DefaultTimeoutConf}}
		}

		for typ, resources := range resType {
			switch typ {
			case envoy_resource.ListenerType:
				err := configureListeners(resources, conf.Conf[0].(api.Conf))
				if err != nil {
					return err
				}
			case envoy_resource.ClusterType:
				err := configureClusters(resources, conf.Conf[0].(api.Conf))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func configureListeners(resources []*core_xds.Resource, conf api.Conf) error {
	for _, resource := range resources {
		configurer := plugin_xds.ListenerConfigurer{Conf: conf, Protocol: resource.Protocol}
		err := configurer.ConfigureListener(resource.Resource.(*envoy_listener.Listener))
		if err != nil {
			return err
		}
	}
	return nil
}

func configureClusters(resources []*core_xds.Resource, conf api.Conf) error {
	for _, resource := range resources {
		configurer := plugin_xds.ClusterConfigurerFromConf(conf, resource.Protocol)
		err := configurer.Configure(resource.Resource.(*envoy_cluster.Cluster))
		if err != nil {
			return err
		}
	}
	return nil
}
