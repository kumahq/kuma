package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var (
	_   core_plugins.EgressPolicyPlugin = &plugin{}
	log                                 = core.Log.WithName("MeshRateLimit")
)

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshRateLimitType, dataplane, resources, opts...)
}

func (p plugin) EgressMatchedPolicies(tags map[string]string, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshRateLimitType, tags, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.ZoneEgressProxy != nil {
		return applyToEgress(rs, proxy)
	}
	if proxy.Dataplane == nil {
		return nil
	}
	policies, ok := proxy.Policies.Dynamic[api.MeshRateLimitType]
	if !ok {
		return nil
	}
	listeners := xds.GatherListeners(rs)
	routes := xds.GatherRoutes(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, proxy); err != nil {
		return err
	}
	if err := applyToGateways(policies.GatewayRules, listeners.Gateway, routes.Gateway, proxy); err != nil {
		return err
	}
	return nil
}

func applyToGateways(
	toRules core_rules.GatewayRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	proxy *core_xds.Proxy,
) error {
	for _, listenerInfo := range gateway_plugin.ExtractGatewayListeners(proxy) {
		address := proxy.Dataplane.Spec.GetNetworking().Address
		port := listenerInfo.Listener.Port
		listenerKey := core_rules.InboundListener{
			Address: address,
			Port:    port,
		}
		gatewayListener, ok := gatewayListeners[listenerKey]
		if !ok {
			continue
		}
		rules, ok := toRules.ToRules.ByListener[listenerKey]
		if !ok {
			continue
		}

		for _, listenerHostname := range listenerInfo.ListenerHostnames {
			route, ok := gatewayRoutes[listenerHostname.EnvoyRouteName(listenerInfo.Listener.EnvoyListenerName)]
			if !ok {
				continue
			}

			if err := configureGateway(rules.Rules, gatewayListener, route); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyToInbounds(
	fromRules core_rules.FromRules,
	inboundListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	proxy *core_xds.Proxy,
) error {
	for _, inbound := range proxy.Dataplane.Spec.GetNetworking().GetInbound() {
		iface := proxy.Dataplane.Spec.Networking.ToInboundInterface(inbound)
		if _, exists := proxy.Policies.RateLimitsInbound[iface]; exists {
			continue
		}

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}
		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}
		rules, ok := fromRules.Rules[listenerKey]
		if !ok {
			continue
		}

		if err := configure(rules, listener, nil); err != nil {
			return err
		}
	}
	return nil
}

func applyToEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	listeners := xds.GatherListeners(rs)
	if listeners.Egress == nil {
		log.V(1).Info("skip applying MeshRateLimit, Egress has no listener",
			"proxyName", proxy.ZoneEgressProxy.ZoneEgressResource.GetMeta().GetName(),
		)
		return nil
	}
	for _, resource := range proxy.ZoneEgressProxy.MeshResourcesList {
		for _, es := range resource.ExternalServices {
			meshName := resource.Mesh.GetMeta().GetName()
			esName, ok := es.Spec.GetTags()[mesh_proto.ServiceTag]
			if !ok {
				continue
			}
			policies, ok := resource.Dynamic[esName]
			if !ok {
				continue
			}
			mrl, ok := policies[api.MeshRateLimitType]
			if !ok {
				continue
			}
			for _, rule := range mrl.FromRules.Rules {
				for _, filterChain := range listeners.Egress.FilterChains {
					if filterChain.Name == names.GetEgressFilterChainName(esName, meshName) {
						configurer := plugin_xds.Configurer{
							Rules:  rule,
							Subset: core_rules.MeshSubset(),
						}
						if err := configurer.ConfigureFilterChain(filterChain); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

func configure(
	fromRules core_rules.Rules,
	listener *envoy_listener.Listener,
	route *envoy_route.RouteConfiguration,
) error {
	configurer := plugin_xds.Configurer{
		Rules: fromRules,
		// Currently, `from` section of MeshRateLimit only allows Mesh targetRef
		Subset: core_rules.MeshSubset(),
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.ConfigureFilterChain(chain); err != nil {
			return err
		}
	}
	if err := configurer.ConfigureRoute(route); err != nil {
		return err
	}

	return nil
}

func configureGateway(
	fromRules core_rules.Rules,
	listener *envoy_listener.Listener,
	route *envoy_route.RouteConfiguration,
) error {
	configurer := plugin_xds.Configurer{
		Rules:  fromRules,
		Subset: core_rules.MeshSubset(),
	}

	for _, chain := range listener.FilterChains {
		if err := configurer.ConfigureFilterChain(chain); err != nil {
			return err
		}
	}
	if err := configurer.ConfigureGatewayRoute(route); err != nil {
		return err
	}

	return nil
}
