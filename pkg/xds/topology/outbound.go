package topology

import (
	"context"
	"net"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func GetOutboundTargets(ctx context.Context, dataplane *mesh_core.DataplaneResource, manager core_manager.ResourceManager) (map[string][]net.SRV, error) {
	outbound := make(map[string][]net.SRV)
	if len(dataplane.Spec.Networking.GetOutbound()) > 0 {
		dataplanes := &mesh_core.DataplaneResourceList{}
		if err := manager.List(ctx, dataplanes, core_store.ListByMesh(dataplane.Meta.GetMesh())); err != nil {
			return nil, err
		}
		for _, oface := range dataplane.Spec.Networking.GetOutbound() {
			outbound[oface.Service] = make([]net.SRV, 0)
		}
		for _, dataplane := range dataplanes.Items {
			for _, inbound := range dataplane.Spec.Networking.GetInbound() {
				service := inbound.Tags[mesh_proto.ServiceTag]
				endpoints, ok := outbound[service]
				if !ok {
					continue
				}
				iface, err := mesh_proto.ParseInboundInterface(inbound.Interface)
				if err != nil {
					return nil, err
				}
				outbound[service] = append(endpoints, net.SRV{Target: iface.DataplaneIP, Port: uint16(iface.DataplanePort)})
			}
		}
	}
	return outbound, nil
}
