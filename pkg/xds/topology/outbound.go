package topology

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// GetOutboundTargets resolves all endpoints reachable from a given dataplane.
func GetOutboundTargets(destinations core_xds.DestinationMap,
	dataplanes *mesh_core.DataplaneResourceList,
	externalServices *mesh_core.ExternalServiceResourceList,
	localClusterName string, mesh *mesh_core.MeshResource) (core_xds.EndpointMap, error) {
	if len(destinations) == 0 {
		return nil, nil
	}
	return BuildEndpointMap(destinations, dataplanes.Items, externalServices.Items, localClusterName, mesh), nil
}

// BuildEndpointMap creates a map of all endpoints that match given selectors.
func BuildEndpointMap(destinations core_xds.DestinationMap,
	dataplanes []*mesh_core.DataplaneResource,
	externalServices []*mesh_core.ExternalServiceResource,
	zone string, mesh *mesh_core.MeshResource) core_xds.EndpointMap {
	if len(destinations) == 0 {
		return nil
	}
	outbound := core_xds.EndpointMap{}
	for _, dataplane := range dataplanes {
		if dataplane.Spec.IsIngress() && mesh.MTLSEnabled() {
			if dataplane.Spec.IsRemoteIngress(zone) {
				for _, ingress := range dataplane.Spec.Networking.GetIngress().GetAvailableServices() {
					if ingress.Mesh != mesh.GetMeta().GetName() {
						continue
					}
					service := ingress.Tags[mesh_proto.ServiceTag]
					selectors, ok := destinations[service]
					if !ok {
						continue
					}
					if !selectors.Matches(ingress.Tags) {
						continue
					}
					outbound[service] = append(outbound[service], core_xds.Endpoint{
						Target: dataplane.Spec.Networking.Address,
						Port:   dataplane.Spec.Networking.Inbound[0].Port,
						Tags:   ingress.Tags,
						Weight: ingress.Instances,
					})
				}
			}
		} else {
			for _, inbound := range dataplane.Spec.Networking.GetInbound() {
				service := inbound.Tags[mesh_proto.ServiceTag]
				selectors, ok := destinations[service]
				if !ok {
					continue
				}
				if !selectors.Matches(inbound.Tags) {
					continue
				}
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

	for _, externalService := range externalServices {
		service := externalService.Spec.GetService()
		selectors, ok := destinations[service]
		if !ok {
			continue
		}
		if !selectors.Matches(externalService.Spec.Tags) {
			continue
		}

		tlsEnabled := false
		if externalService.Spec.Networking.Tls != nil {
			tlsEnabled = externalService.Spec.Networking.Tls.Enabled
		}

		tags := externalService.Spec.GetTags()
		if tlsEnabled {
			tags[mesh_proto.ServiceTag+"_tls"] = "enabled"
		}

		outbound[service] = append(outbound[service], core_xds.Endpoint{
			Target: externalService.Spec.GetHost(),
			Port:   externalService.Spec.GetPortUInt32(),
			Tags:   tags,
			Weight: 1,
			ExternalService: &core_xds.ExternalService{
				TLSEnabled: tlsEnabled,
			},
		})
	}

	return outbound
}
