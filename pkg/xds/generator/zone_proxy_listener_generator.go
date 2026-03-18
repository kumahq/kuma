package generator

import (
	"context"

	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/naming"
	unified_naming "github.com/kumahq/kuma/v2/pkg/core/naming/unified-naming"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v2/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v2/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v2/pkg/envoy/builders/tls"
	plugins_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/v2/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/metadata"
	system_names "github.com/kumahq/kuma/v2/pkg/xds/generator/system_names"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/zoneproxy"
)

// ZoneProxyListenerGenerator generates Envoy listeners for zone proxy listeners
// embedded in a regular Dataplane resource (DataplaneZoneListeners).
// It is a no-op when proxy.DataplaneZoneListeners is nil.
// Only MeshExternalService and MeshIdentity are supported; legacy ExternalService
// and mesh.mtls are not handled here.
type ZoneProxyListenerGenerator struct{}

func (g ZoneProxyListenerGenerator) Generate(
	_ context.Context,
	_ *core_xds.ResourceSet,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	if proxy.DataplaneZoneListeners == nil {
		return nil, nil
	}

	rs := core_xds.NewResourceSet()

	for _, il := range proxy.DataplaneZoneListeners.IngressListeners {
		generated, err := g.generateIngressListener(proxy, xdsCtx, il)
		if err != nil {
			return nil, err
		}
		rs.AddSet(generated)
	}

	for _, el := range proxy.DataplaneZoneListeners.EgressListeners {
		generated, err := g.generateEgressListener(proxy, el)
		if err != nil {
			return nil, err
		}
		rs.AddSet(generated)
	}

	return rs, nil
}

func (g ZoneProxyListenerGenerator) generateIngressListener(
	proxy *core_xds.Proxy,
	xdsCtx xds_context.Context,
	il *core_xds.DataplaneIngressListener,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()
	cp := xdsCtx.ControlPlane
	mr := il.MeshResources
	meshName := mr.Mesh.GetMeta().GetName()
	unifiedNaming := unified_naming.Enabled(proxy.Metadata, mr.Mesh)
	getName := naming.GetNameOrFallbackFunc(unifiedNaming)

	address := il.Listener.Address
	port := il.Listener.Port

	zoneIngressListenerName := naming.ContextualZoneIngressListenerName(il.Listener.Name)
	listenerName := getName(zoneIngressListenerName, envoy_names.GetInboundListenerName(address, port))
	statPrefix := getName(zoneIngressListenerName, "")

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(statPrefix)).
		Configure(envoy_listeners.TLSInspector())

	meshResources := xds_context.Resources{MeshLocalResources: mr.Resources}

	// Only expose services local to this zone.
	localMS := &meshservice_api.MeshServiceResourceList{}
	for _, ms := range meshResources.MeshServices().GetItems() {
		if lbls := ms.GetMeta().GetLabels(); lbls == nil || lbls[mesh_proto.ZoneTag] == "" || lbls[mesh_proto.ZoneTag] == cp.Zone {
			_ = localMS.AddItem(ms)
		}
	}

	dest := zoneproxy.BuildMeshDestinations(
		nil, // no legacy available services
		cp.SystemNamespace,
		meshResources,
		localMS,
		meshResources.MeshMultiZoneServices(),
	)

	services := zoneproxy.GetServices(dest, mr.EndpointMap, nil, unifiedNaming)
	clusters := services.Clusters()

	cds, err := zoneproxy.GenerateCDS(proxy, dest, services, meshName, metadata.OriginIngress, unifiedNaming)
	if err != nil {
		return nil, err
	}
	rs.AddSet(cds)

	eds, err := zoneproxy.GenerateEDS(proxy, mr.EndpointMap, services, meshName, metadata.OriginIngress, unifiedNaming)
	if err != nil {
		return nil, err
	}
	rs.AddSet(eds)

	for _, cluster := range clusters {
		listener.Configure(envoy_listeners.FilterChain(zoneproxy.CreateFilterChain(proxy, cluster)))
	}

	if len(clusters) == 0 {
		return nil, nil
	}

	resource, err := listener.Build()
	if err != nil {
		return nil, err
	}
	rs.Add(&core_xds.Resource{
		Name:     resource.GetName(),
		Origin:   metadata.OriginIngress,
		Resource: resource,
	})
	return rs, nil
}

func (g ZoneProxyListenerGenerator) generateEgressListener(
	proxy *core_xds.Proxy,
	el *core_xds.DataplaneEgressListener,
) (*core_xds.ResourceSet, error) {
	if proxy.WorkloadIdentity == nil {
		return nil, nil
	}

	rs := core_xds.NewResourceSet()
	unifiedNaming := unified_naming.Enabled(proxy.Metadata, el.MeshResources.Mesh)
	getName := naming.GetNameOrFallbackFunc(unifiedNaming)

	address := el.Listener.Address
	port := el.Listener.Port

	zoneEgressListenerName := naming.ContextualZoneEgressListenerName(el.Listener.Name)
	listenerName := getName(zoneEgressListenerName, envoy_names.GetInboundListenerName(address, port))
	statPrefix := getName(zoneEgressListenerName, "")

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.InboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.StatPrefix(statPrefix)).
		Configure(envoy_listeners.TLSInspector())

	downstreamTLS, err := meshIdentityDownstreamTLS(proxy)
	if err != nil {
		return nil, err
	}

	var filterChainBuilders []*envoy_listeners.FilterChainBuilder
	for _, cluster := range g.getMeshExternalServiceClusters(el.MeshResources, unifiedNaming) {
		filterChainBuilders = append(filterChainBuilders,
			g.buildEgressFilterChain(proxy, el.MeshResources, downstreamTLS, cluster, unifiedNaming),
		)
		cds, err := g.genClusterCDS(proxy, el.MeshResources.EndpointMap[cluster.Name()], cluster)
		if err != nil {
			return nil, err
		}
		rs.Add(cds)
	}

	for _, fcb := range filterChainBuilders {
		listener.Configure(envoy_listeners.FilterChain(fcb))
	}

	if len(filterChainBuilders) > 0 {
		resource, err := listener.Build()
		if err != nil {
			return nil, err
		}
		rs.Add(&core_xds.Resource{
			Name:     resource.GetName(),
			Origin:   metadata.OriginEgress,
			Resource: resource,
		})
	}

	return rs, nil
}

// getMeshExternalServiceClusters returns clusters for MeshExternalService BackendRefs only.
func (g ZoneProxyListenerGenerator) getMeshExternalServiceClusters(
	resources *core_xds.MeshProxyResources,
	unifiedNaming bool,
) []envoy_common.Cluster {
	svcAcc := envoy_common.NewServicesAccumulator(nil)
	localResources := xds_context.Resources{MeshLocalResources: resources.Resources}
	destinations := zoneproxy.BuildMeshDestinations(
		nil,
		"",
		localResources,
		localResources.MeshExternalServices(),
	)

	sniUsed := map[string]struct{}{}
	for _, ref := range destinations.BackendRefs {
		endpoints := resources.EndpointMap[naming.GetNameOrFallback(unifiedNaming, ref.Resource().String(), ref.LegacyServiceName)]
		if _, ok := sniUsed[ref.SNI]; ok || len(endpoints) == 0 || !endpoints[0].IsExternalService() {
			continue
		}
		sniUsed[ref.SNI] = struct{}{}
		clusterName := naming.GetNameOrFallback(unifiedNaming, ref.Resource().String(), ref.LegacyServiceName)
		cluster := plugins_xds.NewClusterBuilder().
			WithName(clusterName).
			WithService(ref.LegacyServiceName).
			WithSNI(ref.SNI).
			WithExternalService(true).
			Build()
		svcAcc.AddBackendRef(&ref.ResolvedBackendRef, cluster)
	}
	return svcAcc.Services().Clusters()
}

func (g ZoneProxyListenerGenerator) genClusterCDS(
	proxy *core_xds.Proxy,
	endpoints []core_xds.Endpoint,
	cluster envoy_common.Cluster,
) (*core_xds.Resource, error) {
	ipv6 := proxy.Dataplane.IsIPv6()
	systemCAPath := proxy.Metadata.GetSystemCaPath()
	protocol := endpoints[0].Protocol()

	resource, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, cluster.Name()).
		Configure(envoy_clusters.DefaultTimeout()).
		ConfigureIf(core_meta.IsHTTP(protocol), envoy_clusters.Http()).
		ConfigureIf(core_meta.IsHTTP2Based(protocol), envoy_clusters.Http2()).
		Configure(envoy_clusters.ProvidedCustomEndpointCluster(ipv6, true, endpoints...)).
		Configure(envoy_clusters.MeshExternalServiceClientSideTLS(endpoints, systemCAPath, true)).
		Build()
	if err != nil {
		return nil, err
	}
	return &core_xds.Resource{
		Name:           resource.GetName(),
		Origin:         metadata.OriginEgress,
		Resource:       resource,
		Protocol:       endpoints[0].ExternalService.Protocol,
		ResourceOrigin: endpoints[0].ExternalService.OwnerResource,
	}, nil
}

func (g ZoneProxyListenerGenerator) buildEgressFilterChain(
	proxy *core_xds.Proxy,
	resources *core_xds.MeshProxyResources,
	downstreamTLS *envoy_tls.DownstreamTlsContext,
	cluster envoy_common.Cluster,
	unifiedNaming bool,
) *envoy_listeners.FilterChainBuilder {
	meshName := resources.Mesh.GetMeta().GetName()
	endpoints := resources.EndpointMap[cluster.Name()]
	getName := naming.GetNameOrFallbackFunc(endpoints[0].IsMeshExternalService)
	esName := naming.GetNameOrFallback(unifiedNaming, cluster.Name(), cluster.Service())
	filterChainName := getName(esName, envoy_names.GetEgressFilterChainName(esName, meshName))
	routeConfigName := getName(esName, envoy_names.GetOutboundRouteName(esName))
	virtualHostName := esName

	filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, filterChainName).
		Configure(envoy_listeners.MatchTransportProtocol(core_meta.ProtocolTLS)).
		Configure(envoy_listeners.MatchServerNames(cluster.SNI())).
		Configure(envoy_listeners.DownstreamTlsContext(downstreamTLS))

	if !core_meta.IsHTTPBased(endpoints[0].Protocol()) {
		return filterChain.Configure(envoy_listeners.TcpProxyDeprecatedWithMetadata(esName, cluster))
	}

	routes := envoy_common.Routes{envoy_common.NewRoute(envoy_common.WithCluster(cluster))}

	return filterChain.
		Configure(envoy_listeners.HttpConnectionManager(esName, false, proxy.InternalAddresses, proxy.Metadata.GetIPv6Enabled())).
		Configure(envoy_listeners.HttpOutboundRoute(routeConfigName, virtualHostName, routes, nil))
}

// meshIdentityDownstreamTLS builds a DownstreamTlsContext from the proxy's WorkloadIdentity.
// Returns nil when WorkloadIdentity is nil.
func meshIdentityDownstreamTLS(proxy *core_xds.Proxy) (*envoy_tls.DownstreamTlsContext, error) {
	if proxy.WorkloadIdentity == nil {
		return nil, nil
	}

	validationCtx := func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
		return bldrs_tls.SdsSecretConfigSource(
			system_names.SystemResourceNameCABundle,
			bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
		)
	}
	if proxy.WorkloadIdentity.ExternalValidationSourceConfigurer != nil {
		ext := proxy.WorkloadIdentity.ExternalValidationSourceConfigurer
		validationCtx = func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
			return ext()
		}
	}

	return bldrs_tls.NewDownstreamTLSContext().
		Configure(
			bldrs_tls.DownstreamCommonTlsContext(
				bldrs_tls.NewCommonTlsContext().
					Configure(
						bldrs_tls.CombinedCertificateValidationContext(
							bldrs_tls.NewCombinedCertificateValidationContext().
								Configure(
									bldrs_tls.DefaultValidationContext(
										bldrs_tls.NewDefaultValidationContext(),
									),
								).
								Configure(
									bldrs_tls.ValidationContextSdsSecretConfig(
										bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(validationCtx()),
									),
								),
						),
					).
					Configure(
						bldrs_tls.TlsCertificateSdsSecretConfigs([]*bldrs_common.Builder[envoy_tls.SdsSecretConfig]{
							bldrs_tls.NewTlsCertificateSdsSecretConfigs().Configure(
								proxy.WorkloadIdentity.IdentitySourceConfigurer(),
							),
						}),
					),
			),
		).
		Configure(bldrs_tls.RequireClientCertificate(true)).
		Build()
}
