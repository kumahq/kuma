package topology

import (
	"context"
	"maps"
	"net"
	"strconv"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func BuildExternalServicesEndpointMap(
	ctx context.Context,
	mesh *core_mesh.MeshResource,
	externalServices []*core_mesh.ExternalServiceResource,
	loader datasource.Loader,
	zone string,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}
	if !mesh.ZoneEgressEnabled() {
		fillExternalServicesOutbounds(ctx, outbound, externalServices, mesh, loader, zone)
	}
	return outbound
}

// BuildEgressEndpointMap creates a map of endpoints that match given selectors
// and are not local for the provided zone (external services and services
// behind remote zone ingress only)
func BuildEgressEndpointMap(
	ctx context.Context,
	mesh *core_mesh.MeshResource,
	localZone string,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	externalServices []*core_mesh.ExternalServiceResource,
	loader datasource.Loader,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	fillIngressOutbounds(outbound, zoneIngresses, nil, localZone, mesh, nil, false)

	fillExternalServicesReachableFromZone(ctx, outbound, externalServices, mesh, loader, localZone)

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
	localZone string,
	meshServices []*meshservice_api.MeshServiceResource,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	externalServices []*core_mesh.ExternalServiceResource,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	ingressInstances := fillIngressOutbounds(outbound, zoneIngresses, zoneEgresses, localZone, mesh, nil, mesh.ZoneEgressEnabled())
	endpointWeight := uint32(1)
	if ingressInstances > 0 {
		endpointWeight = ingressInstances
	}

	fillDataplaneOutbounds(outbound, dataplanes, mesh, endpointWeight, localZone)

	fillLocalMeshServices(outbound, meshServices, dataplanes, mesh, endpointWeight, localZone)

	if mesh.ZoneEgressEnabled() {
		fillExternalServicesOutboundsThroughEgress(outbound, externalServices, zoneEgresses, mesh, localZone)
	}

	return outbound
}

// endpointWeight defines default weight for in-cluster endpoint.
// Examples of having service "backend":
//  1. Single-zone deployment, 2 instances in one cluster (zone1)
//     All endpoints have to have the same weight (ex. 1) to achieve fair loadbalancing.
//     Endpoints:
//     * backend-zone1-1 - weight: 1
//     * backend-zone1-2 - weight: 1
//  2. Multi-zone deployment, 2 instances in "zone1" (local zone), 3 instances in "zone2" (remote zone) with 1 Ingress instance
//     Endpoints:
//     * backend-zone1-1 - weight: 1
//     * backend-zone1-2 - weight: 1
//     * ingress-zone2-1 - weight: 3 (all remote endpoints are aggregated to one Ingress, it needs to have weight of instances in other cluster)
//  3. Multi-zone deployment, 2 instances in "zone1" (local zone), 2 instances in "zone2" (remote zone) with 1 Ingress instance
//     Many instances of Ingress will forward the traffic to the same endpoints in "zone2" so we need to lower the weights.
//     Since weights are integers, we cannot put fractional on ingress endpoints weights, we need to adjust "default" weight for local zone
//     Endpoints:
//     * backend-zone1-1 - weight: 2
//     * backend-zone1-2 - weight: 2
//     * ingress-zone2-1 - weight: 3
//     * ingress-zone2-2 - weight: 3
func fillDataplaneOutbounds(
	outbound core_xds.EndpointMap,
	dataplanes []*core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	endpointWeight uint32,
	localZone string,
) {
	for _, dataplane := range dataplanes {
		dpSpec := dataplane.Spec
		dpNetworking := dpSpec.GetNetworking()

		for _, inbound := range dpNetworking.GetHealthyInbounds() {
			inboundTags := maps.Clone(inbound.GetTags())
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
				Locality: GetLocality(localZone, getZone(inboundTags), mesh.LocalityAwareLbEnabled()),
			})
		}
	}
}

func fillLocalMeshServices(
	outbound core_xds.EndpointMap,
	meshServices []*meshservice_api.MeshServiceResource,
	dataplanes []*core_mesh.DataplaneResource,
	mesh *core_mesh.MeshResource,
	endpointWeight uint32,
	localZone string,
) {
	// O(dataplane*meshsvc) can be optimized by sharding both by namespace
	for _, dataplane := range dataplanes {
		dpNetworking := dataplane.Spec.GetNetworking()

		for _, meshSvc := range meshServices {
			tagSelector := mesh_proto.TagSelector(meshSvc.Spec.Selector.DataplaneTags)

			for _, inbound := range dpNetworking.GetHealthyInbounds() {
				if !tagSelector.Matches(inbound.GetTags()) {
					continue
				}
				for _, port := range meshSvc.Spec.Ports {
					if port.TargetPort != inbound.Port {
						continue
					}

					inboundTags := maps.Clone(inbound.GetTags())
					serviceName := meshSvc.DestinationName(port.Port)
					if serviceName == inboundTags[mesh_proto.ServiceTag] {
						continue // it was already added by fillDataplaneOutbounds
					}
					inboundInterface := dpNetworking.ToInboundInterface(inbound)

					outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
						Target:   inboundInterface.DataplaneAdvertisedIP,
						Port:     inboundInterface.DataplanePort,
						Tags:     inboundTags,
						Weight:   endpointWeight,
						Locality: GetLocality(localZone, getZone(inboundTags), mesh.LocalityAwareLbEnabled()),
					})
				}
			}
		}
	}
}

func CrossMeshEndpointTags(
	gateways []*core_mesh.MeshGatewayResource,
	dataplanes []*core_mesh.DataplaneResource,
) []mesh_proto.SingleValueTagSet {
	endpoints := []mesh_proto.SingleValueTagSet{}

	for _, dataplane := range dataplanes {
		if !dataplane.Spec.IsBuiltinGateway() {
			continue
		}

		gateway := SelectGateway(gateways, dataplane.Spec.Matches)
		if gateway == nil {
			continue
		}

		dpSpec := dataplane.Spec
		dpNetworking := dpSpec.GetNetworking()

		dpGateway := dpNetworking.GetGateway()
		dpTags := dpGateway.GetTags()

		for _, listener := range gateway.Spec.GetConf().GetListeners() {
			if !listener.CrossMesh {
				continue
			}
			endpoints = append(endpoints, mesh_proto.Merge(dpTags, gateway.Spec.GetTags(), listener.GetTags()))
		}
	}

	return endpoints
}

func BuildCrossMeshEndpointMap(
	mesh *core_mesh.MeshResource,
	otherMesh *core_mesh.MeshResource,
	localZone string,
	gateways []*core_mesh.MeshGatewayResource,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	ingressInstances := fillIngressOutbounds(
		outbound,
		zoneIngresses,
		zoneEgresses,
		localZone,
		mesh,
		otherMesh,
		mesh.ZoneEgressEnabled(),
	)

	endpointWeight := uint32(1)
	if ingressInstances > 0 {
		endpointWeight = ingressInstances
	}

	for _, dataplane := range dataplanes {
		if !dataplane.Spec.IsBuiltinGateway() {
			continue
		}

		gateway := SelectGateway(gateways, dataplane.Spec.Matches)
		if gateway == nil {
			continue
		}

		dpSpec := dataplane.Spec
		dpNetworking := dpSpec.GetNetworking()

		dpGateway := dpNetworking.GetGateway()
		dpTags := dpGateway.GetTags()
		serviceName := dpTags[mesh_proto.ServiceTag]

		for _, listener := range gateway.Spec.GetConf().GetListeners() {
			if !listener.CrossMesh {
				continue
			}

			outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
				Target:   dpNetworking.GetAddress(),
				Port:     listener.GetPort(),
				Tags:     mesh_proto.Merge(dpTags, gateway.Spec.GetTags(), listener.GetTags()),
				Weight:   endpointWeight,
				Locality: GetLocality(localZone, getZone(dpTags), mesh.LocalityAwareLbEnabled()),
			})
		}
	}

	return outbound
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
	localZone string,
	mesh *core_mesh.MeshResource,
	otherMesh *core_mesh.MeshResource, // otherMesh is set if we are looking for specific crossmesh connections
	routeThroughZoneEgress bool,
) uint32 {
	ziInstances := map[string]struct{}{}

	for _, zi := range zoneIngresses {
		if !zi.IsRemoteIngress(localZone) {
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

		if len(zi.Spec.GetAvailableServices()) > 0 {
			// Consider squashing instances only if available services are reconciled.
			// This is necessary to perform graceful join of new instances of ZoneIngress.
			// When a new instance of ZoneIngress is up, it's immediately synced to all zones.
			// Only after a short period of time (first XDS reconciliation) it will be updated with AvailableServices.
			// If we just take any instance, we may choose an instance without AvailableServices as a representation
			// of all ZoneIngress instances behind one load balancer.
			ziInstances[ziCoordinates] = struct{}{}
		}

		for _, service := range zi.Spec.GetAvailableServices() {
			relevantMesh := mesh
			if otherMesh != nil {
				relevantMesh = otherMesh
			}

			if service.Mesh != relevantMesh.GetMeta().GetName() {
				continue
			}
			if _, hasMeshTag := service.Tags[mesh_proto.MeshTag]; otherMesh != nil && !hasMeshTag {
				// If the mesh tag is set, we assume the service should be available
				// crossMesh but only if we're looking for this other mesh
				continue
			}

			// deep copy map to not modify tags in BuildRemoteEndpointMap
			serviceTags := maps.Clone(service.GetTags())
			serviceName := serviceTags[mesh_proto.ServiceTag]
			serviceInstances := service.GetInstances()
			locality := GetLocality(localZone, getZone(serviceTags), mesh.LocalityAwareLbEnabled())

			// TODO (bartsmykla): We have to check if it will be ok in a situation
			//  where we have few zone ingresses with the same services, as
			//  with zone egresses we will generate endpoints with the same
			//  targets and ports (zone egress ones), which envoy probably will
			//  ignore
			// If zone egresses present, we want to pass the traffic:
			// dp -> zone egress -> zone ingress -> dp
			if routeThroughZoneEgress {
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
					// this is necessary for correct spiffe generation for dp when
					// traffic is routed: egress -> ingress -> egress
					if service.ExternalService {
						endpoint.ExternalService = &core_xds.ExternalService{}
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

func fillExternalServicesReachableFromZone(
	ctx context.Context,
	outbound core_xds.EndpointMap,
	externalServices []*core_mesh.ExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
	zone string,
) {
	for _, externalService := range externalServices {
		if externalService.IsReachableFromZone(zone) {
			createExternalServiceEndpoint(ctx, outbound, externalService, mesh, loader, zone)
		}
	}
}

func fillExternalServicesOutbounds(
	ctx context.Context,
	outbound core_xds.EndpointMap,
	externalServices []*core_mesh.ExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
	zone string,
) {
	for _, externalService := range externalServices {
		createExternalServiceEndpoint(ctx, outbound, externalService, mesh, loader, zone)
	}
}

func createExternalServiceEndpoint(
	ctx context.Context,
	outbound core_xds.EndpointMap,
	externalService *core_mesh.ExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
	zone string,
) {
	service := externalService.Spec.GetService()
	externalServiceEndpoint, err := NewExternalServiceEndpoint(ctx, externalService, mesh, loader, zone)
	if err != nil {
		core.Log.Error(err, "unable to create ExternalService endpoint. Endpoint won't be included in the XDS.", "name", externalService.Meta.GetName(), "mesh", externalService.Meta.GetMesh())
		return
	}
	outbound[service] = append(outbound[service], *externalServiceEndpoint)
}

func fillExternalServicesOutboundsThroughEgress(
	outbound core_xds.EndpointMap,
	externalServices []*core_mesh.ExternalServiceResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
	localZone string,
) {
	for _, externalService := range externalServices {
		// deep copy map to not modify tags in ExternalService.
		serviceTags := maps.Clone(externalService.Spec.GetTags())
		serviceName := serviceTags[mesh_proto.ServiceTag]
		locality := GetLocality(localZone, getZone(serviceTags), mesh.LocalityAwareLbEnabled())

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
	ctx context.Context,
	externalService *core_mesh.ExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
	zone string,
) (*core_xds.Endpoint, error) {
	spec := externalService.Spec
	tls := spec.GetNetworking().GetTls()
	meshName := mesh.GetMeta().GetName()
	// deep copy map to not modify tags in ExternalService.
	tags := maps.Clone(spec.GetTags())

	caCert, err := loadBytes(ctx, tls.GetCaCert(), meshName, loader)
	if err != nil {
		return nil, errors.Wrap(err, "could not load caCert")
	}
	clientCert, err := loadBytes(ctx, tls.GetClientCert(), meshName, loader)
	if err != nil {
		return nil, errors.Wrap(err, "could not load clientCert")
	}
	clientKey, err := loadBytes(ctx, tls.GetClientKey(), meshName, loader)
	if err != nil {
		return nil, errors.Wrap(err, "could not load clientKey")
	}

	es := &core_xds.ExternalService{
		TLSEnabled:               tls.GetEnabled(),
		CaCert:                   caCert,
		ClientCert:               clientCert,
		ClientKey:                clientKey,
		AllowRenegotiation:       tls.GetAllowRenegotiation().GetValue(),
		SkipHostnameVerification: tls.GetSkipHostnameVerification().GetValue(),
		ServerName:               tls.GetServerName().GetValue(),
	}

	if es.TLSEnabled {
		name := externalService.Meta.GetName()

		tags = envoy_tags.Tags(tags).
			WithTags(mesh_proto.ExternalServiceTag, name)
	}

	return &core_xds.Endpoint{
		Target:          spec.GetHost(),
		Port:            spec.GetPortUInt32(),
		Tags:            tags,
		Weight:          1,
		ExternalService: es,
		Locality:        GetLocality(zone, getZone(tags), mesh.LocalityAwareLbEnabled()),
	}, nil
}

func loadBytes(ctx context.Context, ds *v1alpha1.DataSource, mesh string, loader datasource.Loader) ([]byte, error) {
	if ds == nil {
		return nil, nil
	}
	return loader.Load(ctx, mesh, ds)
}

const (
	// Constants for Locality Aware load balancing
	// The Highest priority 0 shall be assigned to all locally available services
	// A priority of 1 is for ExternalServices and services exposed on neighboring ingress-es
	priorityLocal  = 0
	priorityRemote = 1
)

func GetLocality(localZone string, otherZone *string, localityAwareness bool) *core_xds.Locality {
	if otherZone == nil {
		return nil
	}

	// we want to set the Locality even when localityAwareLoadBalancing is not enabled
	// If we set the locality we have an extra visibility about this in /clusters etc.
	// Kuma's LocalityAwareLoadBalancing feature is based only on Priority therefore when
	// it's disabled we need to set Priority to local.
	//
	// Setting this regardless of LocalityAwareLoadBalancing on the mesh also solves
	// the problem that endpoints have problems when moving from one locality to another
	// https://github.com/envoyproxy/envoy/issues/12392

	var priority uint32 = priorityLocal

	if localityAwareness && localZone != *otherZone {
		priority = priorityRemote
	}

	return &core_xds.Locality{
		Zone:     *otherZone,
		Priority: priority,
	}
}

func getZone(tags map[string]string) *string {
	if zone, ok := tags[mesh_proto.ZoneTag]; ok {
		return &zone
	}
	return nil
}
