package generator

import (
	"strings"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

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

func (g OutboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) ([]*model.Resource, error) {
	outbounds := proxy.Dataplane.Spec.Networking.GetOutbound()
	if len(outbounds) == 0 {
		return nil, nil
	}
	resources := &model.ResourceSet{}
	clusters := envoy_common.Clusters{}

	for _, outbound := range outbounds {
		// Determine the list of destination subsets
		// For one outbound listener it may contain many subsets (ex. TrafficRoute to many destinations)
		subsets, err := g.determineSubsets(proxy, outbound)
		if err != nil {
			return nil, err
		}
		clusters.Add(subsets...)

		protocol := g.inferProtocol(proxy, subsets)

		// Generate listener
		listener, err := g.generateLDS(proxy, subsets, outbound, protocol)
		if err != nil {
			return nil, err
		}
		resources.AddNamed(listener)

		// Generate route, routes are only applicable to the HTTP
		if protocol == mesh_core.ProtocolHTTP {
			route, err := g.generateRDS(proxy, subsets, outbound)
			if err != nil {
				return nil, err
			}
			resources.AddNamed(route)
		}
	}

	// Generate clusters. It cannot be generated on the fly with outbound loop because we need to know all subsets of the cluster for every service.
	cdsResources, err := g.generateCDS(ctx, proxy, clusters)
	if err != nil {
		return nil, err
	}
	resources.AddSet(cdsResources)

	edsResources := g.generateEDS(proxy, clusters)
	resources.AddSet(edsResources)

	return resources.List(), nil
}

func (_ OutboundProxyGenerator) generateLDS(proxy *model.Proxy, subsets []envoy_common.ClusterSubset, outbound *kuma_mesh.Dataplane_Networking_Outbound, protocol mesh_core.Protocol) (*envoy_api_v2.Listener, error) {
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
	meshName := proxy.Dataplane.Meta.GetMesh()
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	serviceName := outbound.GetTagsIncludingLegacy()[kuma_mesh.ServiceTag]
	outboundListenerName := envoy_names.GetOutboundListenerName(oface.DataplaneIP, oface.DataplanePort)
	filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
		filterChainBuilder := envoy_listeners.NewFilterChainBuilder()
		switch protocol {
		case mesh_core.ProtocolHTTP:
			// configuration for HTTP case
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName)).
				Configure(envoy_listeners.Tracing(proxy.TracingBackend)).
				Configure(envoy_listeners.HttpAccessLog(meshName, envoy_listeners.TrafficDirectionOutbound, sourceService, serviceName, proxy.Logs[serviceName], proxy)).
				Configure(envoy_listeners.HttpOutboundRoute(envoy_names.GetOutboundRouteName(serviceName)))
		case mesh_core.ProtocolTCP:
			fallthrough
		default:
			// configuration for non-HTTP cases
			filterChainBuilder.
				Configure(envoy_listeners.TcpProxy(serviceName, subsets...)).
				Configure(envoy_listeners.NetworkAccessLog(meshName, envoy_listeners.TrafficDirectionOutbound, sourceService, serviceName, proxy.Logs[serviceName], proxy))
		}
		return filterChainBuilder
	}()
	listener, err := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.OutboundListener(outboundListenerName, oface.DataplaneIP, oface.DataplanePort)).
		Configure(envoy_listeners.FilterChain(filterChainBuilder)).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener %s for service %s", outboundListenerName, serviceName)
	}
	return listener, nil
}

func (o OutboundProxyGenerator) generateCDS(ctx xds_context.Context, proxy *model.Proxy, clusters envoy_common.Clusters) (model.ResourceSet, error) {
	var resources model.ResourceSet
	for _, clusterName := range clusters.ClusterNames() {
		serviceName := clusters.Tags(clusterName)[0][kuma_mesh.ServiceTag]
		tags := clusters.Tags(clusterName)
		healthCheck := proxy.HealthChecks[serviceName]
		circuitBreaker := proxy.CircuitBreakers[serviceName]
		edsCluster, err := envoy_clusters.NewClusterBuilder().
			Configure(envoy_clusters.EdsCluster(clusterName)).
			Configure(envoy_clusters.LbSubset(o.lbSubsets(tags))).
			Configure(envoy_clusters.ClientSideMTLS(ctx, proxy.Metadata, serviceName, tags)).
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

func (_ OutboundProxyGenerator) lbSubsets(tagSets []envoy_common.Tags) [][]string {
	var result [][]string
	uniqueKeys := map[string]bool{}
	for _, tags := range tagSets {
		keys := tags.WithoutTag(kuma_mesh.ServiceTag).Keys()
		joined := strings.Join(keys, ",")
		if !uniqueKeys[joined] {
			uniqueKeys[joined] = true
			result = append(result, keys)
		}
	}
	return result
}

func (_ OutboundProxyGenerator) generateEDS(proxy *model.Proxy, clusters envoy_common.Clusters) model.ResourceSet {
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

// inferProtocol infers protocol for the destination listener. It will only return HTTP when all endpoints are tagged with HTTP.
func (_ OutboundProxyGenerator) inferProtocol(proxy *model.Proxy, clusters []envoy_common.ClusterSubset) mesh_core.Protocol {
	var allEndpoints []model.Endpoint
	for _, cluster := range clusters {
		serviceName := cluster.Tags[kuma_mesh.ServiceTag]
		endpoints := model.EndpointList(proxy.OutboundTargets[serviceName])
		allEndpoints = append(allEndpoints, endpoints...)
	}
	return InferServiceProtocol(allEndpoints)
}

func (_ OutboundProxyGenerator) determineSubsets(proxy *model.Proxy, outbound *kuma_mesh.Dataplane_Networking_Outbound) (subsets []envoy_common.ClusterSubset, err error) {
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
	route := proxy.TrafficRoutes[oface]
	if route == nil { // should not happen since we always generate default route if TrafficRoute is not found
		return nil, errors.Errorf("no TrafficRoute for outbound %s", oface)
	}

	for j, destination := range route.Spec.Conf {
		service, ok := destination.Destination[kuma_mesh.ServiceTag]
		if !ok { // should not happen since we validate traffic route
			return nil, errors.Errorf("trafficroute{name=%q}.%s: mandatory tag %q is missing: %v", route.GetMeta().GetName(), validators.RootedAt("conf").Index(j).Field("destination"), kuma_mesh.ServiceTag, destination.Destination)
		}
		if destination.Weight == 0 {
			// 0 assumes no traffic is passed there. Envoy doesn't support 0 weight, so instead of passing it to Envoy we just skip such cluster.
			continue
		}
		subsets = append(subsets, envoy_common.ClusterSubset{
			ClusterName: service,
			Weight:      destination.Weight,
			Tags:        destination.Destination,
		})
	}
	return
}

func (_ OutboundProxyGenerator) generateRDS(proxy *model.Proxy, subsets []envoy_common.ClusterSubset, outbound *kuma_mesh.Dataplane_Networking_Outbound) (*envoy_api_v2.RouteConfiguration, error) {
	serviceName := outbound.GetTagsIncludingLegacy()[kuma_mesh.ServiceTag]

	return envoy_routes.NewRouteConfigurationBuilder().
		Configure(envoy_routes.CommonRouteConfiguration(envoy_names.GetOutboundRouteName(serviceName))).
		Configure(envoy_routes.TagsHeader(proxy.Dataplane.Spec.Tags())).
		Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder().
			Configure(envoy_routes.CommonVirtualHost(serviceName)).
			Configure(envoy_routes.DefaultRoute(subsets...)))).
		Build()
}
