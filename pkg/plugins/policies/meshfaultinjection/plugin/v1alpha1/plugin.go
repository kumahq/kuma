package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshfaultinjection/plugin/xds"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ core_plugins.EgressPolicyPlugin = &plugin{}

type plugin struct{}

func (p plugin) Order() int { return api.MeshFaultInjectionResourceTypeDescriptor.Order }

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshFaultInjectionType, dataplane, resources, opts...)
}

func (p plugin) EgressMatchedPolicies(tags map[string]string, resources xds_context.Resources, opts ...core_plugins.MatchedPoliciesOption) (core_xds.TypedMatchingPolicies, error) {
	return matchers.EgressMatchedPolicies(api.MeshFaultInjectionType, tags, resources, opts...)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	if proxy.ZoneEgressProxy != nil {
		// MeshFaultInjection no longer supports targeting the legacy ExternalService resource via
		// 'from' on zone egress. MeshExternalService fault injection on zone egress is applied
		// through the Dataplane path below using rules-based SNI matches (applyToZoneProxyListener).
		return nil
	}

	if proxy.Dataplane == nil {
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
	if err := applyToZoneProxyListeners(policies.FromRules, listeners, proxy); err != nil {
		return err
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
		protocol := core_meta.ParseProtocol(inbound.GetProtocolFallback())

		listenerKey := core_rules.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}
		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		inboundRules, ok := fromRules.InboundRules[listenerKey]
		if !ok || len(inboundRules) == 0 {
			continue
		}

		switch protocol {
		case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
			configurer := plugin_xds.Configurer{Rules: inboundRules}
			for _, filterChain := range listener.FilterChains {
				if err := configurer.ConfigureHttpListener(filterChain); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func applyToZoneProxyListeners(
	fromRules core_rules.FromRules,
	listeners policies_xds.Listeners,
	proxy *core_xds.Proxy,
) error {
	if !proxy.Dataplane.Spec.GetNetworking().HasZoneProxyListeners() {
		return nil
	}

	for _, listener := range listeners.ZoneIngress {
		if err := applyToZoneProxyListener(fromRules, listener); err != nil {
			return err
		}
	}
	for _, listener := range listeners.ZoneEgress {
		if err := applyToZoneProxyListener(fromRules, listener); err != nil {
			return err
		}
	}

	return nil
}

func applyToZoneProxyListener(
	fromRules core_rules.FromRules,
	listener *envoy_listener.Listener,
) error {
	if listener == nil {
		return nil
	}

	socketAddress := listener.GetAddress().GetSocketAddress()
	if socketAddress == nil {
		return nil
	}

	listenerKey := core_rules.InboundListener{
		Address: socketAddress.GetAddress(),
		Port:    socketAddress.GetPortValue(),
	}
	inboundRules, ok := fromRules.InboundRules[listenerKey]
	if !ok || len(inboundRules) == 0 {
		return nil
	}

	configurer := plugin_xds.Configurer{Rules: inboundRules}
	for _, filterChain := range listener.FilterChains {
		switch policies_xds.FilterChainProtocol(filterChain) {
		case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
			if err := configurer.ConfigureHttpListener(filterChain); err != nil {
				return err
			}
		}
	}

	return nil
}

