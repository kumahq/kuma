package topology

import (
	"context"
	"maps"
	"net"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	datasource_api "github.com/kumahq/kuma/v3/api/common/v1alpha1/datasource"
	common_tls "github.com/kumahq/kuma/v3/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/datasource"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/destinationname"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	meshzoneaddress_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshzoneaddress/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var outboundLog = core.Log.WithName("xds").WithName("outbound")

// A Dataplane gets one endpoint map per listener kind it exposes, because the same
// destination resolves to different endpoints depending on who is routing to it:
//
//   - BuildDataplaneEndpointMap      - a regular proxy's outbounds
//   - BuildDataplaneZoneIngressEndpointMap - an embedded zone ingress, terminating traffic
//     from other zones and forwarding it to local workloads
//   - BuildDataplaneZoneEgressEndpointMap  - an embedded zone egress, terminating traffic
//     from the local zone and forwarding it outside the mesh
//
// Every destination is backed by a real resource (MeshService, MeshMultiZoneService,
// MeshExternalService); kuma.io/service is not a source of endpoints.

// BuildDataplaneEndpointMap builds the endpoints a regular Dataplane routes to: local and
// remote MeshServices, MeshMultiZoneServices, and MeshExternalServices reached through a
// zone egress.
func BuildDataplaneEndpointMap(
	ctx context.Context,
	localZone string,
	meshServices []*meshservice_api.MeshServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	dataplanes []*core_mesh.DataplaneResource,
	meshZoneAddresses []*meshzoneaddress_api.MeshZoneAddressResource,
	loader datasource.Loader,
	mtlsEnabled bool,
	egressAddresses []core_xds.ZoneEgressInstance,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	fillLocalMeshServices(outbound, meshServices, dataplanes)
	fillRemoteMeshServices(outbound, meshServices, meshZoneAddresses, localZone, mtlsEnabled)
	fillMeshExternalServicesOnDataplane(ctx, outbound, meshExternalServices, egressAddresses, loader)
	// has to be last, it republishes the endpoints the fillers above produced
	fillMeshMultiZoneServices(outbound, meshServices, meshMultiZoneServices)

	return outbound
}

// BuildDataplaneZoneIngressEndpointMap builds the endpoints an embedded zone ingress listener
// forwards to: local MeshServices and the MeshMultiZoneServices covering them. A zone ingress
// only ever terminates traffic destined for its own zone, so remote and external destinations
// are absent by design.
func BuildDataplaneZoneIngressEndpointMap(
	meshServices []*meshservice_api.MeshServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
	dataplanes []*core_mesh.DataplaneResource,
) core_xds.EndpointMap {
	outbound := core_xds.EndpointMap{}

	fillLocalMeshServices(outbound, meshServices, dataplanes)
	// has to be last, it republishes the endpoints the filler above produced
	fillMeshMultiZoneServices(outbound, meshServices, meshMultiZoneServices)

	return outbound
}

// BuildDataplaneZoneEgressEndpointMap builds the endpoints an embedded zone egress listener
// forwards to: the real addresses of every MeshExternalService. This is the far side of the
// hop that BuildDataplaneEndpointMap points at the egress itself.
//
// The result carries per-destination protocol and owner alongside the endpoints, which the
// egress listener needs to pick a filter chain. Always uses unified (KRI) naming, as this is
// new infrastructure that only supports Exclusive MeshServices mode.
func BuildDataplaneZoneEgressEndpointMap(
	ctx context.Context,
	mesh *core_mesh.MeshResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	loader datasource.Loader,
) core_xds.EgressEndpointMap {
	outbound := core_xds.EgressEndpointMap{}
	fillMeshExternalServicesOnEgress(ctx, outbound, meshExternalServices, mesh, loader)
	return outbound
}

// fillMeshMultiZoneServices publishes, under each MeshMultiZoneService destination, the
// endpoints already built for the MeshServices it matched. It reuses what the other fillers
// produced, so it must run after them.
func fillMeshMultiZoneServices(
	outbound core_xds.EndpointMap,
	meshServices []*meshservice_api.MeshServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
) {
	if len(meshMultiZoneServices) == 0 {
		return
	}

	meshServicesByName := make(map[kri.Identifier]*meshservice_api.MeshServiceResource, len(meshServices))
	for _, ms := range meshServices {
		meshServicesByName[kri.From(ms)] = ms
	}

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
				serviceName := destinationname.ResolveLegacyFromDestination(mzSvc, port)

				existingEndpoints := outbound[destinationname.ResolveLegacyFromDestination(ms, port)]
				outbound[serviceName] = append(outbound[serviceName], existingEndpoints...)
			}
		}
	}
}

func fillRemoteMeshServices(
	outbound core_xds.EndpointMap,
	services []*meshservice_api.MeshServiceResource,
	meshZoneAddresses []*meshzoneaddress_api.MeshZoneAddressResource,
	localZone string,
	mtlsEnabled bool,
) {
	if !mtlsEnabled {
		return
	}

	// introduction of MeshIdentity doesn't requires mTLS on mesh
	zoneToEndpoints := map[string][]core_xds.Endpoint{}

	// MeshZoneAddress is the only source of publicly reachable coordinates of a
	// remote zone proxy. On Kubernetes it's reconciled from the zone ingress
	// Service, on Universal it's authored by the user.
	type zoneCoordinates struct {
		zone        string
		coordinates string
	}
	mzaInstances := map[zoneCoordinates]struct{}{}
	for _, mza := range meshZoneAddresses {
		zone := mza.GetMeta().GetLabels()[mesh_proto.ZoneTag]
		if zone == "" || zone == localZone {
			continue
		}
		// many zone proxy instances can be placed behind one load balancer, in
		// which case they share a public address and port. Deduplicate per zone
		// to avoid unnecessary duplicated endpoints, but never across zones.
		key := zoneCoordinates{zone: zone, coordinates: buildCoordinates(mza.Spec.Address, uint32(mza.Spec.Port))}
		if _, ok := mzaInstances[key]; ok {
			continue
		}
		mzaInstances[key] = struct{}{}
		zoneToEndpoints[zone] = append(zoneToEndpoints[zone], core_xds.Endpoint{
			Target: mza.Spec.Address,
			Port:   uint32(mza.Spec.Port),
			Tags:   nil,
			Weight: 1,
		})
	}

	unreachableZones := map[string]struct{}{}
	for _, ms := range services {
		if ms.IsLocalMeshService() {
			continue
		}
		msZone := ms.GetMeta().GetLabels()[mesh_proto.ZoneTag]
		if len(zoneToEndpoints[msZone]) == 0 {
			if _, reported := unreachableZones[msZone]; !reported {
				unreachableZones[msZone] = struct{}{}
				outboundLog.Info("no MeshZoneAddress found for zone, MeshService destinations in that zone get no endpoints",
					"zone", msZone, "mesh", ms.GetMeta().GetMesh())
			}
			continue
		}
		for _, port := range ms.Spec.Ports {
			serviceName := destinationname.ResolveLegacyFromDestination(ms, port)
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

// endpointIdentity returns the tags that make up an endpoint's load-balancing
// identity, sourced from the Dataplane's own resource labels. The inbound's
// protocol is carried alongside them because service-level protocol inference
// (MeshContext.GetServiceProtocol) reads it off the endpoint, and it is a
// per-port property that resource labels cannot express.
func endpointIdentity(dataplane *core_mesh.DataplaneResource, inbound *mesh_proto.Dataplane_Networking_Inbound) map[string]string {
	tags := maps.Clone(dataplane.GetMeta().GetLabels())
	if tags == nil {
		tags = map[string]string{}
	}
	if protocol := inbound.GetProtocol(); protocol != "" {
		tags[mesh_proto.ProtocolTag] = protocol
	}
	return tags
}

// fillLocalMeshServices adds one endpoint per healthy inbound backing a MeshService of this
// zone, pointing straight at the workload.
func fillLocalMeshServices(
	outbound core_xds.EndpointMap,
	meshServices []*meshservice_api.MeshServiceResource,
	dataplanes []*core_mesh.DataplaneResource,
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

					inboundTags := endpointIdentity(dpp, inbound)
					serviceName := destinationname.ResolveLegacyFromDestination(meshSvc, port)
					inboundInterface := dpNetworking.ToInboundInterface(inbound)

					outbound[serviceName] = append(outbound[serviceName], core_xds.Endpoint{
						Target:   inboundInterface.DataplaneIP,
						Port:     inboundInterface.DataplanePort,
						Tags:     inboundTags,
						Weight:   1,
						Locality: GetLocality(getZone(inboundTags)),
					})
				}
			}
		}
	}
}

func buildCoordinates(address string, port uint32) string {
	return net.JoinHostPort(
		address,
		strconv.FormatUint(uint64(port), 10),
	)
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
			caCert, err = loadSecureBytes(ctx, tls.Verification.CaCert, meshName, loader)
			if err != nil {
				return errors.Wrap(err, "could not load caCert")
			}
			es.CaCert = caCert
		}
		if tls.Verification.ClientKey != nil && tls.Verification.ClientCert != nil {
			clientCert, err = loadSecureBytes(ctx, tls.Verification.ClientCert, meshName, loader)
			if err != nil {
				return errors.Wrap(err, "could not load clientCert")
			}
			clientKey, err = loadSecureBytes(ctx, tls.Verification.ClientKey, meshName, loader)
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

// fillMeshExternalServicesOnDataplane points every MeshExternalService destination at the zone
// egress instances rather than at the external address itself: a regular proxy always takes the
// egress hop, and the egress resolves the real endpoints (see fillMeshExternalServicesOnEgress).
// With no egress instances the destination is left without endpoints.
func fillMeshExternalServicesOnDataplane(
	ctx context.Context,
	outbound core_xds.EndpointMap,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	egressAddresses []core_xds.ZoneEgressInstance,
	loader datasource.Loader,
) {
	for _, mes := range meshExternalServices {
		// deep copy map to not modify tags in ExternalService.
		serviceTags := maps.Clone(mes.Meta.GetLabels())
		serviceName := destinationname.ResolveLegacyFromDestination(mes, mes.Spec.Match)
		locality := GetLocality(nil)
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

		for _, ze := range egressAddresses {
			endpoint := core_xds.Endpoint{
				Target: ze.Address,
				Port:   ze.Port,
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

// fillMeshExternalServicesOnEgress resolves every MeshExternalService to its real endpoints,
// one per entry in spec.endpoints. This is the egress's own view of the destination; a
// MeshExternalService whose TLS config cannot be loaded is logged and skipped so one bad
// resource does not drop the whole map.
func fillMeshExternalServicesOnEgress(
	ctx context.Context,
	outbound core_xds.EgressEndpointMap,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	mesh *core_mesh.MeshResource,
	loader datasource.Loader,
) {
	for _, mes := range meshExternalServices {
		es := &core_xds.ExternalService{
			Protocol:      mes.Spec.Match.Protocol,
			OwnerResource: kri.From(mes),
		}
		tags := maps.Clone(mes.Meta.GetLabels())
		if tags == nil {
			tags = map[string]string{}
		}
		tls := mes.Spec.Tls
		if tls != nil && tls.Enabled {
			if err := setTlsConfiguration(ctx, tls, es, mesh.GetMeta().GetName(), loader); err != nil {
				outboundLog.Error(err, "unable to create MeshExternalService endpoint. Endpoint won't be included in the XDS.", "name", mes.Meta.GetName(), "mesh", mes.Meta.GetMesh())
				continue
			}
		}

		// if all ip make it static - it's done in endpoint_cluster_configurer
		var endpoints []core_xds.Endpoint
		for i, endpoint := range pointer.Deref(mes.Spec.Endpoints) {
			if i == 0 && es.ServerName == "" && govalidator.IsDNSName(endpoint.Address) && tls != nil && tls.Enabled {
				es.ServerName = endpoint.Address
			}
			priority := pointer.DerefOr(endpoint.Priority, 0)
			endpoints = append(endpoints, core_xds.Endpoint{
				Target:          endpoint.Address,
				Port:            uint32(endpoint.Port),
				Weight:          1,
				ExternalService: es,
				Tags:            tags,
				Locality: &core_xds.Locality{
					Priority: priority,
					SubZone:  "priority-" + strconv.Itoa(int(priority)),
				},
			})
		}
		if len(endpoints) == 0 {
			continue
		}

		// Always unified (KRI) naming: the embedded egress is new infrastructure that only
		// supports Exclusive MeshServices mode.
		serviceName := destinationname.MustResolve(mes, mes.Spec.Match)
		group := outbound[serviceName]
		group.Protocol = es.Protocol
		group.OwnerResource = es.OwnerResource
		group.Endpoints = append(group.Endpoints, endpoints...)
		outbound[serviceName] = group
	}
}

// loadSecureBytes resolves a SecureDataSource on the control plane. Secret references keep going
// through the mesh-snapshot datasource.Loader so no additional store I/O is introduced; inline
// values are resolved directly via SecureDataSource.ReadByControlPlane. File and EnvVar would read
// the control plane's own filesystem and environment, so they are refused here as well as in the
// validator - a MeshExternalService with an extension, or one synced from another zone, does not
// necessarily go through validation.
func loadSecureBytes(ctx context.Context, sds *datasource_api.SecureDataSource, mesh string, loader datasource.Loader) ([]byte, error) {
	if sds == nil {
		return nil, nil
	}
	switch sds.Type {
	case datasource_api.SecureDataSourceSecretRef:
		if sds.SecretRef == nil {
			return nil, errors.New("secretRef must be defined")
		}
		return loader.Load(ctx, mesh, &system_proto.DataSource{
			Type: &system_proto.DataSource_Secret{Secret: sds.SecretRef.Name},
		})
	case datasource_api.SecureDataSourceInline:
		if sds.InsecureInline == nil {
			return nil, errors.New("insecureInline must be defined")
		}
		return sds.ReadByControlPlane(ctx, nil, mesh)
	default:
		return nil, errors.Errorf("datasource type: %s is not supported on MeshExternalService", sds.Type)
	}
}

const (
	// Constants for Locality Aware load balancing
	// The Highest priority 0 shall be assigned to all locally available services
	// A priority of 1 is for ExternalServices and services exposed on neighboring ingress-es
	priorityLocal  = 0
	priorityRemote = 1
)

func GetLocality(otherZone *string) *core_xds.Locality {
	if otherZone == nil {
		return nil
	}

	// We always set the Locality with a local Priority: this gives extra
	// visibility about it in /clusters etc., and cross-zone priority is now
	// owned by the MeshLoadBalancingStrategy policy.
	//
	// Setting this regardless also solves the problem that endpoints have
	// problems when moving from one locality to another
	// https://github.com/envoyproxy/envoy/issues/12392

	return &core_xds.Locality{
		Zone:     *otherZone,
		Priority: priorityLocal,
	}
}

func getZone(tags map[string]string) *string {
	if zone, ok := tags[mesh_proto.ZoneTag]; ok {
		return &zone
	}
	return nil
}
