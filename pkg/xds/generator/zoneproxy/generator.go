package zoneproxy

import (
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/core/xds/origin"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/v3/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
)

func GenerateCDS(
	proxy *core_xds.Proxy,
	services envoy_common.Services,
	origin origin.Origin,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]

		clusterName := service.BackendRef().Resource().String()

		resource, err := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName).
			Configure(envoy_clusters.EdsCluster()).
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
	origin origin.Origin,
) (*core_xds.ResourceSet, error) {
	rs := core_xds.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]
		endpoints := endpointMap[serviceName]
		clusterName := service.BackendRef().Resource().String()

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

// CreateFilterChain builds filter chains for a listener with mTLS enabled.
// TLSInspector checks the SNI value, which contains service name and tags in
// the format "backend{cluster=2,version=1}". For every unique destination
// from MeshHTTPRoutes and MeshTCPRoutes we create a
// FilterChainsMatcher. Limitation: extra tags on outbound in Universal mode
// do not work across zones. Traffic is not decrypted, so certificates and
// mTLS settings are not needed here.
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
	destinations MeshDestinations,
) envoy_common.Services {
	acc := envoy_common.NewServicesAccumulator()

	sniUsed := map[string]struct{}{}

	for _, br := range destinations.BackendRefs {
		if _, ok := sniUsed[br.SNI]; ok {
			continue
		}

		sniUsed[br.SNI] = struct{}{}

		clusterName := br.Resource().String()

		cluster := plugins_xds.NewClusterBuilder().
			WithName(clusterName).
			WithService(br.EndpointMapKey).
			WithSNI(br.SNI).
			Build()

		acc.AddBackendRef(&br.ResolvedBackendRef, cluster)
	}

	return acc.Services()
}
