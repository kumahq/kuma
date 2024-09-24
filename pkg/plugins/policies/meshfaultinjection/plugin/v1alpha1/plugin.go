package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util "github.com/kumahq/kuma/pkg/plugins/policies/core/egress"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/plugin/xds"
	gateway_plugin "github.com/kumahq/kuma/pkg/plugins/runtime/gateway"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var (
	_   core_plugins.EgressPolicyPlugin = &plugin{}
	log                                 = core.Log.WithName("MeshFaultInjection")
)

type plugin struct{}

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
		return applyToEgress(rs, proxy)
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

	if err := applyToGateways(policies.GatewayRules, listeners.Gateway, proxy); err != nil {
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
		protocol := core_mesh.ParseProtocol(inbound.GetProtocol())
		if _, exists := proxy.Policies.FaultInjections[iface]; exists {
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

		for _, filterChain := range listener.FilterChains {
			if err := configure(rules, filterChain, protocol); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyToGateways(
	rules core_rules.GatewayRules,
	gatewayListeners map[core_rules.InboundListener]*envoy_listener.Listener,
	proxy *core_xds.Proxy,
) error {
	if !proxy.Dataplane.Spec.IsBuiltinGateway() {
		return nil
	}
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
		rules, ok := rules.ToRules.ByListener[listenerKey]
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
		for _, filterChain := range gatewayListener.FilterChains {
			if err := configure(rules.Rules, filterChain, protocol); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyToEgress(rs *core_xds.ResourceSet, proxy *core_xds.Proxy) error {
	listeners := policies_xds.GatherListeners(rs)
	if listeners.Egress == nil {
		log.V(1).Info("skip applying MeshFaultInjection, Egress has no listener",
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
			mfi, ok := policies[api.MeshFaultInjectionType]
			if !ok {
				continue
			}
			protocol := util.GetExternalServiceProtocol(es)
			for _, rule := range mfi.FromRules.Rules {
				for _, filterChain := range listeners.Egress.FilterChains {
					if filterChain.Name == names.GetEgressFilterChainName(esName, meshName) {
						if err := configure(rule, filterChain, protocol); err != nil {
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
	filterChain *envoy_listener.FilterChain,
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

			if err := configurer.ConfigureHttpListener(filterChain); err != nil {
				return err
			}
		}
	}
	return nil
}
