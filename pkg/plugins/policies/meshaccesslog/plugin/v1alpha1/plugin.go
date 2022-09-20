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

type listeners struct {
	inbound  map[xds.InboundListener]*envoy_listener.Listener
	outbound map[mesh_proto.OutboundInterface]*envoy_listener.Listener
}

func gatherListeners(rs *core_xds.ResourceSet) listeners {
	inboundListeners := map[xds.InboundListener]*envoy_listener.Listener{}
	outboundListeners := map[mesh_proto.OutboundInterface]*envoy_listener.Listener{}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		listener := res.Resource.(*envoy_listener.Listener)
		address := listener.GetAddress().GetSocketAddress()

		switch res.Origin {
		case generator.OriginOutbound:
			outboundListeners[mesh_proto.OutboundInterface{
				DataplaneIP:   address.GetAddress(),
				DataplanePort: address.GetPortValue(),
			}] = listener
		case generator.OriginInbound:
			inboundListeners[xds.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		default:
			continue
		}
	}

	return listeners{
		outbound: outboundListeners,
		inbound:  inboundListeners,
	}
}

func (p plugin) applyToInbounds(
	policies xds.TypedMatchingPolicies, inboundListeners map[xds.InboundListener]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
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

		var conf *api.MeshAccessLog_Conf
		if computed := policies.FromRules.Rules[listenerKey].Compute(xds.Subset{{
			Key: mesh_proto.ServiceTag, Value: serviceName,
		}}); computed != nil {
			conf = computed.(*api.MeshAccessLog_Conf)
		} else {
			continue
		}

		for _, backend := range conf.Backends {
			configurer := Configurer{
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
	}
	return nil
}

func (p plugin) applyToOutbounds(
	policies xds.TypedMatchingPolicies, outboundListeners map[mesh_proto.OutboundInterface]*envoy_listener.Listener, dataplane *core_mesh.DataplaneResource,
) error {
	sourceService := dataplane.Spec.GetIdentifyingService()

	for _, outbound := range dataplane.Spec.Networking.GetOutbound() {
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := outboundListeners[oface]
		if !ok {
			continue
		}

		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]

		var conf *api.MeshAccessLog_Conf
		if computed := policies.ToRules.Rules.Compute(xds.Subset{{
			Key: mesh_proto.ServiceTag, Value: serviceName,
		}}); computed != nil {
			conf = computed.(*api.MeshAccessLog_Conf)
		} else {
			continue
		}

		for _, backend := range conf.Backends {
			configurer := Configurer{
				Mesh:               dataplane.GetMeta().GetMesh(),
				TrafficDirection:   envoy.TrafficDirectionOutbound,
				SourceService:      sourceService,
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
	}

	return nil
}

func (p plugin) Apply(rs *core_xds.ResourceSet, ctx xds_context.Context, proxy *core_xds.Proxy) error {
	policies, ok := proxy.Policies.Dynamic[api.MeshAccessLogType]
	if !ok {
		return nil
	}

	listeners := gatherListeners(rs)

	if err := p.applyToInbounds(policies, listeners.inbound, proxy.Dataplane); err != nil {
		return err
	}
	if err := p.applyToOutbounds(policies, listeners.outbound, proxy.Dataplane); err != nil {
		return err
	}

	return nil
}
