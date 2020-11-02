package topology

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

const (
	// Constants for Locality Aware load balancing
	// Highest priority 0 shall be assigned to all locally available services
	// A priority of 1 is for ExternalServices and services exposed on neighboring ingress-es
	priorityLocal  = 0
	priorityRemote = 1
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
		// Ingress routes the request by TLS SNI, therefore for cross cluster communication MTLS is required
		// We ignore Ingress from endpoints if MTLS is disabled, otherwise we would fail anyway.
		if dataplane.Spec.IsIngress() && dataplane.Spec.IsRemoteIngress(zone) && mesh.MTLSEnabled() && dataplane.Spec.HasPublicAddress() {
			for _, ingress := range dataplane.Spec.Networking.GetIngress().GetAvailableServices() {
				if ingress.Mesh != mesh.GetMeta().GetName() {
					continue
				}
				service := ingress.Tags[mesh_proto.ServiceTag]
				outbound[service] = append(outbound[service], core_xds.Endpoint{
					Target:   dataplane.Spec.Networking.Ingress.PublicAddress,
					Port:     dataplane.Spec.Networking.Ingress.PublicPort,
					Tags:     ingress.Tags,
					Weight:   ingress.Instances,
					Locality: localityFromTags(mesh, priorityRemote, ingress.Tags),
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
					Locality: localityFromTags(mesh, priorityLocal, inbound.Tags),
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
			Locality: localityFromTags(mesh, priorityRemote, tags),
		})
	}

	return outbound
}

func localityFromTags(mesh *mesh_core.MeshResource, priority uint32, tags map[string]string) *core_xds.Locality {
	if !mesh.Spec.GetRouting().GetLocalityAwareLoadBalancing() {
		return nil
	}

	region, regionPresent := tags[mesh_proto.RegionTag]
	zone, zonePresent := tags[mesh_proto.ZoneTag]
	subZone, subZonePresent := tags[mesh_proto.SubZoneTag]

	if !regionPresent && !zonePresent && !subZonePresent {
		return nil
	}

	return &core_xds.Locality{
		Region:   region,
		Zone:     zone,
		SubZone:  subZone,
		Priority: priority,
	}
}
