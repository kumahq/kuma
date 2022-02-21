package egress

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

type ExternalServicesGenerator struct {
}

// Generate will generate envoy resources for one mesh (when mTLS enabled)
func (g *ExternalServicesGenerator) Generate(
	proxy *core_xds.Proxy,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	meshResources *core_xds.MeshResources,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	apiVersion := proxy.APIVersion
	endpointMap := meshResources.EndpointMap
	destinations := buildDestinations(meshResources.TrafficRoutes)
	services := g.buildServices(endpointMap)

	g.addFilterChains(
		apiVersion,
		destinations,
		endpointMap,
		meshResources,
		listenerBuilder,
	)

	cds, err := g.generateCDS(
		meshResources.Mesh.GetMeta().GetName(),
		apiVersion,
		services,
		endpointMap,
		proxy.ZoneEgressProxy.ZoneEgressResource.IsIPv6(),
	)
	if err != nil {
		return nil, err
	}
	resources.Add(cds...)

	return resources, nil
}

func (*ExternalServicesGenerator) generateCDS(
	meshName string,
	apiVersion envoy_common.APIVersion,
	services []string,
	endpointMap core_xds.EndpointMap,
	isIPV6 bool,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for _, serviceName := range services {
		endpoints := endpointMap[serviceName]

		if len(endpoints) == 0 {
			log.Info("no endpoints for service", "serviceName", serviceName)

			continue
		}

		// There is a case where multiple meshes contain services with
		// the same names, so we cannot use just "serviceName" as a cluster
		// name as we would overwrite some clusters with the latest one
		clusterName := envoy_names.GetMeshClusterName(meshName, serviceName)

		clusterBuilder := envoy_clusters.NewClusterBuilder(apiVersion).
			Configure(envoy_clusters.ProvidedEndpointCluster(
				clusterName,
				isIPV6,
				endpoints...,
			)).
			Configure(envoy_clusters.ClientSideTLS(endpoints))

		switch endpoints[0].Tags[mesh_proto.ProtocolTag] {
		case core_mesh.ProtocolHTTP:
			clusterBuilder.Configure(envoy_clusters.Http())
		case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
			clusterBuilder.Configure(envoy_clusters.Http2())
		}

		cluster, err := clusterBuilder.Build()
		if err != nil {
			return nil, err
		}

		resources = append(resources, &core_xds.Resource{
			Name:     serviceName,
			Origin:   OriginEgress,
			Resource: cluster,
		})
	}

	return resources, nil
}

func (*ExternalServicesGenerator) buildServices(
	endpointMap core_xds.EndpointMap,
) []string {
	var services []string

	for serviceName, endpoints := range endpointMap {
		if len(endpoints) > 0 && endpoints[0].IsExternalService() {
			services = append(services, serviceName)
		}
	}

	sort.Strings(services)

	return services
}

func (*ExternalServicesGenerator) addFilterChains(
	apiVersion envoy_common.APIVersion,
	destinationsPerService map[string][]envoy_common.Tags,
	endpointMap core_xds.EndpointMap,
	meshResources *core_xds.MeshResources,
	listenerBuilder *envoy_listeners.ListenerBuilder,
) {
	meshName := meshResources.Mesh.GetMeta().GetName()

	sniUsed := map[string]bool{}

	for _, es := range meshResources.ExternalServices {
		serviceName := es.Spec.GetService()

		endpoints := endpointMap[serviceName]

		if len(endpoints) == 0 {
			log.Info("no endpoints for service", "serviceName", serviceName)
			// There is no need to generate filter chain if there is no
			// endpoints for the service
			continue
		}

		destinations := destinationsPerService[serviceName]
		destinations = append(destinations, destinationsPerService[mesh_proto.MatchAllTag]...)

		for _, destination := range destinations {
			meshDestination := destination.
				WithTags(mesh_proto.ServiceTag, serviceName).
				WithTags("mesh", meshName)

			sni := tls.SNIFromTags(meshDestination)

			if sniUsed[sni] {
				continue
			}

			sniUsed[sni] = true

			// There is a case where multiple meshes contain services with
			// the same names, so we cannot use just "serviceName" as a cluster
			// name as we would overwrite some clusters with the latest one
			clusterName := envoy_names.GetMeshClusterName(meshName, serviceName)

			cluster := envoy_common.NewCluster(
				envoy_common.WithName(clusterName),
				envoy_common.WithService(serviceName),
				envoy_common.WithTags(meshDestination.WithoutTags(mesh_proto.ServiceTag)),
			)

			filterChainBuilder := envoy_listeners.NewFilterChainBuilder(apiVersion).Configure(
				envoy_listeners.ServerSideMTLS(meshResources.Mesh),
				envoy_listeners.MatchTransportProtocol("tls"),
				envoy_listeners.MatchServerNames(sni),
				envoy_listeners.NetworkRBAC(
					serviceName,
					// Zone Egress will configure these filter chains only for
					// meshes with mTLS enabled, so we can safely pass here true
					true,
					meshResources.ExternalServicePermissionMap[serviceName],
				),
			)

			protocol := endpoints[0].Tags[mesh_proto.ProtocolTag]

			switch protocol {
			case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
				route := envoy_common.NewRouteFromCluster(cluster)

				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(serviceName, false)).
					Configure(envoy_listeners.HttpOutboundRoute(serviceName, envoy_common.Routes{route}, nil))
			default:
				filterChainBuilder.Configure(
					envoy_listeners.TcpProxyWithMetadata(serviceName, cluster),
				)
			}

			listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
		}
	}
}
