package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/kumahq/kuma/pkg/core/naming"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	gateway_metadata "github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	generator_meta "github.com/kumahq/kuma/pkg/xds/generator/metadata"
	generator_model "github.com/kumahq/kuma/pkg/xds/generator/model"
)

type Listeners struct {
	Inbound         map[core_rules.InboundListener]*envoy_listener.Listener
	Outbound        map[mesh_proto.OutboundInterface]*envoy_listener.Listener
	Egress          *envoy_listener.Listener
	Gateway         map[core_rules.InboundListener]*envoy_listener.Listener
	Ipv4Passthrough *envoy_listener.Listener
	Ipv6Passthrough *envoy_listener.Listener
	DirectAccess    map[generator_model.Endpoint]*envoy_listener.Listener
	Prometheus      *envoy_listener.Listener
}

func GatherListeners(rs *xds.ResourceSet, unifiedResourceNaming bool) Listeners {
	listeners := Listeners{
		Inbound:      map[core_rules.InboundListener]*envoy_listener.Listener{},
		Outbound:     map[mesh_proto.OutboundInterface]*envoy_listener.Listener{},
		Gateway:      map[core_rules.InboundListener]*envoy_listener.Listener{},
		DirectAccess: map[generator_model.Endpoint]*envoy_listener.Listener{},
	}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		listener := res.Resource.(*envoy_listener.Listener)
		address := listener.GetAddress().GetSocketAddress()

		switch res.Origin {
		case generator_meta.OriginOutbound:
			listeners.Outbound[mesh_proto.OutboundInterface{
				DataplaneIP:   address.GetAddress(),
				DataplanePort: address.GetPortValue(),
			}] = listener
		case generator_meta.OriginInbound:
			listeners.Inbound[core_rules.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case generator_meta.OriginEgress:
			listeners.Egress = listener
		case generator_meta.OriginTransparent:
			switch listener.Name {
			case generator_meta.TransparentOutboundNameIPv4:
				listeners.Ipv4Passthrough = listener
			case generator_meta.TransparentOutboundNameIPv6:
				listeners.Ipv6Passthrough = listener
			case naming.ContextualTransparentProxyName("outbound", 4):
				listeners.Ipv4Passthrough = listener
			case naming.ContextualTransparentProxyName("outbound", 6):
				listeners.Ipv4Passthrough = listener
			}
		case generator_meta.OriginDirectAccess:
			listeners.DirectAccess[generator_model.Endpoint{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case gateway_metadata.OriginGateway:
			listeners.Gateway[core_rules.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case generator_meta.OriginPrometheus:
			listeners.Prometheus = listener
		default:
			continue
		}
	}
	return listeners
}
