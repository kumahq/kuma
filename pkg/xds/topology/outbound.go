package topology

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// BuildEndpointMap creates a map of all endpoints that match given selectors.
func BuildEndpointMap(dataplanes []*mesh_core.DataplaneResource, zone string, mesh *mesh_core.MeshResource) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}
	for _, dataplane := range dataplanes {
		if dataplane.Spec.IsIngress() && dataplane.Spec.IsRemoteIngress(zone) && mesh.MTLSEnabled() {
			for _, ingress := range dataplane.Spec.Networking.GetIngress().GetAvailableServices() {
				if ingress.Mesh != mesh.GetMeta().GetName() {
					continue
				}
				service := ingress.Tags[mesh_proto.ServiceTag]
				outbound[service] = append(outbound[service], core_xds.Endpoint{
					Target: dataplane.Spec.Networking.Address,
					Port:   dataplane.Spec.Networking.Inbound[0].Port,
					Tags:   ingress.Tags,
					Weight: ingress.Instances,
				})
			}
			return outbound
		}
		if !dataplane.Spec.IsIngress() {
			for _, inbound := range dataplane.Spec.Networking.GetInbound() {
				service := inbound.Tags[mesh_proto.ServiceTag]
				iface := dataplane.Spec.Networking.ToInboundInterface(inbound)
				// TODO(yskopets): do we need to dedup?
				// TODO(yskopets): sort ?
				outbound[service] = append(outbound[service], core_xds.Endpoint{
					Target: iface.DataplaneIP,
					Port:   iface.DataplanePort,
					Tags:   inbound.Tags,
					Weight: 1,
				})
			}
		}
	}
	return outbound
}
