package v1alpha1

import (
	"context"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

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
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshRateLimitType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		// MeshRateLimit policy is applied only on DPP
		// todo: add support for ExternalService and ZoneEgress, https://github.com/kumahq/kuma/issues/5050
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
	if err := applyToGateways(ctx, policies.FromRules, listeners.Gateway, routes.Gateway, proxy); err != nil {
		return err
	}
	return nil
}

func applyToGateways(
	ctx xds_context.Context,
	fromRules core_rules.FromRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	proxy *core_xds.Proxy,
) error {
	if !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}
	gatewayListerInfos, err := gateway_plugin.GatewayListenerInfoFromProxy(context.TODO(), ctx.Mesh, proxy, ctx.ControlPlane.Zone)
	if err != nil {
		return err
	}
	for _, listenerInfo := range gatewayListerInfos {
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
		rules, ok := fromRules.Rules[listenerKey]
		if !ok {
			continue
		}

		route, ok := gatewayRoutes[listenerInfo.Listener.ResourceName]
		if !ok {
			continue
		}

		if err := configure(rules, gatewayListener, route); err != nil {
			return err
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

func configure(
	fromRules core_rules.Rules,
	listener *envoy_listener.Listener,
	route *envoy_route.RouteConfiguration,
) error {
	var conf api.Conf
	// Currently, `from` section of MeshRateLimit only allows Mesh targetRef
	if computed := fromRules.Compute(core_rules.MeshSubset()); computed != nil {
		conf = computed.Conf.(api.Conf)
	} else {
		return nil
	}

	configurer := plugin_xds.Configurer{
		Http: conf.Local.HTTP,
		Tcp:  conf.Local.TCP,
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
