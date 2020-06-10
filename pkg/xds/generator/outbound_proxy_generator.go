package generator

import (
	"sort"

	"github.com/Kong/kuma/pkg/core/validators"
	envoy_endpoints "github.com/Kong/kuma/pkg/xds/envoy/endpoints"
	envoy_names "github.com/Kong/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/Kong/kuma/pkg/xds/envoy/routes"

	"github.com/pkg/errors"

	kuma_mesh "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	model "github.com/Kong/kuma/pkg/core/xds"
	xds_context "github.com/Kong/kuma/pkg/xds/context"

	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/Kong/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/Kong/kuma/pkg/xds/envoy/listeners"
)

type OutboundProxyGenerator struct {
}

type Clusters map[string][]envoy_common.ClusterInfo

func (c Clusters) ClusterNames() []string {
	var keys []string
	for key := range c {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c Clusters) Add(infos ...envoy_common.ClusterInfo) {
	for _, info := range infos {
		c[info.Name] = append(c[info.Name], info)
	}
}

func (c Clusters) Tags(name string) []envoy_common.Tags {
	var result []envoy_common.Tags
	for _, info := range c[name] {
		result = append(result, info.Tags)
	}
	return result
}

func (g OutboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	outbounds := proxy.Dataplane.Spec.Networking.GetOutbound()
	if len(outbounds) == 0 {
		return nil, nil
	}
	resources := &model.ResourceSet{}
	clusters := Clusters{}
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()
	ofaces, err := proxy.Dataplane.Spec.Networking.GetOutboundInterfaces()
	if err != nil {
		return nil, err
	}

	for i, outbound := range outbounds {
		oface, err := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		if err != nil {
			// ignore invalid outbounds
			continue
		}
		// pick a route
		route := proxy.TrafficRoutes[oface]
		if route == nil { // should not happen since we always generate default route if TrafficRoute is not found
			return nil, errors.Errorf("%s{service=%q}: has no TrafficRoute", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i), oface)
		}

		// determine the list of destination clusters
		outboundClusters, err := g.determineClusters(ctx, proxy, route)
		if err != nil {
			return nil, err
		}
		clusters.Add(outboundClusters...)

		protocol := g.inferProtocol(proxy, outboundClusters)

		// generate LDS resource
		serviceName := outbound.GetTagsIncludingLegacy()[kuma_mesh.ServiceTag]
		outboundListenerName := envoy_names.GetOutboundListenerName(ofaces[i].DataplaneIP, ofaces[i].DataplanePort)
		outboundRouteName := envoy_names.GetOutboundRouteName(serviceName)
		destinationService := serviceName
		filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
			filterChainBuilder := envoy_listeners.NewFilterChainBuilder()
			switch protocol {
			case mesh_core.ProtocolHTTP:
				// configuration for HTTP case
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(serviceName)).
					Configure(envoy_listeners.Tracing(proxy.TracingBackend)).
					Configure(envoy_listeners.HttpAccessLog(meshName, envoy_listeners.TrafficDirectionOutbound, sourceService, destinationService, proxy.Logs[serviceName], proxy)).
					Configure(envoy_listeners.HttpOutboundRoute(outboundRouteName))
			case mesh_core.ProtocolTCP:
				fallthrough
			default:
				// configuration for non-HTTP cases
				filterChainBuilder.
					Configure(envoy_listeners.TcpProxy(serviceName, outboundClusters...)).
					Configure(envoy_listeners.NetworkAccessLog(meshName, envoy_listeners.TrafficDirectionOutbound, sourceService, destinationService, proxy.Logs[serviceName], proxy))
			}
			return filterChainBuilder
		}()
		listener, err := envoy_listeners.NewListenerBuilder().
			Configure(envoy_listeners.OutboundListener(outboundListenerName, ofaces[i].DataplaneIP, ofaces[i].DataplanePort)).
			Configure(envoy_listeners.FilterChain(filterChainBuilder)).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %s", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i), outboundListenerName)
		}
		resources.AddNamed(listener)

		// generate RDS resources
		rdsResources, err := g.generateRds(protocol, serviceName, outboundRouteName, outboundClusters, proxy.Dataplane.Spec.Tags())
		if err != nil {
			return nil, err
		}
		resources.Add(rdsResources...)
	}

	cdsResources, err := g.generateCDS(ctx, proxy, clusters)
	if err != nil {
		return nil, err
	}
	resources.AddSet(cdsResources)

	edsResources := g.generateEDS(proxy, clusters)
	resources.AddSet(edsResources)

	return resources.List(), nil
}

func (_ OutboundProxyGenerator) generateCDS(ctx xds_context.Context, proxy *model.Proxy, clusters Clusters) (model.ResourceSet, error) {
	var resources model.ResourceSet
	for _, clusterName := range clusters.ClusterNames() {
		serviceName := clusters.Tags(clusterName)[0][kuma_mesh.ServiceTag]
		healthCheck := proxy.HealthChecks[serviceName]
		circuitBreaker := proxy.CircuitBreakers[serviceName]
		edsCluster, err := envoy_clusters.NewClusterBuilder().
			Configure(envoy_clusters.EdsCluster(clusterName)).
			Configure(envoy_clusters.LbSubset(clusters.Tags(clusterName))).
			Configure(envoy_clusters.ClientSideMTLSWithSNI(ctx, proxy.Metadata, serviceName, clusterName)).
			Configure(envoy_clusters.OutlierDetection(circuitBreaker)).
			Configure(envoy_clusters.HealthCheck(healthCheck)).
			Build()
		if err != nil {
			return model.ResourceSet{}, err
		}
		resources.AddNamed(edsCluster)
	}
	return resources, nil
}

func (_ OutboundProxyGenerator) generateEDS(proxy *model.Proxy, clusters Clusters) model.ResourceSet {
	var resources model.ResourceSet
	for _, clusterName := range clusters.ClusterNames() {
		serviceName := clusters.Tags(clusterName)[0][kuma_mesh.ServiceTag]
		endpoints := model.EndpointList(proxy.OutboundTargets[serviceName])
		loadAssignment := envoy_endpoints.CreateClusterLoadAssignment(clusterName, endpoints)
		resources.Add(&model.Resource{
			Name:     clusterName,
			Resource: loadAssignment,
		})
	}
	return resources
}

func (_ OutboundProxyGenerator) inferProtocol(proxy *model.Proxy, clusters []envoy_common.ClusterInfo) mesh_core.Protocol {
	var allEndpoints []model.Endpoint
	for _, cluster := range clusters {
		serviceName := cluster.Tags[kuma_mesh.ServiceTag]
		endpoints := model.EndpointList(proxy.OutboundTargets[serviceName])
		allEndpoints = append(allEndpoints, endpoints...)
	}
	return InferServiceProtocol(allEndpoints)
}

func (_ OutboundProxyGenerator) determineClusters(ctx xds_context.Context, proxy *model.Proxy, route *mesh_core.TrafficRouteResource) (clusters []envoy_common.ClusterInfo, err error) {
	for j, destination := range route.Spec.Conf {
		service, ok := destination.Destination[kuma_mesh.ServiceTag]
		if !ok {
			return nil, errors.Errorf("trafficroute{name=%q}.%s: mandatory tag %q is missing: %v", route.GetMeta().GetName(), validators.RootedAt("conf").Index(j).Field("destination"), kuma_mesh.ServiceTag, destination.Destination)
		}
		if destination.Weight == 0 {
			// Envoy doesn't support 0 weight
			continue
		}
		clusters = append(clusters, envoy_common.ClusterInfo{
			Name:   service,
			Weight: destination.Weight,
			Tags:   destination.Destination,
		})
	}
	return
}

func (_ OutboundProxyGenerator) generateRds(protocol mesh_core.Protocol, service string, outboundRouteName string, clusters []envoy_common.ClusterInfo, tags kuma_mesh.MultiValueTagSet) ([]*model.Resource, error) {
	resources := &model.ResourceSet{}
	switch protocol {
	case mesh_core.ProtocolHTTP:
		// generate RDS resource
		routeConfiguration, err := envoy_routes.NewRouteConfigurationBuilder().
			Configure(envoy_routes.CommonRouteConfiguration(outboundRouteName)).
			Configure(envoy_routes.TagsHeader(tags)).
			Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder().
				Configure(envoy_routes.CommonVirtualHost(service)).
				Configure(envoy_routes.DefaultRoute(clusters...)))).
			Build()
		if err != nil {
			return nil, err
		}
		resources.AddNamed(routeConfiguration)
	}
	return resources.List(), nil
}
