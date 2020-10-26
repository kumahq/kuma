package topology

import (
	"github.com/kumahq/kuma/pkg/core"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/datasource"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

// BuildEndpointMap creates a map of all endpoints that match given selectors.
func BuildEndpointMap(
	mesh *mesh_core.MeshResource,
	zone string,
	dataplanes []*mesh_core.DataplaneResource,
	externalServices []*mesh_core.ExternalServiceResource,
	loader datasource.Loader,
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
					Target: dataplane.Spec.Networking.Address,
					Port:   dataplane.Spec.Networking.Inbound[0].Port,
					Tags:   ingress.Tags,
					Weight: ingress.Instances,
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

		externalServiceEndpoint, err := buildExternalServiceEndpoint(externalService, mesh.Meta.GetMesh(), loader)
		if err != nil {
			core.Log.Error(err, "unable to create ExternalService endpoint. Endpoint won't be included in the XDS.", "name", externalService.Meta.GetName(), "mesh", externalService.Meta.GetMesh())
			continue
		}
		outbound[service] = append(outbound[service], *externalServiceEndpoint)
	}

	return outbound
}

func buildExternalServiceEndpoint(externalService *mesh_core.ExternalServiceResource, mesh string, loader datasource.Loader) (*core_xds.Endpoint, error) {
	es := &core_xds.ExternalService{
		TLSEnabled: externalService.Spec.GetNetworking().GetTls().GetEnabled(),
		CaCert:     externalService.Spec.GetNetworking().GetTls().GetCaCert().ConvertToEnvoy(),
		ClientCert: externalService.Spec.GetNetworking().GetTls().GetClientCert().ConvertToEnvoy(),
		ClientKey:  externalService.Spec.GetNetworking().GetTls().GetClientKey().ConvertToEnvoy(),
	}

	tags := externalService.Spec.GetTags()
	if es.TLSEnabled {
		tags[`kuma.io/external-service-name`] = externalService.Meta.GetName()
	}

	return &core_xds.Endpoint{
		Target:          externalService.Spec.GetHost(),
		Port:            externalService.Spec.GetPortUInt32(),
		Tags:            tags,
		Weight:          1,
		ExternalService: es,
	}, nil
}
