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
	// The Highest priority 0 shall be assigned to all locally available services
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
	zoneEgresses []*core_mesh.ZoneEgressResource,
	externalServices []*core_mesh.ExternalServiceResource,
	loader datasource.Loader,
) core_xds.EndpointMap {
	outbound := BuildEdsEndpointMap(mesh, zone, dataplanes, zoneIngresses, zoneEgresses, externalServices)

	if len(zoneEgresses) == 0 || !mesh.MTLSEnabled() {
		fillExternalServicesOutbounds(outbound, externalServices, mesh, loader, zone)
	}

	return outbound
}

// BuildRemoteEndpointMap creates a map of endpoints that match given selectors
// and are not local for the provided zone (external services and services
// behind remote zone ingress only)
func BuildRemoteEndpointMap(
	mesh *core_mesh.MeshResource,
	zone string,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	externalServices []*core_mesh.ExternalServiceResource,
	loader datasource.Loader,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	fillIngressOutbounds(outbound, zoneIngresses, nil, zone, mesh)
	fillExternalServicesOutbounds(outbound, externalServices, mesh, loader, zone)

	for serviceName, endpoints := range outbound {
		var newEndpoints []core_xds.Endpoint

		for _, endpoint := range endpoints {
			endpoint.Tags["mesh"] = mesh.GetMeta().GetName()
			newEndpoints = append(newEndpoints, endpoint)
		}

		outbound[serviceName] = newEndpoints
	}

	return outbound
}

func BuildEdsEndpointMap(
	mesh *core_mesh.MeshResource,
	zone string,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	externalServices []*core_mesh.ExternalServiceResource,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}
	ingressInstances := fillIngressOutbounds(
		outbound,
		zoneIngresses,
		zoneEgresses,
		zone,
		mesh,
	)
	endpointWeight := uint32(1)
	if ingressInstances > 0 {
		endpointWeight = ingressInstances
	}
	fillDataplaneOutbounds(outbound, dataplanes, mesh, endpointWeight)

	if len(zoneEgresses) > 0 && mesh.MTLSEnabled() {
		fillExternalServicesOutboundsThroughEgress(outbound, externalServices, zoneEgresses, mesh)
	}

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
func fillDataplaneOutbounds(
	outbound core_xds.EndpointMap,
	dataplanes []*core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	endpointWeight uint32,
) {
	for _, dataplane := range dataplanes {
		dpSpec := dataplane.Spec
		dpNetworking := dpSpec.GetNetworking()

		for _, inbound := range dpNetworking.GetHealthyInbounds() {
			inboundTags := inbound.GetTags()
			serviceName := inboundTags[mesh_proto.ServiceTag]
			inboundInterface := dpNetworking.ToInboundInterface(inbound)
			inboundAddress := inboundInterface.DataplaneAdvertisedIP
			inboundPort := inboundInterface.DataplanePort

			// TODO(yskopets): do we need to dedup?
			// TODO(yskopets): sort ?
			outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
				Target:   inboundAddress,
				Port:     inboundPort,
				Tags:     inboundTags,
				Weight:   endpointWeight,
				Locality: localityFromTags(mesh, priorityLocal, inboundTags),
			})
		}
	}
}

func buildCoordinates(address string, port uint32) string {
	return net.JoinHostPort(
		address,
		strconv.FormatUint(uint64(port), 10),
	)
}

func fillIngressOutbounds(
	outbound core_xds.EndpointMap,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	zone string,
	mesh *core_mesh.MeshResource,
) uint32 {
	ziInstances := map[string]struct{}{}

	for _, zi := range zoneIngresses {
		if !zi.IsRemoteIngress(zone) {
			continue
		}

		if !mesh.MTLSEnabled() {
			// Ingress routes the request by TLS SNI, therefore for cross
			// cluster communication MTLS is required.
			// We ignore Ingress from endpoints if MTLS is disabled, otherwise
			// we would fail anyway.
			continue
		}

		if !zi.HasPublicAddress() {
			// Zone Ingress is not reachable yet from other clusters.
			// This may happen when Ingress Service is pending waiting on
			// External IP on Kubernetes.
			continue
		}

		ziNetworking := zi.Spec.GetNetworking()
		ziAddress := ziNetworking.GetAdvertisedAddress()
		ziPort := ziNetworking.GetAdvertisedPort()
		ziCoordinates := buildCoordinates(ziAddress, ziPort)

		if _, ok := ziInstances[ziCoordinates]; ok {
			// many Ingress instances can be placed in front of one load
			// balancer (all instances can have the same public address and
			// port).
			// In this case we only need one Instance avoiding creating
			// unnecessary duplicated endpoints
			continue
		}

		ziInstances[ziCoordinates] = struct{}{}

		for _, service := range zi.Spec.GetAvailableServices() {
			if service.Mesh != mesh.GetMeta().GetName() {
				continue
			}

			serviceTags := service.GetTags()
			serviceName := serviceTags[mesh_proto.ServiceTag]
			serviceInstances := service.GetInstances()
			locality := localityFromTags(mesh, priorityRemote, serviceTags)

			// TODO (bartsmykla): We have to check if it will be ok in a situation
			//  where we have few zone ingresses with the same services, as
			//  with zone egresses we will generate endpoints with the same
			//  targets and ports (zone egress ones), which envoy probably will
			//  ignore
			// If zone egresses present, we want to pass the traffic:
			// dp -> zone egress -> zone ingress -> dp
			if len(zoneEgresses) > 0 {
				for _, ze := range zoneEgresses {
					zeNetworking := ze.Spec.GetNetworking()
					zeAddress := zeNetworking.GetAddress()
					zePort := zeNetworking.GetPort()

					endpoint := core_xds.Endpoint{
						Target:   zeAddress,
						Port:     zePort,
						Tags:     serviceTags,
						Weight:   1,
						Locality: locality,
					}

					outbound[serviceName] = append(outbound[serviceName], endpoint)
				}
			} else {
				endpoint := core_xds.Endpoint{
					Target:   ziAddress,
					Port:     ziPort,
					Tags:     serviceTags,
					Weight:   serviceInstances,
					Locality: locality,
				}

				outbound[serviceName] = append(outbound[serviceName], endpoint)
			}
		}
	}

	return uint32(len(ziInstances))
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

func fillExternalServicesOutboundsThroughEgress(
	outbound core_xds.EndpointMap,
	externalServices []*core_mesh.ExternalServiceResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
) {
	for _, externalService := range externalServices {
		serviceTags := externalService.Spec.GetTags()
		serviceName := serviceTags[mesh_proto.ServiceTag]
		locality := localityFromTags(mesh, priorityRemote, serviceTags)

		for _, ze := range zoneEgresses {
			zeNetworking := ze.Spec.GetNetworking()
			zeAddress := zeNetworking.GetAddress()
			zePort := zeNetworking.GetPort()

			endpoint := core_xds.Endpoint{
				Target: zeAddress,
				Port:   zePort,
				Tags:   serviceTags,
				// AS it's a role of zone egress to load balance traffic between
				// instances, we can safely set weight to 1
				Weight:          1,
				Locality:        locality,
				ExternalService: &core_xds.ExternalService{},
			}

			outbound[serviceName] = append(outbound[serviceName], endpoint)
		}
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
