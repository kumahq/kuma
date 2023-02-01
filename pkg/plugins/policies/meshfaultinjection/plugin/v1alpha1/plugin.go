package v1alpha1

import (
	"context"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/plugin/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshFaultInjectionType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.Dataplane == nil {
		// MeshRateLimit policy is applied only on DPP
		// todo: add support for ExternalService and ZoneEgress, https://github.com/kumahq/kuma/issues/5050
		return nil
	}
	policies, ok := proxy.Policies.Dynamic[api.MeshFaultInjectionType]
	if !ok {
		return nil
	}
	listeners := policies_xds.GatherListeners(rs)

	if err := applyToInbounds(policies.FromRules, listeners.Inbound, proxy); err != nil {
		return err
	}

	if err := applyToGateways(ctx, policies.FromRules, listeners.Gateway, proxy); err != nil {
		return err
	}
	return nil
}

func applyToInbounds(
	fromRules core_xds.FromRules,
	inboundListeners map[core_xds.InboundListener]*envoy_listener.Listener,
	proxy *core_xds.Proxy,
) error {
	for _, inbound := range proxy.Dataplane.Spec.GetNetworking().GetInbound() {
		iface := proxy.Dataplane.Spec.Networking.ToInboundInterface(inbound)
		protocol := core_mesh.ParseProtocol(inbound.GetProtocol())
		if _, exists := proxy.Policies.FaultInjections[iface]; exists {
			continue
		}

		listenerKey := core_xds.InboundListener{
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

		if err := configure(rules, listener, protocol); err != nil {
			return err
		}
	}
	return nil
}

func applyToGateways(
	ctx xds_context.Context,
	fromRules core_xds.FromRules,
	gatewayListeners map[core_xds.InboundListener]*envoy_listener.Listener,
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
		protocol := core_mesh.ParseProtocol(mesh_proto.MeshGateway_Listener_Protocol_name[int32(listenerInfo.Listener.Protocol)])
		listenerKey := core_xds.InboundListener{
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

		if err := configure(rules, gatewayListener, protocol); err != nil {
			return err
		}
	}
	return nil
}

func configure(
	fromRules core_xds.Rules,
	listener *envoy_listener.Listener,
	protocol core_mesh.Protocol,
) error {
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		for _, rule := range fromRules {
			conf := rule.Conf.(api.Conf)
			from := rule.Subset

			configurer := plugin_xds.Configurer{
				FaultInjections: pointer.Deref(conf.Http),
				From:            from,
			}

			for _, chain := range listener.FilterChains {
				if err := configurer.ConfigureHttpListener(chain); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
