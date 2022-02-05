package topology

import (
	"context"
	"net"
	"strconv"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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
	mesh *core_mesh.MeshResource,
	zone string,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	externalServices []*core_mesh.ExternalServiceResource,
	loader datasource.Loader,
) core_xds.EndpointMap {
	outbound := BuildEdsEndpointMap(mesh, zone, dataplanes, zoneIngresses)
	fillExternalServicesOutbounds(outbound, externalServices, mesh, loader, zone)
	return outbound
}

func BuildEdsEndpointMap(
	mesh *core_mesh.MeshResource,
	zone string,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}
	ingressInstances := fillIngressOutbounds(outbound, zoneIngresses, zone, mesh)
	endpointWeight := uint32(1)
	if ingressInstances > 0 {
		endpointWeight = ingressInstances
	}
	fillDataplaneOutbounds(outbound, dataplanes, mesh, endpointWeight)
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
func fillDataplaneOutbounds(outbound core_xds.EndpointMap, dataplanes []*core_mesh.DataplaneResource, mesh *core_mesh.MeshResource, endpointWeight uint32) {
	for _, dataplane := range dataplanes {
		for _, inbound := range dataplane.Spec.GetNetworking().GetHealthyInbounds() {
			service := inbound.Tags[mesh_proto.ServiceTag]
			iface := dataplane.Spec.Networking.ToInboundInterface(inbound)
			// TODO(yskopets): do we need to dedup?
			// TODO(yskopets): sort ?
			outbound[service] = append(outbound[service], core_xds.Endpoint{
				Target:   iface.DataplaneAdvertisedIP,
				Port:     iface.DataplanePort,
				Tags:     inbound.Tags,
				Weight:   endpointWeight,
				Locality: localityFromTags(mesh, priorityLocal, inbound.Tags),
			})
		}
	}
}

func fillIngressOutbounds(
	outbound core_xds.EndpointMap,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	zone string,
	mesh *core_mesh.MeshResource,
) uint32 {
	ingressInstances := map[string]bool{}
	for _, zi := range zoneIngresses {
		if !zi.IsRemoteIngress(zone) {
			continue
		}
		if !mesh.MTLSEnabled() {
			// Ingress routes the request by TLS SNI, therefore for cross cluster communication MTLS is required
			// We ignore Ingress from endpoints if MTLS is disabled, otherwise we would fail anyway.
			continue
		}
		if !zi.HasPublicAddress() {
			continue // Zone Ingress is not reachable yet from other clusters. This may happen when Ingress Service is pending waiting on External IP on Kubernetes.
		}
		ingressCoordinates := net.JoinHostPort(zi.Spec.GetNetworking().GetAdvertisedAddress(), strconv.FormatUint(uint64(zi.Spec.GetNetworking().GetAdvertisedPort()), 10))
		if ingressInstances[ingressCoordinates] {
			continue // many Ingress instances can be placed in front of one load balancer (all instances can have the same public address and port). In this case we only need one Instance avoiding creating unnecessary duplicated endpoints
		}
		ingressInstances[ingressCoordinates] = true
		for _, service := range zi.Spec.GetAvailableServices() {
			if service.Mesh != mesh.GetMeta().GetName() {
				continue
			}
			serviceName := service.Tags[mesh_proto.ServiceTag]
			outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
				Target:   zi.Spec.GetNetworking().GetAdvertisedAddress(),
				Port:     zi.Spec.GetNetworking().GetAdvertisedPort(),
				Tags:     service.Tags,
				Weight:   service.Instances,
				Locality: localityFromTags(mesh, priorityRemote, service.Tags),
			})
		}
	}
	return uint32(len(ingressInstances))
}

func fillExternalServicesOutbounds(outbound core_xds.EndpointMap, externalServices []*core_mesh.ExternalServiceResource, mesh *core_mesh.MeshResource, loader datasource.Loader, zone string) {
	for _, externalService := range externalServices {
		service := externalService.Spec.GetService()

		externalServiceEndpoint, err := NewExternalServiceEndpoint(externalService, mesh, loader, zone)
		if err != nil {
			core.Log.Error(err, "unable to create ExternalService endpoint. Endpoint won't be included in the XDS.", "name", externalService.Meta.GetName(), "mesh", externalService.Meta.GetMesh())
			continue
		}
		outbound[service] = append(outbound[service], *externalServiceEndpoint)
	}
}

// NewExternalServiceEndpoint builds a new Endpoint from an ExternalServiceResource.
func NewExternalServiceEndpoint(
	externalService *core_mesh.ExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
	zone string,
) (*core_xds.Endpoint, error) {
	spec := externalService.Spec
	tls := spec.GetNetworking().GetTls()
	meshName := mesh.GetMeta().GetName()
	tags := spec.GetTags()

	es := &core_xds.ExternalService{
		TLSEnabled:         tls.GetEnabled(),
		CaCert:             convertToEnvoy(tls.GetCaCert(), meshName, loader),
		ClientCert:         convertToEnvoy(tls.GetClientCert(), meshName, loader),
		ClientKey:          convertToEnvoy(tls.GetClientKey(), meshName, loader),
		AllowRenegotiation: tls.GetAllowRenegotiation().GetValue(),
		ServerName:         tls.GetServerName().GetValue(),
	}

	if es.TLSEnabled {
		name := externalService.Meta.GetName()

		tags = envoy.Tags(tags).
			WithTags(mesh_proto.ExternalServiceTag, name)
	}

	var priority uint32 = priorityRemote
	if esZone, ok := spec.GetTags()[mesh_proto.ZoneTag]; ok && esZone == zone {
		priority = priorityLocal
	}

	return &core_xds.Endpoint{
		Target:          spec.GetHost(),
		Port:            spec.GetPortUInt32(),
		Tags:            tags,
		Weight:          1,
		ExternalService: es,
		Locality:        localityFromTags(mesh, priority, tags),
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

func localityFromTags(mesh *core_mesh.MeshResource, priority uint32, tags map[string]string) *core_xds.Locality {
	zone, zonePresent := tags[mesh_proto.ZoneTag]

	if !zonePresent {
		// this means that we are running in standalone since in multi-zone Kuma always adds Zone tag automatically
		return nil
	}

	if !mesh.Spec.GetRouting().GetLocalityAwareLoadBalancing() {
		// we want to set the Locality even when localityAwareLoadBalancing is not enabled
		// If we set the locality we have an extra visibility about this in /clusters etc.
		// Kuma's LocalityAwareLoadBalancing feature is based only on Priority therefore when it's disabled we need to set Priority to local
		//
		// Setting this regardless of LocalityAwareLoadBalancing on the mesh also solves the problem that endpoints have problems when moving from one locality to another
		// https://github.com/envoyproxy/envoy/issues/12392
		priority = priorityLocal
	}

	return &core_xds.Locality{
		Zone:     zone,
		Priority: priority,
	}
}
