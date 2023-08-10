package egress

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

type InternalServicesGenerator struct{}

// Generate will generate envoy resources for one mesh (when mTLS enabled)
func (g *InternalServicesGenerator) Generate(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	meshResources *core_xds.MeshResources,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	zone := xdsCtx.ControlPlane.Zone
	apiVersion := proxy.APIVersion
	endpointMap := meshResources.EndpointMap
	trafficRoutes := meshResources.Resources[core_mesh.TrafficRouteType].(*core_mesh.TrafficRouteResourceList)
	destinations := buildDestinations(
		trafficRoutes.Items,
		nil,
	)
	services := g.buildServices(endpointMap, meshResources.Mesh.ZoneEgressEnabled(), zone)
	meshName := meshResources.Mesh.GetMeta().GetName()

	g.addFilterChains(
		apiVersion,
		destinations,
		proxy,
		listenerBuilder,
		meshResources,
		services,
	)

	cds, err := g.generateCDS(meshName, apiVersion, services, destinations)
	if err != nil {
		return nil, err
	}
	resources.Add(cds...)

	eds, err := g.generateEDS(meshName, apiVersion, services, endpointMap)
	if err != nil {
		return nil, err
	}
	resources.Add(eds...)

	return resources, nil
}

func (*InternalServicesGenerator) generateEDS(
	meshName string,
	apiVersion core_xds.APIVersion,
	services map[string]bool,
	endpointMap core_xds.EndpointMap,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for serviceName := range services {
		endpoints := endpointMap[serviceName]
		// There is a case where multiple meshes contain services with
		// the same names, so we cannot use just "serviceName" as a cluster
		// name as we would overwrite some clusters with the latest one
		clusterName := envoy_names.GetMeshClusterName(meshName, serviceName)

		cla, err := envoy_endpoints.CreateClusterLoadAssignment(clusterName, endpoints, apiVersion)
		if err != nil {
			return nil, err
		}

		resources = append(resources, &core_xds.Resource{
			Name:     clusterName,
			Origin:   OriginEgress,
			Resource: cla,
		})
	}

	return resources, nil
}

func (*InternalServicesGenerator) generateCDS(
	meshName string,
	apiVersion core_xds.APIVersion,
	services map[string]bool,
	destinationsPerService map[string][]tags.Tags,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for serviceName := range services {
		tagSlice := tags.TagsSlice(append(destinationsPerService[serviceName], destinationsPerService[mesh_proto.MatchAllTag]...))

		tagKeySlice := tagSlice.ToTagKeysSlice().Transform(tags.Without(mesh_proto.ServiceTag), tags.With("mesh"))

		// There is a case where multiple meshes contain services with
		// the same names, so we cannot use just "serviceName" as a cluster
		// name as we would overwrite some clusters with the latest one
		clusterName := envoy_names.GetMeshClusterName(meshName, serviceName)

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
			Origin:   OriginEgress,
			Resource: edsCluster,
		})
	}

	return resources, nil
}

func (*InternalServicesGenerator) buildServices(
	endpointMap core_xds.EndpointMap,
	zoneEgressEnabled bool,
	localZone string,
) map[string]bool {
	services := map[string]bool{}

	for serviceName, endpoints := range endpointMap {
		if len(endpoints) == 0 {
			continue
		}
		internalService := !endpoints[0].IsExternalService()
		zoneExternalService := zoneEgressEnabled && endpoints[0].IsExternalService() && !endpoints[0].IsReachableFromZone(localZone)
		if internalService || zoneExternalService {
			services[serviceName] = true
		}
	}

	return services
}

func (*InternalServicesGenerator) addFilterChains(
	apiVersion core_xds.APIVersion,
	destinationsPerService map[string][]tags.Tags,
	proxy *core_xds.Proxy,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	meshResources *core_xds.MeshResources,
	services map[string]bool,
) {
	meshName := meshResources.Mesh.GetMeta().GetName()

	sniUsed := map[string]bool{}

	for _, zoneIngress := range proxy.ZoneEgressProxy.ZoneIngresses {
		for _, service := range zoneIngress.Spec.GetAvailableServices() {
			serviceName := service.Tags[mesh_proto.ServiceTag]
			if service.Mesh != meshName || !services[serviceName] {
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

				listenerBuilder.Configure(envoy_listeners.FilterChain(
					envoy_listeners.NewFilterChainBuilder(apiVersion, envoy_common.AnonymousResource).Configure(
						envoy_listeners.MatchTransportProtocol("tls"),
						envoy_listeners.MatchServerNames(sni),
						envoy_listeners.TcpProxyDeprecatedWithMetadata(clusterName, envoy_common.NewCluster(
							envoy_common.WithName(clusterName),
							envoy_common.WithService(serviceName),
							envoy_common.WithTags(meshDestination.WithoutTags(mesh_proto.ServiceTag)),
						)),
					),
				))
			}
		}
	}
}
