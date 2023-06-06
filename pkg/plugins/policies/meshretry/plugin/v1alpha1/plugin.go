package v1alpha1

import (
	"context"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct{}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshRetryType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshRetryType]
	if !ok {
		return nil
	}

	listeners := xds.GatherListeners(rs)
	routes := xds.GatherRoutes(rs)

	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Dataplane, proxy.Routing); err != nil {
		return err
	}

	if err := applyToGateway(ctx, policies.ToRules, routes.Gateway, listeners.Gateway, proxy); err != nil {
		return err
	}

	return nil
}

func applyToOutbounds(
	rules core_rules.ToRules,
	outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener,
	dataplane *core_mesh.DataplaneResource,
	routing core_xds.Routing,
) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)
		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]

		configurer := plugin_xds.Configurer{
			Retry:    core_rules.ComputeConf[api.Conf](rules.Rules, core_rules.MeshService(serviceName)),
			Protocol: xds.InferProtocol(routing, serviceName),
		}

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}
	}

	return nil
}

func applyToGateway(
	ctx xds_context.Context,
	rules core_rules.ToRules,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
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
		configurer := plugin_xds.Configurer{
			Retry:    core_rules.ComputeConf[api.Conf](rules.Rules, core_rules.MeshSubset()),
			Protocol: core_mesh.ParseProtocol(listenerInfo.Listener.Protocol.String()),
		}

		listener, ok := gatewayListeners[core_rules.InboundListener{
			Address: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
			Port:    listenerInfo.Listener.Port,
		}]
		if !ok {
			continue
		}

		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}

		route, ok := gatewayRoutes[listenerInfo.Listener.ResourceName]
		if !ok {
			continue
		}

		if err := configurer.ConfigureRoute(route); err != nil {
			return err
		}
	}

	return nil
}
