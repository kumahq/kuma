package zoneproxy

import (
	"slices"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	"github.com/kumahq/kuma/pkg/core/naming"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/origin"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

func GenerateCDS(
	proxy *core_xds.Proxy,
	destinations MeshDestinations,
	services envoy_common.Services,
	meshName string,
	origin origin.Origin,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()

	unifiedNaming := proxy.Metadata.HasFeature(xds_types.FeatureUnifiedResourceNaming)
	matchAll := destinations.KumaIoServices[mesh_proto.MatchAllTag]

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]

		var tagsSlice envoy_tags.TagsSlice
		for _, cluster := range service.Clusters() {
			tagsSlice = append(tagsSlice, cluster.Tags())
		}

		tagKeySlice := slices.Concat(tagsSlice, matchAll).
			ToTagKeysSlice().
			Transform(envoy_tags.Without(mesh_proto.ServiceTag))

		clusterName := resolveClusterName(service.BackendRef(), serviceName, meshName, unifiedNaming)

		resource, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName).
			Configure(envoy_clusters.EdsCluster()).
			Configure(envoy_clusters.LbSubset(tagKeySlice)).
			Configure(envoy_clusters.DefaultTimeout()).
			Build()
		if err != nil {
			return nil, err
		}

		rs.Add(&core_xds.Resource{
			Name:     resource.GetName(),
			Origin:   origin,
			Resource: resource,
		})
	}

	return rs, nil
}

func GenerateEDS(
	proxy *core_xds.Proxy,
	endpointMap core_xds.EndpointMap,
	services envoy_common.Services,
	meshName string,
	origin origin.Origin,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()

	unifiedNaming := proxy.Metadata.HasFeature(xds_types.FeatureUnifiedResourceNaming)

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]
		endpoints := endpointMap[serviceName]
		clusterName := resolveClusterName(service.BackendRef(), serviceName, meshName, unifiedNaming)

		cla, err := envoy_endpoints.CreateClusterLoadAssignment(clusterName, endpoints, proxy.APIVersion)
		if err != nil {
			return nil, err
		}

		rs.Add(&core_xds.Resource{
			Name:     clusterName,
			Origin:   origin,
			Resource: cla,
		})
	}

	return rs, nil
}

// CreateFilterChain adds filter chains to a listener. Generated listener assumes that
// mTLS is on. Using TLSInspector we sniff SNI value. SNI value has service name
// and tag values specified with the following format:
// "backend{cluster=2,version=1}". We take all possible destinations from
// TrafficRoutes + MeshHTTPRoutes + GatewayRoutes and generate
// FilterChainsMatcher for each unique destination. This approach has
// a limitation: additional tags on outbound in Universal mode won't work across
// different zones. Traffic is NOT decrypted here, therefore we don't need
// certificates and mTLS settings.
func CreateFilterChain(
	proxy *core_xds.Proxy,
	cluster envoy_common.Cluster,
) *envoy_listeners.FilterChainBuilder {
	return envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.MatchTransportProtocol(core_meta.ProtocolTLS)).
		Configure(envoy_listeners.MatchServerNames(cluster.SNI())).
		Configure(envoy_listeners.TcpProxyDeprecatedWithMetadata(cluster.Name(), cluster))
}

func GetServices(
	proxy *core_xds.Proxy,
	destinations MeshDestinations,
	endpointMap core_xds.EndpointMap,
	availableServices []*mesh_proto.ZoneIngress_AvailableService,
) envoy_common.Services {
	acc := envoy_common.NewServicesAccumulator(nil)

	unifiedNaming := proxy.Metadata.HasFeature(xds_types.FeatureUnifiedResourceNaming)
	getName := naming.GetNameOrFallbackFunc(unifiedNaming)
	matchAll := destinations.KumaIoServices[mesh_proto.MatchAllTag]
	sniUsed := map[string]struct{}{}

	for _, service := range availableServices {
		serviceName := service.Tags[mesh_proto.ServiceTag]
		kumaIoServices := destinations.KumaIoServices[serviceName]
		clusterName := envoy_names.GetMeshClusterName(service.Mesh, serviceName)
		endpoints := endpointMap[serviceName]

		for _, destination := range slices.Concat(kumaIoServices, matchAll) {
			sni := tls.SNIFromTags(destination.
				WithTags(mesh_proto.ServiceTag, serviceName).
				WithTags("mesh", service.Mesh),
			)

			if _, ok := sniUsed[sni]; ok {
				continue
			}

			sniUsed[sni] = struct{}{}

			// relevantTags is a set of tags for which it actually makes sense to do LB split on.
			// If the endpoint list is the same with or without the tag, we should just not do the split.
			// However, we should preserve full SNI, because the client expects Zone Proxy to support it.
			// This solves the problem that Envoy deduplicate endpoints of the same address and different metadata.
			// example 1:
			// Ingress1 (10.0.0.1) supports service:a,version:1 and service:a,version:2
			// Ingress2 (10.0.0.2) supports service:a,version:1 and service:a,version:2
			// If we want to split by version, we don't need to do LB subset on version.
			//
			// example 2:
			// Ingress1 (10.0.0.1) supports service:a,version:1
			// Ingress2 (10.0.0.2) supports service:a,version:2
			// If we want to split by version, we need LB subset.
			relevantTags := envoy_tags.Tags{}
			for key, value := range destination {
				matchedTargets := map[string]struct{}{}
				allTargets := map[string]struct{}{}
				for _, endpoint := range endpoints {
					address := endpoint.Address()
					if endpoint.Tags[key] == value || value == mesh_proto.MatchAllTag {
						matchedTargets[address] = struct{}{}
					}
					allTargets[address] = struct{}{}
				}
				if len(matchedTargets) < len(allTargets) {
					relevantTags[key] = value
				}
			}

			cluster := plugins_xds.NewClusterBuilder().
				WithName(clusterName).
				WithSNI(sni).
				WithMesh(service.Mesh).
				WithService(serviceName).
				WithTags(relevantTags).
				Build()

			acc.Add(cluster)
		}
	}

	for _, br := range destinations.BackendRefs {
		if _, ok := sniUsed[br.SNI]; ok {
			continue
		}

		sniUsed[br.SNI] = struct{}{}

		clusterName := getName(br.Resource().String(), br.LegacyServiceName)

		cluster := plugins_xds.NewClusterBuilder().
			WithName(clusterName).
			WithService(br.LegacyServiceName).
			WithSNI(br.SNI).
			WithMesh(br.Mesh).
			Build()

		acc.AddBackendRef(&br.ResolvedBackendRef, cluster)
	}

	return acc.Services()
}

func resolveClusterName(
	br *resolve.ResolvedBackendRef,
	serviceName string,
	meshName string,
	unifiedNaming bool,
) string {
	switch {
	case !br.ReferencesRealResource():
		return envoy_names.GetMeshClusterName(meshName, serviceName)
	case unifiedNaming:
		return br.Resource().String()
	default:
		return serviceName
	}
}
