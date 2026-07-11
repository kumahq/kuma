package v1alpha1

import (
	"slices"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/merge"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/v3/pkg/xds/generator/gateway"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/v3/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func (p plugin) Order() int { return api.MeshTimeoutResourceTypeDescriptor.Order }

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

	listeners := xds.GatherListeners(rs)
	clusters := xds.GatherClusters(rs)
	routes := xds.GatherRoutes(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, clusters.Inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToZoneProxyListeners(policies, listeners, clusters, proxy); err != nil {
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

	rctx := outbound.RootContext[api.Conf](ctx.Mesh.Resource, policies.ToRules.ResourceRules)

	for _, r := range util_slices.Filter(rs.List(), core_xds.HasAssociatedServiceResource) {
		svcCtx := rctx.
			WithID(kri.NoSectionName(r.ResourceOrigin)).
			WithID(r.ResourceOrigin)
		if err := applyToRealResource(svcCtx, r); err != nil {
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

		protocol := core_meta.ParseProtocol(inbound.GetProtocolFallback())

		inboundRules := fromRules.InboundRules[listenerKey]
		conf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](inboundRules)
		applyCommonConf := len(inboundRules) == 0 || hasCatchAllInboundRule(inboundRules)
		configurer := plugin_xds.ListenerConfigurer{
			Conf:             conf,
			Rules:            inboundRules,
			SkipCommonConfig: !applyCommonConf,
		}

		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}

		cluster, ok := inboundClusters[envoy_names.GetInboundClusterName(inbound.ServicePort, listenerKey.Port)]
		if !ok {
			continue
		}

		if applyCommonConf {
			clusterConfigurer := plugin_xds.ClusterConfigurerFromConf(conf, protocol)
			if err := clusterConfigurer.Configure(cluster); err != nil {
				return err
			}
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
			Element:  subsetutils.KumaServiceTagElement(serviceName),
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
	protocol core_meta.Protocol,
	clusters ...*envoy_cluster.Cluster,
) error {
	conf, _ := getConf(rules, subsetutils.KumaServiceTagElement(serviceName))
	configurer := plugin_xds.ClusterConfigurerFromConf(conf, protocol)
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

		inboundRules := gatewayRules.InboundRules[key]
		inboundConf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](inboundRules)
		if len(inboundRules) == 0 || hasCatchAllInboundRule(inboundRules) {
			if err := plugin_xds.ConfigureGatewayListener(
				inboundConf,
				listenerInfo.Listener.Protocol,
				gatewayListeners[key],
			); err != nil {
				return err
			}
		}
		if err := plugin_xds.EnsureMatchFilterState(gatewayListeners[key], inboundRules); err != nil {
			return err
		}

		toRules, hasToRules := gatewayRules.ToRules.ByListener[key]
		conf, commonOk := getConf(toRules.Rules, subsetutils.MeshElement())
		for _, listenerHostname := range listenerInfo.ListenerHostnames {
			route, ok := gatewayRoutes[listenerHostname.EnvoyRouteName(listenerInfo.Listener.EnvoyListenerName)]
			if ok && hasToRules {
				for _, vh := range route.VirtualHosts {
					for _, r := range vh.Routes {
						routeConf, routeOk := getConf(toRules.Rules, subsetutils.MeshElement().WithKeyValue(core_rules.RuleMatchesHashTag, r.Name))
						if !routeOk {
							if !commonOk {
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
			if ok {
				if err := plugin_xds.ConfigureMatchedRoutes(route, inboundRules); err != nil {
					return err
				}
			}
		}

		if !hasToRules {
			continue
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

func applyToZoneProxyListeners(
	policies core_xds.TypedMatchingPolicies,
	listeners xds.Listeners,
	clusters xds.Clusters,
	proxy *core_xds.Proxy,
) error {
	networking := proxy.Dataplane.Spec.GetNetworking()
	if !networking.HasZoneProxyListeners() {
		return nil
	}
	filterChainTargets := xds.GatherFilterChainTargets(listeners, clusters)

	for _, listener := range networking.GetListeners() {
		var (
			envoyListener *envoy_listener.Listener
			ok            bool
		)

		switch listener.GetType() {
		case mesh_proto.Dataplane_Networking_Listener_ZoneIngress:
			envoyListener, ok = listeners.ZoneIngress[naming.ContextualZoneIngressListenerName(listener.GetSectionName())]
		case mesh_proto.Dataplane_Networking_Listener_ZoneEgress:
			envoyListener, ok = listeners.ZoneEgress[naming.ContextualZoneEgressListenerName(listener.GetSectionName())]
		default:
			continue
		}
		if !ok {
			continue
		}

		inboundRules, err := buildListenerScopedInboundRules(policies, listener.GetSectionName())
		if err != nil {
			return err
		}
		if len(inboundRules) == 0 {
			continue
		}

		if err := applyToZoneProxyListener(envoyListener, filterChainTargets, inboundRules); err != nil {
			return err
		}
	}

	return nil
}

func applyToZoneProxyListener(
	listener *envoy_listener.Listener,
	filterChainTargets map[*envoy_listener.FilterChain]*envoy_cluster.Cluster,
	inboundRules []*rules_inbound.Rule,
) error {
	commonConf := rules_inbound.MatchesAllIncomingTraffic[api.Conf](inboundRules)
	applyCommonConf := hasCatchAllInboundRule(inboundRules)
	for _, filterChain := range listener.FilterChains {
		cluster := filterChainTargets[filterChain]
		protocol := xds.FilterChainProtocol(filterChain)
		if applyCommonConf {
			if err := plugin_xds.ConfigureFilterChain(commonConf, filterChain); err != nil {
				return err
			}
			if err := applyZoneProxyClusterConf(cluster, protocol, commonConf); err != nil {
				return err
			}
		}

		matchedRules := zoneProxyFilterChainRules(inboundRules, filterChain)
		if conf, ok, err := mergeZoneProxyRuleConfs(matchedRules); err != nil {
			return err
		} else if ok {
			if err := plugin_xds.ConfigureFilterChain(conf, filterChain); err != nil {
				return err
			}
			if err := applyZoneProxyClusterConf(cluster, protocol, conf); err != nil {
				return err
			}
		}

		if err := plugin_xds.ConfigureMatchedRoutesOnFilterChain(filterChain, inboundRules); err != nil {
			return err
		}
	}

	return nil
}

func applyZoneProxyClusterConf(
	cluster *envoy_cluster.Cluster,
	protocol core_meta.Protocol,
	conf api.Conf,
) error {
	if cluster == nil {
		return nil
	}
	configurer := plugin_xds.ClusterConfigurerFromConf(conf, protocol)
	return configurer.Configure(cluster)
}

func matchesZoneProxyFilterChain(rule *rules_inbound.Rule, filterChain *envoy_listener.FilterChain) bool {
	if rule.Match == nil {
		return false
	}
	serverNames := filterChain.GetFilterChainMatch().GetServerNames()
	if len(serverNames) == 0 {
		return false
	}
	if rule.Match.SpiffeID != nil || rule.Match.SNI == nil {
		return false
	}
	return containsString(serverNames, rule.Match.SNI.Value)
}

func zoneProxyFilterChainRules(inboundRules []*rules_inbound.Rule, filterChain *envoy_listener.FilterChain) []*rules_inbound.Rule {
	var matched []*rules_inbound.Rule
	for _, rule := range inboundRules {
		if matchesZoneProxyFilterChain(rule, filterChain) {
			matched = append(matched, rule)
		}
	}
	return matched
}

func mergeZoneProxyRuleConfs(rules []*rules_inbound.Rule) (api.Conf, bool, error) {
	confs := make([]any, 0, len(rules))
	for _, rule := range rules {
		conf, ok := rule.Conf.(api.Conf)
		if !ok {
			continue
		}
		confs = append(confs, conf)
	}
	if len(confs) == 0 {
		return api.Conf{}, false, nil
	}

	merged, err := merge.Confs(confs)
	if err != nil {
		return api.Conf{}, false, err
	}
	if len(merged) == 0 {
		return api.Conf{}, false, nil
	}

	conf, ok := merged[0].(api.Conf)
	if !ok {
		return api.Conf{}, false, errors.Errorf("unexpected merged zone proxy conf type: %T", merged[0])
	}
	return conf, true, nil
}

func hasCatchAllInboundRule(rules []*rules_inbound.Rule) bool {
	for _, rule := range rules {
		if rule.Match == nil {
			return true
		}
	}
	return false
}

func containsString(values []string, value string) bool {
	return slices.Contains(values, value)
}

func buildListenerScopedInboundRules(
	policies core_xds.TypedMatchingPolicies,
	sectionName string,
) ([]*rules_inbound.Rule, error) {
	if len(policies.DataplanePolicies) == 0 {
		return nil, nil
	}

	filtered := api.MeshTimeoutResourceTypeDescriptor.NewList()
	for _, resource := range policies.DataplanePolicies {
		policy, ok := resource.GetSpec().(core_model.Policy)
		if !ok {
			continue
		}
		targetRef := policy.GetTargetRef()
		if targetRef.Kind == common_api.Dataplane {
			if sn := pointer.Deref(targetRef.SectionName); sn != "" && sn != sectionName {
				continue
			}
		}
		if err := filtered.AddItem(resource); err != nil {
			return nil, err
		}
	}

	return rules_inbound.BuildRules(filtered)
}

func getConf(
	rules core_rules.Rules,
	element subsetutils.Element,
) (api.Conf, bool) {
	if computed := rules.Compute(element); computed != nil {
		return computed.Conf.(api.Conf), true
	}
	return api.Conf{}, false
}

func applyToRealResource(rctx *outbound.ResourceContext[api.Conf], r *core_xds.Resource) error {
	switch envoyResource := r.Resource.(type) {
	case *envoy_listener.Listener:
		configurer := plugin_xds.ListenerConfigurer{Conf: rctx.Conf()}
		if err := configurer.ConfigureListener(envoyResource); err != nil {
			return err
		}

		for _, fc := range envoyResource.FilterChains {
			if err := listeners_v3.UpdateHTTPConnectionManager(fc, func(hcm *envoy_hcm.HttpConnectionManager) error {
				for _, vh := range hcm.GetRouteConfig().VirtualHosts {
					for _, route := range vh.Routes {
						if !kri.IsValid(route.Name) {
							continue
						}

						id, err := kri.FromString(route.Name)
						if err != nil {
							return err
						}

						routeCtx := rctx.
							WithID(kri.NoSectionName(id)).
							WithID(id)

						plugin_xds.ConfigureRouteAction(
							route.GetRoute(),
							pointer.Deref(routeCtx.Conf().Http).RequestTimeout,
							pointer.Deref(routeCtx.Conf().Http).StreamIdleTimeout,
						)
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}

	case *envoy_cluster.Cluster:
		configurer := plugin_xds.ClusterConfigurerFromConf(rctx.Conf(), r.Protocol)
		if err := configurer.Configure(envoyResource); err != nil {
			return err
		}
	}
	return nil
}
