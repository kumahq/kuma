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
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
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
	resources := core_xds.NewResourceSet()
	apiVersion := proxy.APIVersion
	endpointMap := meshResources.EndpointMap
	localResources := xds_context.Resources{MeshLocalResources: meshResources.Resources}
	destinations := zoneproxy.BuildMeshDestinations(
		nil,
		localResources,
		nil,
		nil,
		localResources.MeshExternalServices().Items,
		"",
		xdsCtx.Mesh.ResolveResourceIdentifier,
	)
	services := g.buildServices(endpointMap)

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
		proxy.Metadata.GetDynamicMetadata(core_xds.FieldSystemCaPath),
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
	systemCaPath string,
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

		clusterBuilder := envoy_clusters.NewClusterBuilder(apiVersion, clusterName)
		isMes := isMeshExternalService(endpoints)

		if isMes {
			clusterBuilder.WithName(serviceName)
			clusterBuilder.
				Configure(envoy_clusters.ProvidedCustomEndpointCluster(isIPV6, isMes, endpoints...)).
				Configure(
					envoy_clusters.MeshExternalServiceClientSideTLS(endpoints, systemCaPath, true),
				)
		} else {
			clusterBuilder.
				Configure(envoy_clusters.ProvidedEndpointCluster(
					isIPV6,
					endpoints...,
				)).
				Configure(envoy_clusters.ClientSideTLS(endpoints))
		}
		clusterBuilder.
			Configure(envoy_clusters.DefaultTimeout())

		switch endpoints[0].Protocol() {
		case core_mesh.ProtocolHTTP:
			clusterBuilder.Configure(envoy_clusters.Http())
		case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
			clusterBuilder.Configure(envoy_clusters.Http2())
		}

		cluster, err := clusterBuilder.Build()
		if err != nil {
			return nil, err
		}

		resource := &core_xds.Resource{
			Name:     serviceName,
			Origin:   OriginEgress,
			Resource: cluster,
		}

		if isMes {
			resource.ResourceOrigin = endpoints[0].ExternalService.OwnerResource
			resource.Protocol = endpoints[0].ExternalService.Protocol
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (*ExternalServicesGenerator) buildServices(
	endpointMap core_xds.EndpointMap,
) map[string]bool {
	services := map[string]bool{}

	for serviceName, endpoints := range endpointMap {
		if len(endpoints) > 0 && endpoints[0].IsExternalService() {
			services[serviceName] = true
		}
	}

	return services
}

func (g *ExternalServicesGenerator) addFilterChains(
	apiVersion core_xds.APIVersion,
	meshDestinations zoneproxy.MeshDestinations,
	endpointMap core_xds.EndpointMap,
	meshResources *core_xds.MeshResources,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	services map[string]bool,
	secretsTracker core_xds.SecretsTracker,
) {
	meshName := meshResources.Mesh.GetMeta().GetName()
	sniUsed := map[string]bool{}
	esNames := []string{}
	for _, es := range meshResources.ExternalServices {
		esNames = append(esNames, es.Spec.GetService())
	}

	for _, esName := range esNames {
		if !services[esName] {
			continue
		}
		endpoints := endpointMap[esName]
		destinations := meshDestinations.KumaIoServices[esName]
		destinations = append(destinations, meshDestinations.KumaIoServices[mesh_proto.MatchAllTag]...)
		for _, destination := range destinations {
			meshDestination := destination.
				WithTags(mesh_proto.ServiceTag, esName).
				WithTags("mesh", meshName)

			sni := tls.SNIFromTags(meshDestination)
			if sniUsed[sni] {
				continue
			}

			sniUsed[sni] = true
			g.configureFilterChain(
				apiVersion,
				esName,
				sni,
				meshName,
				endpoints,
				meshDestination,
				meshResources,
				secretsTracker,
				listenerBuilder,
			)
		}
	}

	for _, mes := range meshDestinations.BackendRefs {
		if !services[mes.DestinationName] {
			return
		}
		endpoints := endpointMap[mes.DestinationName]
		if sniUsed[mes.SNI] {
			continue
		}
		sniUsed[mes.SNI] = true
		relevantTags := tags.Tags{}
		g.configureFilterChain(
			apiVersion,
			mes.DestinationName,
			mes.SNI,
			meshName,
			endpoints,
			relevantTags,
			meshResources,
			secretsTracker,
			listenerBuilder,
		)
	}
}

func (*ExternalServicesGenerator) configureFilterChain(
	apiVersion core_xds.APIVersion,
	esName string,
	sni string,
	meshName string,
	endpoints []core_xds.Endpoint,
	meshDestination tags.Tags,
	meshResources *core_xds.MeshResources,
	secretsTracker core_xds.SecretsTracker,
	listenerBuilder *envoy_listeners.ListenerBuilder,
) {
	// There is a case where multiple meshes contain services with
	// the same names, so we cannot use just "serviceName" as a cluster
	// name as we would overwrite some clusters with the latest one
	clusterName := envoy_names.GetMeshClusterName(meshName, esName)
	if isMeshExternalService(endpoints) {
		clusterName = esName
	}

	cluster := envoy_common.NewCluster(
		envoy_common.WithName(clusterName),
		envoy_common.WithService(esName),
		envoy_common.WithTags(meshDestination.WithoutTags(mesh_proto.ServiceTag)),
		envoy_common.WithExternalService(true),
	)

	filterChainName := names.GetEgressFilterChainName(esName, meshName)
	if isMeshExternalService(endpoints) {
		filterChainName = esName
	}

	filterChainBuilder := envoy_listeners.NewFilterChainBuilder(apiVersion, filterChainName).Configure(
		envoy_listeners.ServerSideMTLS(meshResources.Mesh, secretsTracker, nil, nil),
		envoy_listeners.MatchTransportProtocol("tls"),
		envoy_listeners.MatchServerNames(sni),
		envoy_listeners.NetworkRBAC(
			esName,
			// Zone Egress will configure these filter chains only for
			// meshes with mTLS enabled, so we can safely pass here true
			true,
			meshResources.ExternalServicePermissionMap[esName],
		),
	)
	protocol := endpoints[0].Protocol()

	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		routes := envoy_common.Routes{}

		for _, rl := range meshResources.ExternalServiceRateLimits[esName] {
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

		routeConfigName := envoy_names.GetOutboundRouteName(esName)
		if isMeshExternalService(endpoints) {
			routeConfigName = esName
		}

		filterChainBuilder.
			Configure(envoy_listeners.HttpConnectionManager(esName, false)).
			Configure(envoy_listeners.FaultInjection(meshResources.ExternalServiceFaultInjections[esName]...)).
			Configure(envoy_listeners.RateLimit(meshResources.ExternalServiceRateLimits[esName])).
			Configure(envoy_listeners.AddFilterChainConfigurer(&v3.HttpOutboundRouteConfigurer{
				Name:    routeConfigName,
				Service: esName,
				Routes:  routes,
				DpTags:  nil,
			}))
	default:
		filterChainBuilder.Configure(
			envoy_listeners.TcpProxyDeprecatedWithMetadata(esName, cluster),
		)
	}
	listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
}

func isMeshExternalService(endpoints []core_xds.Endpoint) bool {
	if len(endpoints) > 0 {
		return endpoints[0].IsMeshExternalService()
	}
	return false
}
