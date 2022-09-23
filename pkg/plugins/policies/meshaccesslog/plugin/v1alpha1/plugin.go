package v1alpha1

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	plugin_xds "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/plugin/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

var _ core_plugins.PolicyPlugin = &plugin{}

type plugin struct {
}

func NewPlugin() core_plugins.Plugin {
	return &plugin{}
}

func (p plugin) MatchedPolicies(dataplane *core_mesh.DataplaneResource, resources xds_context.Resources) (core_xds.TypedMatchingPolicies, error) {
	return matchers.MatchedPolicies(api.MeshAccessLogType, dataplane, resources)
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshAccessLogType]
	if !ok {
		return nil
	}

	listeners := gatherListeners(rs)

	if err := applyToInbounds(policies.FromRules, listeners.inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToOutbounds(policies.ToRules, listeners.outbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToTransparentProxyListeners(policies, listeners.ipv4Passthrough, listeners.ipv6Passthrough, proxy.Dataplane); err != nil {
		return err
	}
	if err := applyToDirectAccess(policies.ToRules, listeners.directAccess, proxy.Dataplane); err != nil {
		return err
	}

	return nil
}

func applyToInbounds(
	rules xds.FromRules, inboundListeners map[xds.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
) error {
	for _, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
		iface := dataplane.Spec.Networking.ToInboundInterface(inbound)

		listenerKey := xds.InboundListener{
			Address: iface.DataplaneIP,
			Port:    iface.DataplanePort,
		}
		listener, ok := inboundListeners[listenerKey]
		if !ok {
			continue
		}

		serviceName := inbound.GetTags()[mesh_proto.ServiceTag]

		if err := configureInbound(rules.Rules[listenerKey], dataplane, xds.MeshService(serviceName), listener); err != nil {
			return err
		}
	}
	return nil
}

func applyToOutbounds(
	rules xds.ToRules, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]

		if err := configureOutbound(rules, dataplane, xds.MeshService(serviceName), serviceName, listener); err != nil {
			return err
		}
	}

	return nil
}

func applyToTransparentProxyListeners(
	policies xds.TypedMatchingPolicies, ipv4 *envoy_listener.Listener, ipv6 *envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
) error {
	// TODO inbound listener?
	if ipv4 != nil {
		if err := configureOutbound(
			policies.ToRules,
			dataplane,
			xds.MeshService(core_mesh.PassThroughService),
			"external",
			ipv4,
		); err != nil {
			return err
		}
	}

	if ipv6 != nil {
		return configureOutbound(
			policies.ToRules,
			dataplane,
			xds.MeshService(core_mesh.PassThroughService),
			"external",
			ipv6,
		)
	}

	return nil
}

func applyToDirectAccess(
	rules xds.ToRules, directAccess map[generator.Endpoint]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
) error {
	for endpoint, listener := range directAccess {
		name := generator.DirectAccessEndpointName(endpoint)
		return configureOutbound(
			rules,
			dataplane,
			xds.MeshService(core_mesh.PassThroughService),
			name,
			listener,
		)
	}

	return nil
}

type listeners struct {
	inbound         map[xds.InboundListener]*envoy_listener.Listener
	outbound        map[mesh_proto.OutboundInterface]*envoy_listener.Listener
	ipv4Passthrough *envoy_listener.Listener
	ipv6Passthrough *envoy_listener.Listener
	directAccess    map[generator.Endpoint]*envoy_listener.Listener
}

func gatherListeners(rs *core_xds.ResourceSet) listeners {
	listeners := listeners{
		inbound:      map[xds.InboundListener]*envoy_listener.Listener{},
		outbound:     map[mesh_proto.OutboundInterface]*envoy_listener.Listener{},
		directAccess: map[generator.Endpoint]*envoy_listener.Listener{},
	}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		listener := res.Resource.(*envoy_listener.Listener)
		address := listener.GetAddress().GetSocketAddress()

		switch res.Origin {
		case generator.OriginOutbound:
			listeners.outbound[mesh_proto.OutboundInterface{
				DataplaneIP:   address.GetAddress(),
				DataplanePort: address.GetPortValue(),
			}] = listener
		case generator.OriginInbound:
			listeners.inbound[xds.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case generator.OriginTransparent:
			switch listener.Name {
			case generator.OutboundNameIPv4:
				listeners.ipv4Passthrough = listener
			case generator.OutboundNameIPv6:
				listeners.ipv6Passthrough = listener
			}
		case generator.OriginDirectAccess:
			listeners.directAccess[generator.Endpoint{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		default:
			continue
		}
	}

	return listeners
}

func configureInbound(
	fromRules xds.Rules,
	dataplane *core_mesh.DataplaneResource,
	subset xds.Subset,
	listener *envoy_listener.Listener,
) error {
	serviceName := dataplane.Spec.GetIdentifyingService()

	var conf *api.MeshAccessLog_Conf
	if computed := fromRules.Compute(subset); computed != nil {
		conf = computed.(*api.MeshAccessLog_Conf)
	} else {
		return nil
	}

	for _, backend := range conf.Backends {
		configurer := plugin_xds.Configurer{
			Mesh:               dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionInbound,
			SourceService:      mesh_proto.ServiceUnknown,
			DestinationService: serviceName,
			Backend:            backend,
			Dataplane:          dataplane,
		}

		for _, chain := range listener.FilterChains {
			if err := configurer.Configure(chain); err != nil {
				return err
			}
		}
	}

	return nil
}

func configureOutbound(
	toRules xds.ToRules,
	dataplane *core_mesh.DataplaneResource,
	subset xds.Subset,
	destinationServiceName string,
	listener *envoy_listener.Listener,
) error {
	sourceService := dataplane.Spec.GetIdentifyingService()

	var conf *api.MeshAccessLog_Conf
	if computed := toRules.Rules.Compute(subset); computed != nil {
		conf = computed.(*api.MeshAccessLog_Conf)
	} else {
		return nil
	}

	for _, backend := range conf.Backends {
		configurer := plugin_xds.Configurer{
			Mesh:               dataplane.GetMeta().GetMesh(),
			TrafficDirection:   envoy.TrafficDirectionOutbound,
			SourceService:      sourceService,
			DestinationService: destinationServiceName,
			Backend:            backend,
			Dataplane:          dataplane,
		}

		for _, chain := range listener.FilterChains {
			if err := configurer.Configure(chain); err != nil {
				return err
			}
		}
	}

	return nil
}
