package egress

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
	"github.com/kumahq/kuma/pkg/xds/generator/zoneproxy"
)

type ExternalServicesGenerator struct{}

// Generate will generate envoy resources for one mesh (when mTLS enabled)
func (g *ExternalServicesGenerator) Generate(
	ctx context.Context,
	xdsCtx xds_context.Context,
	proxy *core_xds.Proxy,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	meshResources *core_xds.MeshResources,
) (*core_xds.ResourceSet, error) {
	zone := xdsCtx.ControlPlane.Zone
	resources := core_xds.NewResourceSet()
	apiVersion := proxy.APIVersion
	endpointMap := meshResources.EndpointMap
	destinations := zoneproxy.BuildMeshDestinations(
		nil,
		xds_context.Resources{MeshLocalResources: meshResources.Resources},
	)
	services := g.buildServices(endpointMap, zone, meshResources)

	g.addFilterChains(
		apiVersion,
		destinations,
		endpointMap,
		meshResources,
		listenerBuilder,
		services,
		proxy.SecretsTracker,
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
	apiVersion core_xds.APIVersion,
	services map[string]bool,
	endpointMap core_xds.EndpointMap,
	isIPV6 bool,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for serviceName := range services {
		endpoints := endpointMap[serviceName]

		if len(endpoints) == 0 {
			log.Info("no endpoints for service", "serviceName", serviceName)

			continue
		}

		// There is a case where multiple meshes contain services with
		// the same names, so we cannot use just "serviceName" as a cluster
		// name as we would overwrite some clusters with the latest one
		clusterName := envoy_names.GetMeshClusterName(meshName, serviceName)

		clusterBuilder := envoy_clusters.NewClusterBuilder(apiVersion, clusterName).
			Configure(envoy_clusters.ProvidedEndpointCluster(
				isIPV6,
				endpoints...,
			)).
			Configure(envoy_clusters.ClientSideTLS(endpoints)).
			Configure(envoy_clusters.DefaultTimeout())

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
			Name:     cluster.GetName(),
			Origin:   OriginEgress,
			Resource: cluster,
		})
	}

	return resources, nil
}

func (*ExternalServicesGenerator) buildServices(
	endpointMap core_xds.EndpointMap,
	localZone string,
	meshResources *core_xds.MeshResources,
) map[string]bool {
	services := map[string]bool{}

	for serviceName, endpoints := range endpointMap {
		if len(endpoints) > 0 && endpoints[0].IsExternalService() &&
			(!meshResources.Mesh.ZoneEgressEnabled() || endpoints[0].IsReachableFromZone(localZone)) {
			services[serviceName] = true
		}
	}

	return services
}

func (g *ExternalServicesGenerator) addFilterChains(
	apiVersion core_xds.APIVersion,
	destinationsPerService map[string][]tags.Tags,
	endpointMap core_xds.EndpointMap,
	meshResources *core_xds.MeshResources,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	services map[string]bool,
	secretsTracker core_xds.SecretsTracker,
) {
	meshName := meshResources.Mesh.GetMeta().GetName()
	sniUsed := map[string]bool{}

	for _, es := range meshResources.ExternalServices {
		serviceName := es.Spec.GetService()
		if !services[serviceName] {
			continue
		}

		endpoints := endpointMap[serviceName]
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
				envoy_common.WithExternalService(true),
			)

			filterChainBuilder := envoy_listeners.NewFilterChainBuilder(apiVersion, names.GetEgressFilterChainName(serviceName, meshName)).Configure(
				envoy_listeners.ServerSideMTLS(meshResources.Mesh, secretsTracker),
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
				routes := envoy_common.Routes{}

				for _, rl := range meshResources.ExternalServiceRateLimits[serviceName] {
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

				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(serviceName, false)).
					Configure(envoy_listeners.FaultInjection(meshResources.ExternalServiceFaultInjections[serviceName]...)).
					Configure(envoy_listeners.RateLimit(meshResources.ExternalServiceRateLimits[serviceName])).
					Configure(envoy_listeners.HttpOutboundRoute(serviceName, routes, nil))
			default:
				filterChainBuilder.Configure(
					envoy_listeners.TcpProxyDeprecatedWithMetadata(serviceName, cluster),
				)
			}

			listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
		}
	}
}
