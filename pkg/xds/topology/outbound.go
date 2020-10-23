package topology

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// BuildEndpointMap creates a map of all endpoints that match given selectors.
func BuildEndpointMap(
	dataplanes []*mesh_core.DataplaneResource,
	zone string,
	mesh *mesh_core.MeshResource,
	externalServices []*mesh_core.ExternalServiceResource,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}
	for _, dataplane := range dataplanes {
		if dataplane.Spec.IsIngress() && dataplane.Spec.IsRemoteIngress(zone) && mesh.MTLSEnabled() {
			for _, ingress := range dataplane.Spec.Networking.GetIngress().GetAvailableServices() {
				if ingress.Mesh != mesh.GetMeta().GetName() {
					continue
				}
				service := ingress.Tags[mesh_proto.ServiceTag]
				outbound[service] = append(outbound[service], core_xds.Endpoint{
					Target:   dataplane.Spec.Networking.Address,
					Port:     dataplane.Spec.Networking.Inbound[0].Port,
					Tags:     ingress.Tags,
					Weight:   ingress.Instances,
					Locality: localityFromTags(ingress.Tags),
				})
			}
			continue
		}
		if !dataplane.Spec.IsIngress() {
			for _, inbound := range dataplane.Spec.Networking.GetInbound() {
				service := inbound.Tags[mesh_proto.ServiceTag]
				iface := dataplane.Spec.Networking.ToInboundInterface(inbound)
				// TODO(yskopets): do we need to dedup?
				// TODO(yskopets): sort ?
				outbound[service] = append(outbound[service], core_xds.Endpoint{
					Target:   iface.DataplaneIP,
					Port:     iface.DataplanePort,
					Tags:     inbound.Tags,
					Weight:   1,
					Locality: localityFromTags(inbound.Tags),
				})
			}
		}
	}

	for _, externalService := range externalServices {
		service := externalService.Spec.GetService()

		tlsEnabled := false
		if externalService.Spec.Networking.Tls != nil {
			tlsEnabled = externalService.Spec.Networking.Tls.Enabled
		}

		tags := externalService.Spec.GetTags()
		if tlsEnabled {
			tags[mesh_proto.ExternalServiceTag] = externalService.Meta.GetName()
		}

		outbound[service] = append(outbound[service], core_xds.Endpoint{
			Target: externalService.Spec.GetHost(),
			Port:   externalService.Spec.GetPortUInt32(),
			Tags:   tags,
			Weight: 1,
			ExternalService: &core_xds.ExternalService{
				TLSEnabled: tlsEnabled,
			},
			Locality: localityFromTags(tags),
		})
	}

	return outbound
}

func localityFromTags(tags map[string]string) *core_xds.Locality {
	region, regionPresent := tags[mesh_proto.RegionTag]
	zone, zonePresent := tags[mesh_proto.ZoneTag]
	subZone, subZonePresent := tags[mesh_proto.SubZoneTag]

	if !regionPresent && !zonePresent && !subZonePresent {
		return nil
	}

	return &core_xds.Locality{
		Region:  region,
		Zone:    zone,
		SubZone: subZone,
	}
}
