package topology

import (
	"context"
	"maps"
	"net"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/datasource"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

var outboundLog = core.Log.WithName("xds").WithName("outbound")

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
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	loader datasource.Loader,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	fillIngressOutbounds(outbound, zoneIngresses, nil, localZone, mesh, nil, false, map[core_xds.ServiceName]struct{}{})

	fillExternalServicesReachableFromZone(ctx, outbound, externalServices, meshExternalServices, mesh, loader, localZone)

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

func BuildIngressEndpointMap(
	ctx context.Context,
	mesh *core_mesh.MeshResource,
	localZone string,
	meshServices []*meshservice_api.MeshServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	dataplanes []*core_mesh.DataplaneResource,
	externalServices []*core_mesh.ExternalServiceResource,
	gateways []*core_mesh.MeshGatewayResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	loader datasource.Loader,
	mtlsEnabled bool,
) core_xds.EndpointMap {
	// Build EDS endpoint map just like for regular DPP, but without list of Ingress.
	// This way we only keep local endpoints.
	outbound := BuildEdsEndpointMap(ctx, mesh, localZone, meshServices, meshMultiZoneServices, meshExternalServices, dataplanes, nil, zoneEgresses, externalServices, loader, mtlsEnabled)
	fillLocalCrossMeshOutbounds(outbound, mesh, dataplanes, gateways, 1, localZone)
	return outbound
}

func BuildEdsEndpointMap(
	ctx context.Context,
	mesh *core_mesh.MeshResource,
	localZone string,
	meshServices []*meshservice_api.MeshServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	externalServices []*core_mesh.ExternalServiceResource,
	loader datasource.Loader,
	mtlsEnabled bool,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	meshServicesByKri := make(map[kri.Identifier]*meshservice_api.MeshServiceResource, len(meshServices))
	for _, ms := range meshServices {
		meshServicesByKri[kri.From(ms)] = ms
	}

	fillLocalMeshServices(outbound, meshServices, dataplanes, mesh, localZone)
	// we want to prefer endpoints build by MeshService
	// this way we can for example stop cross-zone traffic by default using kuma.io/service
	meshServiceDestinations := map[core_xds.ServiceName]struct{}{}
	for name := range outbound {
		meshServiceDestinations[name] = struct{}{}
	}

	ingressInstances := fillIngressOutbounds(outbound, zoneIngresses, zoneEgresses, localZone, mesh, nil, mesh.ZoneEgressEnabled(), meshServiceDestinations)
	endpointWeight := uint32(1)
	if ingressInstances > 0 {
		endpointWeight = ingressInstances
	}

	fillDataplaneOutbounds(outbound, dataplanes, mesh, endpointWeight, localZone, meshServiceDestinations)

	fillRemoteMeshServices(outbound, meshServices, zoneIngresses, mesh, localZone, mtlsEnabled)

	fillExternalServicesOutboundsThroughEgress(ctx, outbound, externalServices, meshExternalServices, zoneEgresses, mesh, localZone, loader)

	// it has to be last because it reuses endpoints for other cases
	fillMeshMultiZoneServices(outbound, meshServicesByKri, meshMultiZoneServices)

	return outbound
}

func fillMeshMultiZoneServices(
	outbound core_xds.EndpointMap,
	meshServicesByName map[kri.Identifier]*meshservice_api.MeshServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
) {
	for _, mzSvc := range meshMultiZoneServices {
		for _, matchedMs := range mzSvc.Status.MeshServices {
			ri := kri.Identifier{
				ResourceType: meshservice_api.MeshServiceType,
				Mesh:         matchedMs.Mesh,
				Zone:         matchedMs.Zone,
				Namespace:    matchedMs.Namespace,
				Name:         matchedMs.Name,
			}
			ms, ok := meshServicesByName[ri]
			if !ok {
				continue
			}
			if !ms.IsLocalMeshService() && ms.Spec.State != meshservice_api.StateAvailable {
				// we don't want to load balance to zones that has no available endpoints.
				// we check this only for non-local services, because if service is unavailable in the local zone it has no endpoints.
				// if a new local endpoint just become healthy, we can add it immediately without waiting for state to be reconciled.
				continue
			}
			for _, port := range mzSvc.Spec.Ports {
				serviceName := destinationname.MustResolve(false, mzSvc, port)

				existingEndpoints := outbound[destinationname.MustResolve(false, ms, port)]
				outbound[serviceName] = append(outbound[serviceName], existingEndpoints...)
			}
		}
	}
}

func fillRemoteMeshServices(
	outbound core_xds.EndpointMap,
	services []*meshservice_api.MeshServiceResource,
	zoneIngress []*core_mesh.ZoneIngressResource,
	mesh *core_mesh.MeshResource,
	localZone string,
	mtlsEnabled bool,
) {
	if !mtlsEnabled {
		return
	}

	ziInstances := map[string]struct{}{}

	// introduction of MeshIdentity doesn't requires mTLS on mesh
	zoneToEndpoints := map[string][]core_xds.Endpoint{}
	for _, zi := range zoneIngress {
		if !zi.IsRemoteIngress(localZone) {
			continue
		}

		ziAddress := zi.Spec.GetNetworking().GetAdvertisedAddress()
		ziPort := zi.Spec.GetNetworking().GetAdvertisedPort()
		ziCoordinates := buildCoordinates(ziAddress, ziPort)

		if _, ok := ziInstances[ziCoordinates]; ok {
			// many Ingress instances can be placed in front of one load
			// balancer (all instances can have the same public address and
			// port).
			// In this case we only need one Instance avoiding creating
			// unnecessary duplicated endpoints
			continue
		}

		zoneToEndpoints[zi.Spec.Zone] = append(zoneToEndpoints[zi.Spec.Zone], core_xds.Endpoint{
			Target:   ziAddress,
			Port:     ziPort,
			Tags:     nil,
			Weight:   1,
			Locality: GetLocality(localZone, &zi.Spec.Zone, mesh.LocalityAwareLbEnabled()),
		})
	}

	for _, ms := range services {
		if ms.IsLocalMeshService() {
			continue
		}
		msZone := ms.GetMeta().GetLabels()[mesh_proto.ZoneTag]
		for _, port := range ms.Spec.Ports {
			serviceName := destinationname.MustResolve(false, ms, port)
			for _, endpoint := range zoneToEndpoints[msZone] {
				ep := endpoint
				ep.Locality = &core_xds.Locality{
					Zone:     msZone,
					Priority: priorityRemote,
				}
				ep.Tags = map[string]string{
					mesh_proto.ServiceTag: serviceName,
					mesh_proto.ZoneTag:    msZone,
				}
				outbound[serviceName] = append(outbound[serviceName], ep)
			}
		}
	}
}

type MeshServiceIdentity struct {
	Resource   *meshservice_api.MeshServiceResource
	Identities map[string]struct{}
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
	meshServiceDestinations map[core_xds.ServiceName]struct{},
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

			if _, ok := meshServiceDestinations[serviceName]; ok {
				continue
			}

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
	localZone string,
) {
	dppsForMs := meshservice.MatchDataplanesWithMeshServices(dataplanes, meshServices, true)
	for meshSvc, dpps := range dppsForMs {
		if !meshSvc.IsLocalMeshService() {
			continue
		}

		for _, dpp := range dpps {
			dpNetworking := dpp.Spec.GetNetworking()
			for _, inbound := range dpNetworking.GetHealthyInbounds() {
				for _, port := range meshSvc.Spec.Ports {
					if !meshservice.MatchInboundWithMeshServicePort(inbound, port) {
						continue
					}

					inboundTags := maps.Clone(inbound.GetTags())
					serviceName := destinationname.MustResolve(false, meshSvc, port)
					inboundInterface := dpNetworking.ToInboundInterface(inbound)

					outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
						Target:   inboundInterface.DataplaneAdvertisedIP,
						Port:     inboundInterface.DataplanePort,
						Tags:     inboundTags,
						Weight:   1,
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
		map[core_xds.ServiceName]struct{}{},
	)

	endpointWeight := uint32(1)
	if ingressInstances > 0 {
		endpointWeight = ingressInstances
	}
	fillLocalCrossMeshOutbounds(outbound, mesh, dataplanes, gateways, endpointWeight, localZone)
	return outbound
}

func fillLocalCrossMeshOutbounds(
	outbound core_xds.EndpointMap,
	mesh *core_mesh.MeshResource,
	dataplanes []*core_mesh.DataplaneResource,
	gateways []*core_mesh.MeshGatewayResource,
	endpointWeight uint32,
	localZone string,
) {
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
	meshServiceDestinations map[core_xds.ServiceName]struct{},
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

			if _, ok := meshServiceDestinations[serviceName]; ok {
				continue
			}

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
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
	zone string,
) {
	for _, externalService := range externalServices {
		if externalService.IsReachableFromZone(zone) {
			createExternalServiceEndpoint(ctx, outbound, externalService, mesh, loader, zone)
		}
	}
	for _, mes := range meshExternalServices {
		if mes.IsReachableFromZone(zone) {
			err := createMeshExternalServiceEndpoint(ctx, outbound, mes, mesh, loader, zone)
			if err != nil {
				outboundLog.Error(err, "unable to create MeshExternalService endpoint. Endpoint won't be included in the XDS.", "name", mes.Meta.GetName(), "mesh", mes.Meta.GetMesh())
				continue
			}
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

func createMeshExternalServiceEndpoint(
	ctx context.Context,
	outbounds core_xds.EndpointMap,
	mes *meshexternalservice_api.MeshExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
	zone string,
) error {
	es := &core_xds.ExternalService{
		Protocol:      mes.Spec.Match.Protocol,
		OwnerResource: kri.From(mes),
	}
	tags := maps.Clone(mes.Meta.GetLabels())
	if tags == nil {
		tags = map[string]string{}
	}
	meshName := mesh.GetMeta().GetName()
	tls := mes.Spec.Tls
	if tls != nil && tls.Enabled {
		err := setTlsConfiguration(ctx, tls, es, meshName, loader)
		if err != nil {
			return err
		}
	}

	// if all ip make it static - it's done in endpoint_cluster_configurer
	for i, endpoint := range pointer.Deref(mes.Spec.Endpoints) {
		if i == 0 && es.ServerName == "" && govalidator.IsDNSName(endpoint.Address) && tls != nil && tls.Enabled {
			es.ServerName = endpoint.Address
		}
		outboundEndpoint := &core_xds.Endpoint{
			Target:          endpoint.Address,
			Port:            uint32(endpoint.Port),
			Weight:          1,
			ExternalService: es,
			Tags:            tags,
			Locality:        GetLocality(zone, getZone(tags), mesh.LocalityAwareLbEnabled()),
		}
		legacyName := destinationname.MustResolve(false, mes, mes.Spec.Match)
		outbounds[legacyName] = append(outbounds[legacyName], *outboundEndpoint)
	}
	return nil
}

func setTlsConfiguration(ctx context.Context, tls *meshexternalservice_api.Tls, es *core_xds.ExternalService, meshName string, loader datasource.Loader) error {
	var caCert, clientCert, clientKey []byte
	es.TLSEnabled = tls.Enabled
	es.FallbackToSystemCa = true
	es.AllowRenegotiation = tls.AllowRenegotiation

	if tls.Version != nil {
		if tls.Version.Min != nil {
			es.MinTlsVersion = pointer.To(common_tls.ToTlsVersion(tls.Version.Min))
		}
		if tls.Version.Max != nil {
			es.MaxTlsVersion = pointer.To(common_tls.ToTlsVersion(tls.Version.Max))
		}
	}
	var err error
	if tls.Verification != nil {
		if tls.Verification.CaCert != nil {
			caCert, err = loadBytes(ctx, tls.Verification.CaCert.ConvertToProto(), meshName, loader)
			if err != nil {
				return errors.Wrap(err, "could not load caCert")
			}
			es.CaCert = caCert
		}
		if tls.Verification.ClientKey != nil && tls.Verification.ClientCert != nil {
			clientCert, err = loadBytes(ctx, tls.Verification.ClientCert.ConvertToProto(), meshName, loader)
			if err != nil {
				return errors.Wrap(err, "could not load clientCert")
			}
			clientKey, err = loadBytes(ctx, tls.Verification.ClientKey.ConvertToProto(), meshName, loader)
			if err != nil {
				return errors.Wrap(err, "could not load clientKey")
			}
			es.ClientCert = clientCert
			es.ClientKey = clientKey
		}
		if pointer.Deref(tls.Verification.ServerName) != "" {
			es.ServerName = pointer.Deref(tls.Verification.ServerName)
		}
		for _, san := range pointer.Deref(tls.Verification.SubjectAltNames) {
			es.SANs = append(es.SANs, core_xds.SAN{
				MatchType: core_xds.MatchType(san.Type),
				Value:     san.Value,
			})
		}
		// Server name and SNI we need to add
		// mes.Spec.Tls.Verification.SubjectAltNames
		switch tls.Verification.Mode {
		case meshexternalservice_api.TLSVerificationSkipSAN:
			es.ServerName = ""
			es.SANs = []core_xds.SAN{}
			es.SkipHostnameVerification = true
		case meshexternalservice_api.TLSVerificationSkipCA:
			es.CaCert = nil
			es.FallbackToSystemCa = false
		case meshexternalservice_api.TLSVerificationSkipAll:
			es.FallbackToSystemCa = false
			es.CaCert = nil
			es.ClientKey = nil
			es.ClientCert = nil
			es.ServerName = ""
			es.SANs = []core_xds.SAN{}
			es.SkipHostnameVerification = true
		}
	}
	return nil
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
		outboundLog.Error(err, "unable to create ExternalService endpoint. Endpoint won't be included in the XDS.", "name", externalService.Meta.GetName(), "mesh", externalService.Meta.GetMesh())
		return
	}
	outbound[service] = append(outbound[service], *externalServiceEndpoint)
}

func fillExternalServicesOutboundsThroughEgress(
	ctx context.Context,
	outbound core_xds.EndpointMap,
	externalServices []*core_mesh.ExternalServiceResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	zoneEgresses []*core_mesh.ZoneEgressResource,
	mesh *core_mesh.MeshResource,
	localZone string,
	loader datasource.Loader,
) {
	if mesh.ZoneEgressEnabled() {
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
	for _, mes := range meshExternalServices {
		// deep copy map to not modify tags in ExternalService.
		serviceTags := maps.Clone(mes.Meta.GetLabels())
		serviceName := destinationname.MustResolve(false, mes, mes.Spec.Match)
		locality := GetLocality(localZone, getZone(serviceTags), mesh.LocalityAwareLbEnabled())
		tls := mes.Spec.Tls
		es := &core_xds.ExternalService{
			Protocol:      mes.Spec.Match.Protocol,
			OwnerResource: kri.From(mes),
		}
		if tls != nil && tls.Enabled {
			err := setTlsConfiguration(ctx, tls, es, mes.Meta.GetMesh(), loader)
			if err != nil {
				outboundLog.Error(err, "unable to create MeshExternalService endpoint for egress. Endpoint won't be included in the XDS.", "name", mes.Meta.GetName(), "mesh", mes.Meta.GetMesh())
				continue
			}
		}

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
				ExternalService: es,
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
