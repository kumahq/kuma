package topology

import (
	"context"
	"net"
	"strconv"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
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
	mesh *mesh_core.MeshResource,
	zone string,
	dataplanes []*mesh_core.DataplaneResource,
	externalServices []*mesh_core.ExternalServiceResource,
	loader datasource.Loader,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}
	ingressInstances := fillIngressOutbounds(outbound, dataplanes, zone, mesh)
	endpointWeight := uint32(1)
	if ingressInstances > 0 {
		endpointWeight = ingressInstances
	}
	fillDataplaneOutbounds(outbound, dataplanes, mesh, endpointWeight)
	fillExternalServicesOutbounds(outbound, externalServices, mesh, loader)
	return outbound
}

// endpointWeight defines default weight for in-cluster endpoint.
// Examples of having service "backend":
// 1) Standalone deployment, 2 instances in one cluster (zone1)
//    All endpoints have to have the same weight (ex. 1) to achieve fair loadbalancing.
//    Endpoints:
//    * backend-zone1-1 - weight: 1
//    * backend-zone1-2 - weight: 1
// 2) Multi-zone deployment, 2 instances in "zone1" (local zone), 3 instances in "zone2" (remote zone) with 1 Ingress instance
//    Endpoints:
//    * backend-zone1-1 - weight: 1
//    * backend-zone1-2 - weight: 1
//    * ingress-zone2-1 - weight: 3 (all remote endpoints are aggregated to one Ingress, it needs to have weight of instances in other cluster)
// 3) Multi-zone deployment, 2 instances in "zone1" (local zone), 2 instances in "zone2" (remote zone) with 1 Ingress instance
//    Many instances of Ingress will forward the traffic to the same endpoints in "zone2" so we need to lower the weights.
//    Since weights are integers, we cannot put fractional on ingress endpoints weights, we need to adjust "default" weight for local zone
//    Endpoints:
//    * backend-zone1-1 - weight: 2
//    * backend-zone1-2 - weight: 2
//    * ingress-zone2-1 - weight: 3
//    * ingress-zone2-2 - weight: 3
func fillDataplaneOutbounds(outbound core_xds.EndpointMap, dataplanes []*mesh_core.DataplaneResource, mesh *mesh_core.MeshResource, endpointWeight uint32) {
	for _, dataplane := range dataplanes {
		if dataplane.Spec.IsIngress() {
			continue
		}
		for _, inbound := range dataplane.Spec.GetNetworking().GetHealthyInbounds() {
			service := inbound.Tags[mesh_proto.ServiceTag]
			iface := dataplane.Spec.Networking.ToInboundInterface(inbound)
			// TODO(yskopets): do we need to dedup?
			// TODO(yskopets): sort ?
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target:   iface.DataplaneIP,
				Port:     iface.DataplanePort,
				Tags:     inbound.Tags,
				Weight:   endpointWeight,
				Locality: localityFromTags(mesh, priorityLocal, inbound.Tags),
			})
		}
	}
}

func fillIngressOutbounds(outbound core_xds.EndpointMap, dataplanes []*mesh_core.DataplaneResource, zone string, mesh *mesh_core.MeshResource) uint32 {
	ingressInstances := map[string]bool{}
	for _, dataplane := range dataplanes {
		if !dataplane.Spec.IsIngress() {
			continue
		}
		if !dataplane.Spec.IsRemoteIngress(zone) {
			continue // we only need Ingress for other zones, we don't want to direct request to the same zone through Ingress
		}
		if !mesh.MTLSEnabled() {
			// Ingress routes the request by TLS SNI, therefore for cross cluster communication MTLS is required
			// We ignore Ingress from endpoints if MTLS is disabled, otherwise we would fail anyway.
			continue
		}
		if !dataplane.Spec.HasPublicAddress() {
			continue // Dataplane is not reachable yet from other clusters. This may happen when Ingress Service is pending waiting on External IP on Kubernetes.
		}
		ingressCoordinates := net.JoinHostPort(dataplane.Spec.Networking.Ingress.PublicAddress,
			strconv.FormatUint(uint64(dataplane.Spec.Networking.Ingress.PublicPort), 10))
		if ingressInstances[ingressCoordinates] {
			continue // many Ingress instances can be placed in front of one load balancer (all instances can have the same public address and port). In this case we only need one Instance avoiding creating unnecessary duplicated endpoints
		}
		ingressInstances[ingressCoordinates] = true
		for _, service := range dataplane.Spec.Networking.GetIngress().GetAvailableServices() {
			if service.Mesh != mesh.GetMeta().GetName() {
				continue
			}
			serviceName := service.Tags[mesh_proto.ServiceTag]
			outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
				Target:   dataplane.Spec.Networking.Ingress.PublicAddress,
				Port:     dataplane.Spec.Networking.Ingress.PublicPort,
				Tags:     service.Tags,
				Weight:   service.Instances,
				Locality: localityFromTags(mesh, priorityRemote, service.Tags),
			})
		}
	}
	return uint32(len(ingressInstances))
}

func fillExternalServicesOutbounds(outbound core_xds.EndpointMap, externalServices []*mesh_core.ExternalServiceResource, mesh *mesh_core.MeshResource, loader datasource.Loader) {
	for _, externalService := range externalServices {
		service := externalService.Spec.GetService()

		externalServiceEndpoint, err := buildExternalServiceEndpoint(externalService, mesh, loader)
		if err != nil {
			core.Log.Error(err, "unable to create ExternalService endpoint. Endpoint won't be included in the XDS.", "name", externalService.Meta.GetName(), "mesh", externalService.Meta.GetMesh())
			continue
		}
		outbound[service] = append(outbound[service], *externalServiceEndpoint)
	}
}

func buildExternalServiceEndpoint(externalService *mesh_core.ExternalServiceResource, mesh *mesh_core.MeshResource, loader datasource.Loader) (*core_xds.Endpoint, error) {
	host, _, err := net.SplitHostPort(externalService.Spec.Networking.Address)
	if err != nil {
		return nil, err
	}
	if !govalidator.IsDNSName(host) {
		host = ""
	}
	es := &core_xds.ExternalService{
		Host:       host,
		TLSEnabled: externalService.Spec.GetNetworking().GetTls().GetEnabled(),
		CaCert: convertToEnvoy(
			externalService.Spec.GetNetworking().GetTls().GetCaCert(),
			mesh.GetMeta().GetName(), loader),
		ClientCert: convertToEnvoy(
			externalService.Spec.GetNetworking().GetTls().GetClientCert(),
			mesh.GetMeta().GetName(), loader),
		ClientKey: convertToEnvoy(
			externalService.Spec.GetNetworking().GetTls().GetClientKey(),
			mesh.GetMeta().GetName(), loader),
	}

	tags := externalService.Spec.GetTags()
	if es.TLSEnabled {
		tags = envoy.Tags(tags).WithTags(mesh_proto.ExternalServiceTag, externalService.Meta.GetName())
	}

	return &core_xds.Endpoint{
		Target:          externalService.Spec.GetHost(),
		Port:            externalService.Spec.GetPortUInt32(),
		Tags:            tags,
		Weight:          1,
		ExternalService: es,
		Locality:        localityFromTags(mesh, priorityRemote, tags),
	}, nil
}

func convertToEnvoy(ds *v1alpha1.DataSource, mesh string, loader datasource.Loader) []byte {
	if ds == nil {
		return nil
	}

	data, err := loader.Load(context.Background(), mesh, ds)
	if err != nil {
		core.Log.Error(err, "failed to resolve ingress's public name, skipping dataplane")
		return nil
	}

	return data
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
