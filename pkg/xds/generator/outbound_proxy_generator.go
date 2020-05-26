package generator

import (
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
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	meshName := ctx.Mesh.Resource.GetMeta().GetName()
	ofaces, err := proxy.Dataplane.Spec.Networking.GetOutboundInterfaces()
	if err != nil {
		return nil, err
	}
	for i, outbound := range outbounds {
		// pick a route
		route := proxy.TrafficRoutes[outbound.Service]
		if route == nil {
			return nil, errors.Errorf("%s{service=%q}: has no TrafficRoute", validators.RootedAt("dataplane").Field("networking").Field("outbound").Index(i), outbound.Service)
		}

		// determine the list of destination clusters
		clusters, err := g.determineClusters(ctx, proxy, route)
		if err != nil {
			return nil, err
		}

		// generate CDS and EDS resources
		edsResources, endpoints, err := g.generateEds(ctx, proxy, clusters)
		if err != nil {
			return nil, err
		}
		resources.Add(edsResources...)

		protocol := InferServiceProtocol(endpoints)

		// generate LDS resource
		outboundListenerName := envoy_names.GetOutboundListenerName(ofaces[i].DataplaneIP, ofaces[i].DataplanePort)
		outboundRouteName := envoy_names.GetOutboundRouteName(outbound.Service)
		destinationService := outbound.Service
		trafficDirection := "OUTBOUND"
		filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
			filterChainBuilder := envoy_listeners.NewFilterChainBuilder()
			switch protocol {
			case mesh_core.ProtocolHTTP:
				// configuration for HTTP case
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(outbound.Service)).
					Configure(envoy_listeners.Tracing(proxy.TracingBackend)).
					Configure(envoy_listeners.HttpAccessLog(meshName, trafficDirection, sourceService, destinationService, proxy.Logs[outbound.Service], proxy)).
					Configure(envoy_listeners.HttpOutboundRoute(outboundRouteName))
			case mesh_core.ProtocolTCP:
				fallthrough
			default:
				// configuration for non-HTTP cases
				filterChainBuilder.
					Configure(envoy_listeners.TcpProxy(outbound.Service, clusters...)).
					Configure(envoy_listeners.NetworkAccessLog(meshName, trafficDirection, sourceService, destinationService, proxy.Logs[outbound.Service], proxy))
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
		rdsResources, err := g.generateRds(protocol, outbound.Service, outboundRouteName, clusters, proxy.Dataplane.Spec.Tags())
		if err != nil {
			return nil, err
		}
		resources.Add(rdsResources...)
	}

	return resources.List(), nil
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
			Name:   envoy_names.GetDestinationClusterName(service, destination.Destination),
			Weight: destination.Weight,
			Tags:   destination.Destination,
		})
	}
	return
}

func (_ OutboundProxyGenerator) generateEds(ctx xds_context.Context, proxy *model.Proxy, clusters []envoy_common.ClusterInfo) (resources []*model.Resource, allEndpoints []model.Endpoint, _ error) {
	for _, cluster := range clusters {
		serviceName := cluster.Tags[kuma_mesh.ServiceTag]
		healthCheck := proxy.HealthChecks[serviceName]
		edsCluster, err := envoy_clusters.NewClusterBuilder().
			Configure(envoy_clusters.EdsCluster(cluster.Name)).
			Configure(envoy_clusters.ClientSideMTLS(ctx, proxy.Metadata, cluster.Tags[kuma_mesh.ServiceTag])).
			Configure(envoy_clusters.HealthCheck(healthCheck)).
			Build()
		if err != nil {
			return nil, nil, err
		}
		resources = append(resources, &model.Resource{
			Name:     cluster.Name,
			Resource: edsCluster,
		})
		endpoints := model.EndpointList(proxy.OutboundTargets[serviceName]).Filter(cluster.Tags)
		resources = append(resources, &model.Resource{
			Name:     cluster.Name,
			Resource: envoy_endpoints.CreateClusterLoadAssignment(cluster.Name, endpoints),
		})
		allEndpoints = append(allEndpoints, endpoints...)
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
