package topology

import (
	"context"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"

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
			log.Info("Unable to create ExternalService endpoint", err)
			continue
		}
		outbound[service] = append(outbound[service], *externalServiceEndpoint)
	}

	return outbound
}

func buildExternalServiceEndpoint(externalService *mesh_core.ExternalServiceResource, mesh string, loader datasource.Loader) (*core_xds.Endpoint, error) {
	tlsEnabled := false
	var caCert, clientCert, clientKey []byte = nil, nil, nil

	tags := externalService.Spec.GetTags()
	if externalService.Spec.Networking.Tls != nil {
		var err error
		tlsEnabled = externalService.Spec.Networking.Tls.Enabled
		if tlsEnabled {
			tags[`kuma.io/external-service-name`] = externalService.Meta.GetName()
		}

		if externalService.Spec.Networking.Tls.CaCert != nil {
			caCert, err = loader.Load(context.Background(), mesh, externalService.Spec.Networking.Tls.CaCert)
			if err != nil {
				return nil, errors.Wrap(err, "Error getting CA certificate")
			}
		}

		if externalService.Spec.Networking.Tls.ClientCert != nil {
			clientCert, err = loader.Load(context.Background(), mesh, externalService.Spec.Networking.Tls.ClientCert)
			if err != nil {
				return nil, errors.Wrap(err, "Error getting client certificate")
			}
		}

		if externalService.Spec.Networking.Tls.ClientKey != nil {
			clientKey, err = loader.Load(context.Background(), mesh, externalService.Spec.Networking.Tls.ClientKey)
			if err != nil {
				return nil, errors.Wrap(err, "Error getting client key")
			}
		}
	}

	return &core_xds.Endpoint{
		Target: externalService.Spec.GetHost(),
		Port:   externalService.Spec.GetPortUInt32(),
		Tags:   tags,
		Weight: 1,
		ExternalService: &core_xds.ExternalService{
			TLSEnabled: tlsEnabled,
			CaCert:     caCert,
			ClientCert: clientCert,
			ClientKey:  clientKey,
		},
	}, nil
}
