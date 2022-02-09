package egress

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

type ExternalServicesGenerator struct {
}

// Generate will generate envoy resources for one, provided in ResourceInfo mesh
func (g *ExternalServicesGenerator) Generate(
	_ xds_context.Context,
	info *ResourceInfo,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	apiVersion := info.Proxy.APIVersion
	endpointMap := info.EndpointMap
	destinations := g.buildDestinations(info.TrafficRoutes)
	services := g.buildServices(endpointMap)

	g.addFilterChains(apiVersion, destinations, endpointMap, info)

	cds, err := g.generateCDS(apiVersion, services, endpointMap)
	if err != nil {
		return nil, err
	}
	resources.Add(cds...)

	return resources, nil
}

func (*ExternalServicesGenerator) generateCDS(
	apiVersion envoy_common.APIVersion,
	services []string,
	endpointMap core_xds.EndpointMap,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for _, serviceName := range services {
		endpoints := endpointMap[serviceName]

		if len(endpoints) == 0 {
			log.Info("no endpoints for service", "serviceName", serviceName)

			continue
		}

		clusterBuilder := envoy_clusters.NewClusterBuilder(apiVersion).
			Configure(envoy_clusters.ProvidedEndpointCluster(
				serviceName,
				false, // TODO (bartsmykla)
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

func (ExternalServicesGenerator) buildServices(
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
	info *ResourceInfo,
) {
	meshName := info.Mesh.GetMeta().GetName()
	externalServices := info.ExternalServices
	mesh := info.Mesh

	sniUsed := map[string]bool{}

	xdsContext := xds_context.Context{
		Mesh: xds_context.MeshContext{
			Resource: mesh,
		},
	}

	for _, es := range externalServices {
		serviceName := es.GetMeta().GetName()

		endpoints := endpointMap[serviceName]

		if len(endpoints) == 0 {
			// TODO (bartsmykla): throw warning maybe
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

			cluster := envoy_common.NewCluster(
				envoy_common.WithService(serviceName),
				envoy_common.WithTags(meshDestination.WithoutTags(mesh_proto.ServiceTag)),
			)

			filterChainBuilder := envoy_listeners.NewFilterChainBuilder(apiVersion).Configure(
				envoy_listeners.ServerSideMTLS(xdsContext),
				envoy_listeners.MatchTransportProtocol("tls"),
				envoy_listeners.MatchServerNames(sni),
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

			info.Resources.Listener.Configure(envoy_listeners.FilterChain(filterChainBuilder))
		}
	}
}

func (*ExternalServicesGenerator) buildDestinations(
	trafficRoutes []*core_mesh.TrafficRouteResource,
) map[string][]envoy_common.Tags {
	destinations := map[string][]envoy_common.Tags{}

	for _, tr := range trafficRoutes {
		for _, split := range tr.Spec.Conf.GetSplitWithDestination() {
			service := split.Destination[mesh_proto.ServiceTag]
			destinations[service] = append(destinations[service], split.Destination)
		}

		for _, http := range tr.Spec.Conf.Http {
			for _, split := range http.GetSplitWithDestination() {
				service := split.Destination[mesh_proto.ServiceTag]
				destinations[service] = append(destinations[service], split.Destination)
			}
		}
	}

	return destinations
}
