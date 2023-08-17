package zoneproxy

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

func GenerateCDS(
	services []string,
	destinationsPerService map[string][]envoy_tags.Tags,
	apiVersion core_xds.APIVersion,
	meshName string,
	origin string,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource
	for _, service := range services {
		clusterName := envoy_names.GetMeshClusterName(meshName, service)

		tagSlice := envoy_tags.TagsSlice(append(destinationsPerService[service], destinationsPerService[mesh_proto.MatchAllTag]...))
		tagKeySlice := tagSlice.ToTagKeysSlice().Transform(
			envoy_tags.Without(mesh_proto.ServiceTag),
		)

		edsCluster, err := envoy_clusters.NewClusterBuilder(apiVersion, clusterName).
			Configure(envoy_clusters.EdsCluster()).
			Configure(envoy_clusters.LbSubset(tagKeySlice)).
			Configure(envoy_clusters.DefaultTimeout()).
			Build()
		if err != nil {
			return nil, err
		}
		resources = append(resources, &core_xds.Resource{
			Name:     clusterName,
			Origin:   origin,
			Resource: edsCluster,
		})
	}

	return resources, nil
}

func GenerateEDS(
	services []string,
	endpointMap core_xds.EndpointMap,
	apiVersion core_xds.APIVersion,
	meshName string,
	origin string,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for _, service := range services {
		endpoints := endpointMap[service]
		clusterName := envoy_names.GetMeshClusterName(meshName, service)

		cla, err := envoy_endpoints.CreateClusterLoadAssignment(clusterName, endpoints, apiVersion)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &core_xds.Resource{
			Name:     clusterName,
			Origin:   origin,
			Resource: cla,
		})
	}

	return resources, nil
}

// AddFilterChains generates one Ingress Listener. Generated listener assumes that
// mTLS is on. Using TLSInspector we sniff SNI value. SNI value has service name
// and tag values specified with the following format:
// "backend{cluster=2,version=1}". We take all possible destinations from
// TrafficRoutes + MeshHTTPRoutes + GatewayRoutes and generate
// FilterChainsMatcher for each unique destination. This approach has
// a limitation: additional tags on outbound in Universal mode won't work across
// different zones. Traffic is NOT decrypted here, therefore we don't need
// certificates and mTLS settings.
func AddFilterChains(
	availableServices []*mesh_proto.ZoneIngress_AvailableService,
	apiVersion core_xds.APIVersion,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	destinationsPerService map[string][]envoy_tags.Tags,
) {
	sniUsed := map[string]bool{}
	for _, service := range availableServices {
		serviceName := service.Tags[mesh_proto.ServiceTag]
		destinations := destinationsPerService[serviceName]
		destinations = append(destinations, destinationsPerService[mesh_proto.MatchAllTag]...)
		clusterName := envoy_names.GetMeshClusterName(service.Mesh, serviceName)

		for _, destination := range destinations {
			sni := tls.SNIFromTags(destination.
				WithTags(mesh_proto.ServiceTag, serviceName).
				WithTags("mesh", service.Mesh),
			)
			if sniUsed[sni] {
				continue
			}
			sniUsed[sni] = true

			cluster := envoy_common.NewCluster(
				envoy_common.WithName(clusterName),
				envoy_common.WithTags(destination),
			)

			filterChain := envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(apiVersion, envoy_common.AnonymousResource).Configure(
					envoy_listeners.MatchTransportProtocol("tls"),
					envoy_listeners.MatchServerNames(sni),
					envoy_listeners.TcpProxyDeprecatedWithMetadata(
						clusterName,
						cluster,
					),
				),
			)

			listenerBuilder.Configure(filterChain)
		}
	}
}
