package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/metadata"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/generator"
	egress_generator "github.com/kumahq/kuma/pkg/xds/generator/egress"
)

type Listeners struct {
	Inbound         map[core_rules.InboundListener]*envoy_listener.Listener
	Outbound        map[mesh_proto.OutboundInterface]*envoy_listener.Listener
	Egress          *envoy_listener.Listener
	Gateway         map[core_rules.InboundListener]*envoy_listener.Listener
	Ipv4Passthrough *envoy_listener.Listener
	Ipv6Passthrough *envoy_listener.Listener
	DirectAccess    map[generator.Endpoint]*envoy_listener.Listener
	Prometheus      *envoy_listener.Listener
}

func GatherListeners(rs *xds.ResourceSet) *Listeners {
	listeners := Listeners{
		Inbound:      map[core_rules.InboundListener]*envoy_listener.Listener{},
		Outbound:     map[mesh_proto.OutboundInterface]*envoy_listener.Listener{},
		Gateway:      map[core_rules.InboundListener]*envoy_listener.Listener{},
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
			listeners.Inbound[core_rules.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case egress_generator.OriginEgress:
			listeners.Egress = listener
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
			listeners.Gateway[core_rules.InboundListener{
				Address: address.GetAddress(),
				Port:    address.GetPortValue(),
			}] = listener
		case generator.OriginPrometheus:
			listeners.Prometheus = listener
		default:
			continue
		}
	}
	return &listeners
}

func (l *Listeners) ApplyToMeshServiceOutboundListeners(
	dataplane *core_mesh.DataplaneResource,
	meshCtx xds_context.MeshContext,
	configure func(*envoy_listener.Listener, core_mesh.Protocol, core_rules.Subset) error,
) error {
	for _, outbound := range dataplane.Spec.Networking.GetOutbounds(mesh_proto.WithMeshServiceBackendRefFilter) {
		meshService, ok := meshCtx.MeshServiceByName[outbound.BackendRef.Name]
		if !ok {
			continue
		}
		oface := dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listener, ok := l.Outbound[oface]
		if !ok {
			continue
		}

		port, ok := meshService.FindPort(outbound.BackendRef.Port)
		if !ok {
			continue
		}
		subset := core_rules.MeshService(meshService, port)

		err := configure(listener, port.AppProtocol, subset)
		if err != nil {
			return err
		}
	}
	return nil
}
