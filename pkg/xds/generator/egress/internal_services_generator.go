package egress

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

type InternalServicesGenerator struct {
}

// Generate will generate envoy resources for one, provided in ResourceInfo mesh
func (g *InternalServicesGenerator) Generate(
	_ xds_context.Context,
	info *ResourceInfo,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	apiVersion := info.Proxy.APIVersion
	endpointMap := info.EndpointMap
	destinations := g.buildDestinations(info.TrafficRoutes)
	services := g.buildServices(endpointMap)

	log.Info("endpointMap", "endpointMap", endpointMap)
	log.Info("destinations", "destinations", destinations)
	log.Info("services", "services", services)

	g.addFilterChains(apiVersion, destinations, endpointMap, info)

	cds, err := g.generateCDS(apiVersion, services, destinations)
	if err != nil {
		return nil, err
	}
	resources.Add(cds...)

	eds, err := g.generateEDS(apiVersion, services, endpointMap)
	if err != nil {
		return nil, err
	}
	resources.Add(eds...)

	return resources, nil
}

func (*InternalServicesGenerator) generateEDS(
	apiVersion envoy_common.APIVersion,
	services []string,
	endpointMap core_xds.EndpointMap,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for _, serviceName := range services {
		endpoints := endpointMap[serviceName]

		cla, err := envoy_endpoints.CreateClusterLoadAssignment(serviceName, endpoints, apiVersion)
		if err != nil {
			return nil, err
		}

		resources = append(resources, &core_xds.Resource{
			Name:     serviceName,
			Origin:   OriginEgress,
			Resource: cla,
		})
	}

	return resources, nil
}

func (*InternalServicesGenerator) generateCDS(
	apiVersion envoy_common.APIVersion,
	services []string,
	destinationsPerService map[string][]envoy_common.Tags,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for _, serviceName := range services {
		tagSlice := envoy_common.TagsSlice(append(destinationsPerService[serviceName], destinationsPerService[mesh_proto.MatchAllTag]...))

		tagKeySlice := tagSlice.ToTagKeysSlice().Transform(envoy_common.Without(mesh_proto.ServiceTag), envoy_common.With("mesh"))

		edsCluster, err := envoy_clusters.NewClusterBuilder(apiVersion).
			Configure(envoy_clusters.EdsCluster(serviceName)).
			Configure(envoy_clusters.LbSubset(tagKeySlice)).
			Configure(envoy_clusters.DefaultTimeout()).
			Build()

		if err != nil {
			return nil, err
		}

		resources = append(resources, &core_xds.Resource{
			Name:     serviceName,
			Origin:   OriginEgress,
			Resource: edsCluster,
		})
	}

	return resources, nil
}

func (InternalServicesGenerator) buildServices(
	endpointMap core_xds.EndpointMap,
) []string {
	var services []string

	for serviceName, endpoints := range endpointMap {
		if len(endpoints) > 0 && !endpoints[0].IsExternalService() {
			services = append(services, serviceName)
		}
	}

	sort.Strings(services)

	return services
}

func (*InternalServicesGenerator) addFilterChains(
	apiVersion envoy_common.APIVersion,
	destinationsPerService map[string][]envoy_common.Tags,
	endpointMap core_xds.EndpointMap,
	info *ResourceInfo,
) {
	zoneIngresses := info.ZoneIngresses
	meshName := info.Mesh.GetMeta().GetName()

	sniUsed := map[string]bool{}

	for _, zoneIngress := range zoneIngresses {
		for _, service := range zoneIngress.Spec.GetAvailableServices() {
			serviceName := service.Tags[mesh_proto.ServiceTag]
			if service.Mesh != meshName {
				continue
			}

			endpoints := endpointMap[serviceName]

			if len(endpoints) == 0 {
				// There is no need to generate filter chain if there is no
				// endpoints for the service
				continue
			}

			if endpoints[0].IsExternalService() {
				// This generator is for internal services only
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

				info.Resources.Listener.Configure(envoy_listeners.FilterChain(
					envoy_listeners.NewFilterChainBuilder(apiVersion).Configure(
						envoy_listeners.MatchTransportProtocol("tls"),
						envoy_listeners.MatchServerNames(sni),
						envoy_listeners.TcpProxyWithMetadata(serviceName, envoy_common.NewCluster(
							envoy_common.WithService(serviceName),
							envoy_common.WithTags(meshDestination.WithoutTags(mesh_proto.ServiceTag)),
						)),
					),
				))
			}
		}
	}
}

func (*InternalServicesGenerator) buildDestinations(
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
