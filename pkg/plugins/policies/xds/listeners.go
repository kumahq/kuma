package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

type Listeners struct {
	Inbound         map[xds.InboundListener]*envoy_listener.Listener
	Outbound        map[mesh_proto.OutboundInterface]*envoy_listener.Listener
	Gateway         map[xds.InboundListener]*envoy_listener.Listener
	Ipv4Passthrough *envoy_listener.Listener
	Ipv6Passthrough *envoy_listener.Listener
	DirectAccess    map[generator.Endpoint]*envoy_listener.Listener
}

func GatherListeners(rs *xds.ResourceSet) Listeners {
	listeners := Listeners{
		Inbound:      map[xds.InboundListener]*envoy_listener.Listener{},
		Outbound:     map[mesh_proto.OutboundInterface]*envoy_listener.Listener{},
		Gateway:      map[xds.InboundListener]*envoy_listener.Listener{},
		DirectAccess: map[generator.Endpoint]*envoy_listener.Listener{},
	}

	for _, res := range rs.Resources(envoy_resource.ListenerType) {
		listener := res.Resource.(*envoy_listener.Listener)
		address := listener.GetAddress().GetSocketAddress()

		switch res.Origin {
		case generator.OriginOutbound:
			listeners.Outbound[mesh_proto.OutboundInterface{
				DataplaneIP:   address.GetAddress(),
				DataplanePort: address.GetPortValue(),
			}] = listener
		case generator.OriginInbound:
			listeners.Inbound[xds.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case generator.OriginTransparent:
			switch listener.Name {
			case generator.OutboundNameIPv4:
				listeners.Ipv4Passthrough = listener
			case generator.OutboundNameIPv6:
				listeners.Ipv6Passthrough = listener
			}
		case generator.OriginDirectAccess:
			listeners.DirectAccess[generator.Endpoint{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case metadata.OriginGateway:
			listeners.Gateway[xds.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		default:
			continue
		}
	}

	return listeners
}
