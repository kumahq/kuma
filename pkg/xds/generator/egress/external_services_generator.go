package egress

import (
	"slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	"github.com/kumahq/kuma/pkg/core/naming"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
	"github.com/kumahq/kuma/pkg/xds/generator/zoneproxy"
)

func genExternalResources(
	proxy *core_xds.Proxy,
	resources *core_xds.MeshResources,
	secretsTracker core_xds.SecretsTracker,
	unifiedNaming bool,
) (*core_xds.ResourceSet, []*envoy_listeners.FilterChainBuilder, error) {
	rs := core_xds.NewResourceSet()

	var filterChainBuilders []*envoy_listeners.FilterChainBuilder

	for _, cluster := range getExternalServicesClusters(resources, unifiedNaming) {
		filterChainBuilders = append(
			filterChainBuilders,
			buildExternalServiceFilterChain(proxy, resources, secretsTracker, cluster, unifiedNaming),
		)

		cds, err := genExternalServicesCDS(proxy, resources.EndpointMap[cluster.Service()], cluster)
		if err != nil {
			return nil, nil, err
		}
		rs.Add(cds)
	}

	return rs, filterChainBuilders, nil
}

func getExternalServicesClusters(
	resources *core_xds.MeshResources,
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

	meshName := resources.Mesh.GetMeta().GetName()
	matchAll := destinations.KumaIoServices[mesh_proto.MatchAllTag]
	sniUsed := map[string]struct{}{}

	for _, es := range resources.ExternalServices {
		esName := es.Spec.GetService()
		endpoints := resources.EndpointMap[esName]

		if len(endpoints) == 0 || !endpoints[0].IsExternalService() {
			continue
		}

		for _, dest := range slices.Concat(destinations.KumaIoServices[esName], matchAll) {
			destTags := dest.WithTags("mesh", meshName)

			sni := tls.SNIFromTags(destTags.WithTags(mesh_proto.ServiceTag, esName))
			if _, ok := sniUsed[sni]; ok {
				continue
			}

			sniUsed[sni] = struct{}{}

			// There is a case where multiple meshes contain services with
			// the same names, so we cannot use just "serviceName" as a cluster
			// name as we would overwrite some clusters with the latest one
			cluster := xds.NewClusterBuilder().
				WithName(envoy_names.GetMeshClusterName(meshName, esName)).
				WithService(esName).
				WithSNI(sni).
				WithExternalService(true).
				WithTags(destTags).
				Build()

			svcAcc.Add(cluster)
		}
	}

	for _, ref := range destinations.BackendRefs {
		endpoints := resources.EndpointMap[ref.LegacyServiceName]
		if _, ok := sniUsed[ref.SNI]; ok || len(endpoints) == 0 || !endpoints[0].IsExternalService() {
			continue
		}

		sniUsed[ref.SNI] = struct{}{}

		clusterName := naming.GetNameOrFallback(
			unifiedNaming,
			ref.Resource().String(),
			ref.LegacyServiceName,
		)

		cluster := xds.NewClusterBuilder().
			WithName(clusterName).
			WithService(ref.LegacyServiceName).
			WithSNI(ref.SNI).
			WithExternalService(true).
			Build()

		svcAcc.AddBackendRef(&ref.ResolvedBackendRef, cluster)
	}

	return svcAcc.Services().Clusters()
}

func genExternalServicesCDS(
	proxy *core_xds.Proxy,
	endpoints []core_xds.Endpoint,
	cluster envoy_common.Cluster,
) (*core_xds.Resource, error) {
	ipv6 := proxy.ZoneEgressProxy.ZoneEgressResource.IsIPv6()
	systemCAPath := proxy.Metadata.GetSystemCaPath()

	protocol := endpoints[0].Protocol()
	isMES := endpoints[0].IsMeshExternalService()

	resource, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, cluster.Name()).
		Configure(envoy_clusters.DefaultTimeout()).
		ConfigureIf(core_meta.IsHTTP(protocol), envoy_clusters.Http()).
		ConfigureIf(core_meta.IsHTTP2Based(protocol), envoy_clusters.Http2()).
		ConfigureIf(isMES, envoy_clusters.ProvidedCustomEndpointCluster(ipv6, true, endpoints...)).
		ConfigureIf(isMES, envoy_clusters.MeshExternalServiceClientSideTLS(endpoints, systemCAPath, true)).
		ConfigureIf(!isMES, envoy_clusters.ProvidedEndpointCluster(ipv6, endpoints...)).
		ConfigureIf(!isMES, envoy_clusters.ClientSideTLS(endpoints)).
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

func buildExternalServiceFilterChain(
	proxy *core_xds.Proxy,
	resources *core_xds.MeshResources,
	secretsTracker core_xds.SecretsTracker,
	cluster envoy_common.Cluster,
	unifiedNaming bool,
) *envoy_listeners.FilterChainBuilder {
	meshName := resources.Mesh.GetMeta().GetName()
	endpoints := resources.EndpointMap[cluster.Service()]
	getName := naming.GetNameOrFallbackFunc(endpoints[0].IsMeshExternalService)
	esName := naming.GetNameOrFallback(unifiedNaming, cluster.Name(), cluster.Service())
	filterChainName := getName(esName, envoy_names.GetEgressFilterChainName(esName, meshName))
	routeConfigName := getName(esName, envoy_names.GetOutboundRouteName(esName))
	virtualHostName := esName

	filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, filterChainName).
		Configure(envoy_listeners.ServerSideMTLS(resources.Mesh, secretsTracker, nil, nil, unifiedNaming, false)).
		Configure(envoy_listeners.MatchTransportProtocol(core_meta.ProtocolTLS)).
		Configure(envoy_listeners.MatchServerNames(cluster.SNI())).
		// Zone Egress will configure these filter chains only for meshes with mTLS enabled, so we can safely pass here true
		Configure(envoy_listeners.NetworkRBAC(esName, true, resources.ExternalServicePermissionMap[esName]))

	// Protocol is not HTTP based, so we can use TCP proxy instead of HTTP connection manager and return early
	if !core_meta.IsHTTPBased(endpoints[0].Protocol()) {
		return filterChain.Configure(envoy_listeners.TcpProxyDeprecatedWithMetadata(esName, cluster))
	}

	var routes envoy_common.Routes
	for _, rl := range resources.ExternalServiceRateLimits[esName] {
		if rl.Spec.GetConf().GetHttp() == nil {
			continue
		}

		routes = append(routes, envoy_common.NewRoute(
			envoy_common.WithCluster(cluster),
			envoy_common.WithMatchHeaderRegex(tags.TagsHeaderName, tags.MatchSourceRegex(rl)),
			envoy_common.WithRateLimit(rl.Spec),
		))
	}
	// Add the default fall-back route
	routes = append(routes, envoy_common.NewRoute(envoy_common.WithCluster(cluster)))

	return filterChain.
		Configure(envoy_listeners.HttpConnectionManager(esName, false, proxy.InternalAddresses)).
		Configure(envoy_listeners.FaultInjection(resources.ExternalServiceFaultInjections[esName]...)).
		Configure(envoy_listeners.RateLimit(resources.ExternalServiceRateLimits[esName])).
		Configure(envoy_listeners.HttpOutboundRoute(routeConfigName, virtualHostName, routes, nil))
}
