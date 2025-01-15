package v1alpha1

import (
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
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
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

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshRetryType, dataplane, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshRetryType]
	if !ok {
		return nil
	}

	listeners := xds.GatherListeners(rs)
	routes := xds.GatherRoutes(rs)

	if err := applyToOutbounds(policies.ToRules, listeners.Outbound, proxy.Outbounds, proxy.Dataplane, ctx.Mesh); err != nil {
		return err
	}

	if err := applyToGateway(policies.GatewayRules, routes.Gateway, listeners.Gateway, proxy); err != nil {
		return err
	}

	err := applyToRealResources(rs, policies.ToRules.ResourceRules, ctx.Mesh)
	if err != nil {
		return err
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
		serviceName := outbound.LegacyOutbound.GetService()

		configurer := plugin_xds.DeprecatedConfigurer{
			Element:  subsetutils.MeshServiceElement(serviceName),
			Rules:    rules.Rules,
			Protocol: meshCtx.GetServiceProtocol(serviceName),
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
	rules core_rules.GatewayRules,
	gatewayRoutes map[string]*envoy_route.RouteConfiguration,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	proxy *core_xds.Proxy,
) error {
	for _, listenerInfo := range gateway_plugin.ExtractGatewayListeners(proxy) {
		listenerKey := core_rules.InboundListener{
			Address: proxy.Dataplane.Spec.GetNetworking().GetAddress(),
			Port:    listenerInfo.Listener.Port,
		}
		listener, ok := gatewayListeners[listenerKey]
		if !ok {
			continue
		}

		toRules := rules.ToRules.ByListener[listenerKey]
		if !ok {
			continue
		}

		var protocol core_mesh.Protocol
		switch listenerInfo.Listener.Protocol {
		case mesh_proto.MeshGateway_Listener_HTTP, mesh_proto.MeshGateway_Listener_HTTPS:
			protocol = core_mesh.ProtocolHTTP
		case mesh_proto.MeshGateway_Listener_TCP, mesh_proto.MeshGateway_Listener_TLS:
			protocol = core_mesh.ProtocolTCP
		}
		configurer := plugin_xds.DeprecatedConfigurer{
			Rules:    toRules.Rules,
			Protocol: protocol,
			Element:  subsetutils.MeshElement(),
		}

		if err := configurer.ConfigureListener(listener); err != nil {
			return err
		}

		for _, listenerHostname := range listenerInfo.ListenerHostnames {
			route, ok := gatewayRoutes[listenerHostname.EnvoyRouteName(listenerInfo.Listener.EnvoyListenerName)]
			if !ok {
				continue
			}

			if err := configurer.ConfigureRoute(route); err != nil {
				return err
			}
		}
	}

	return nil
}

func applyToRealResources(rs *core_xds.ResourceSet, rules outbound.ResourceRules, meshCtx xds_context.MeshContext) error {
	for uri, resType := range rs.IndexByOrigin() {
		conf := rules.Compute(uri, meshCtx.Resources)
		if conf == nil {
			continue
		}

		for typ, resources := range resType {
			switch typ {
			case envoy_resource.ListenerType:
				err := configureListeners(resources, conf.Conf[0].(api.Conf))
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
		configurer := plugin_xds.Configurer{
			Conf:     conf,
			Protocol: resource.Protocol,
		}

		err := configurer.ConfigureListener(resource.Resource.(*envoy_listener.Listener))
		if err != nil {
			return err
		}
	}
	return nil
}
