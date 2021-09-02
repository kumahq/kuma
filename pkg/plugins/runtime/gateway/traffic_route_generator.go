package gateway

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/runtime/gateway/merge"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func selectTrafficRoutes(in []model.Resource) []*core_mesh.TrafficRouteResource {
	routes := make([]*core_mesh.TrafficRouteResource, 0, len(in))

	for _, r := range in {
		if trafficRoute, ok := r.(*core_mesh.TrafficRouteResource); ok {
			routes = append(routes, trafficRoute)
		}
	}

	return routes
}

// TrafficRouteGenerator generates Kuma gateway routes from TrafficRoute resources.
type TrafficRouteGenerator struct{}

func (*TrafficRouteGenerator) SupportsProtocol(p mesh_proto.Gateway_Listener_Protocol) bool {
	switch p {
	case mesh_proto.Gateway_Listener_HTTP,
		mesh_proto.Gateway_Listener_HTTPS:
		return true
	default:
		return false
	}
}

func (*TrafficRouteGenerator) GenerateHost(ctx xds_context.Context, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	resources := ResourceAggregator{}

	trafficRoute := merge.TrafficRoute(selectTrafficRoutes(info.Host.Routes)...)
	if trafficRoute == nil {
		return nil, nil
	}

	log.V(1).Info("applying merged traffic routes",
		"listener-port", info.Listener.Port,
		"listener-name", info.Listener.ResourceName,
	)

	services := envoy.Services{}
	routes := generateRoutes(trafficRoute, info)

	services.Add(routes.Clusters()...)

	// Attach these routes to the current host.
	info.Resources.VirtualHost.Configure(envoy_routes.Routes(routes))

	// Build the clusters.
	if err := resources.Add(generateClusters(ctx, services, info)); err != nil {
		return nil, err
	}

	// Build the cluster load assignments.
	if err := resources.Add(generateEndpoints(ctx, services)); err != nil {
		return nil, err
	}

	return resources.Get(), nil
}

// XXX See OutboundProxyGenerator ...
func generateEndpoints(ctx xds_context.Context, services envoy.Services) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	for _, serviceName := range services.Names() {
		// Endpoints for ExternalServices are specified in load
		// assignment in DNS Cluster. We are not allowed to add
		// endpoints with DNS names through EDS.
		if services[serviceName].HasExternalService() {
			continue
		}

		for _, cluster := range services[serviceName].Clusters() {
			loadAssignment, err := ctx.ControlPlane.CLACache.GetCLA(
				context.Background(), ctx.Mesh.Resource.Meta.GetName(), ctx.Mesh.Hash, cluster, envoy.APIV3)
			if err != nil {
				return nil, errors.Wrapf(err, "could not get ClusterLoadAssignment for %s", serviceName)
			}

			resources.Add(&core_xds.Resource{
				Name:     cluster.Name(),
				Origin:   OriginGateway,
				Resource: loadAssignment,
			})
		}
	}

	return resources, nil
}

// XXX See OutboundProxyGenerator ...
func generateClusters(_ xds_context.Context, services envoy.Services, info *GatewayResourceInfo) (*core_xds.ResourceSet, error) {
	resources := ResourceAggregator{}

	for _, serviceName := range services.Names() {
		service := services[serviceName]

		// TODO(jpeach) these policies should be matched to the
		// listener. The xDS server machinery will match them to
		// the Dataplane, which is not what we need.

		var healthCheck *core_mesh.HealthCheckResource       // TODO(jpeach) proxy.Policies.HealthChecks[serviceName]
		var circuitBreaker *core_mesh.CircuitBreakerResource // TODO(jpeach) proxy.Policies.CircuitBreakers[serviceName]

		for _, cluster := range service.Clusters() {
			// Infer the protocol we should use for this cluster by
			// cross-checking the protocol tags of each matching endpoint.
			clusterService := cluster.Tags()[mesh_proto.ServiceTag]
			protocol := generator.InferServiceProtocol(info.Proxy.Routing.OutboundTargets[clusterService])

			edsClusterBuilder := envoy_clusters.NewClusterBuilder(envoy.APIV3).
				Configure(
					envoy_clusters.Timeout(protocol, cluster.Timeout()),
					envoy_clusters.CircuitBreaker(circuitBreaker),
					envoy_clusters.OutlierDetection(circuitBreaker),
					envoy_clusters.HealthCheck(protocol, healthCheck),
				)

			if service.HasExternalService() {
				/* TODO(jpeach)
				edsClusterBuilder.
					Configure(envoy_clusters.StrictDNSCluster(cluster.Name(), proxy.Routing.OutboundTargets[serviceName],
						proxy.Dataplane.IsIPv6())).
					Configure(envoy_clusters.ClientSideTLS(proxy.Routing.OutboundTargets[serviceName]))
				*/
			} else {
				edsClusterBuilder.
					Configure(
						envoy_clusters.EdsCluster(cluster.Name()),
						envoy_clusters.LB(cluster.LB()),
						// TODO(jpeach) envoy_clusters.ClientSideMTLS(...)
					)
			}

			switch protocol {
			case core_mesh.ProtocolHTTP:
				edsClusterBuilder.Configure(envoy_clusters.Http())
			case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
				edsClusterBuilder.Configure(envoy_clusters.Http2())
			}

			if err := resources.Add(BuildResourceSet(edsClusterBuilder)); err != nil {
				return nil, errors.Wrapf(err, "build CDS for cluster %s failed", cluster.Name())
			}
		}
	}

	return resources.Get(), nil
}

// XXX See OutboundProxyGenerator ...
func generateRoutes(route *core_mesh.TrafficRouteResource, info *GatewayResourceInfo) envoy.Routes {
	routes := envoy.Routes{}
	clusterCache := map[string]string{}

	// TODO(jpeach) refactor this code so that it can be shared with OutboundProxyGenerator.
	clustersFromSplit := func(splits []*mesh_proto.TrafficRoute_Split) []envoy.Cluster {
		var clusters []envoy.Cluster
		for _, destination := range splits {
			service := destination.GetDestination()[mesh_proto.ServiceTag]

			// '*' isn't a real service destination, so we
			// can't generate a working cluster from it. This
			// check catches the default TrafficRoute policy.
			if service == mesh_proto.MatchAllTag {
				continue
			}

			if destination.GetWeight().GetValue() == 0 {
				// 0 assumes no traffic is passed there. Envoy doesn't
				// support 0 weight, so instead of passing it to Envoy
				// we just skip such cluster.
				continue
			}

			name := service
			if len(destination.GetDestination()) > 1 {
				name = envoy_names.GetSplitClusterName(service, info.Resources.Discriminator())
			}

			// TODO(jpeach) This is wrong for gateway. Each
			// destination in the Traffic Route can be an
			// arbitrary service and possibly external.

			/*
				// We assume that all the targets are either ExternalServices or not
				// therefore we check only the first one
				isExternalService := false
				if endpoints := proxy.Routing.OutboundTargets[service]; len(endpoints) > 0 {
					ep := endpoints[0]
					if ep.IsExternalService() {
						isExternalService = true
					}
				}

			*/

			// TODO(jpeach) The timeoutConf for this cluster should be the
			// Timeout policy that matches the listener source tags and the
			// current split destination.
			//
			// However, we don't know the listener tags any more, because
			// we already collapsed the listeners to generate the list of
			// routes for each listener. And then we merged all the routes
			// into a single TrafficRoute resource.

			cluster := envoy.NewCluster(
				envoy.WithService(service),
				envoy.WithName(name),
				envoy.WithWeight(destination.GetWeight().GetValue()),
				envoy.WithTags(destination.GetDestination()),
				envoy.WithLB(route.Spec.GetConf().GetLoadBalancer()),
				/* TODO(jpeach) envoy.WithTimeout(timeoutConf), */
				/* TODO(jpeach) envoy.WithExternalService(isExternalService), */
			)

			if name, ok := clusterCache[cluster.Tags().String()]; ok {
				cluster.SetName(name)
			} else {
				clusterCache[cluster.Tags().String()] = cluster.Name()
			}

			clusters = append(clusters, cluster)
		}

		return clusters
	}

	// TODO(jpeach) The builtin "route-all-default" breaks Gateway.
	// It's supposed to apply to an outbound and send all traffic
	// to the service the outbound matches. However, we don't have
	// an outbound, so there is nowhere for a default route to
	// automatically send traffic.
	//
	// This will be resolved when we add the "selectors" field to
	// TrafficRoute since "route-all-default" will no longer be a
	// match for gateways.

	for _, http := range route.Spec.GetConf().GetHttp() {
		route := envoy.Route{
			Match:    http.Match,
			Modify:   http.Modify,
			Clusters: clustersFromSplit(http.GetSplitWithDestination()),
		}
		routes = append(routes, route)
	}

	if defaultDestination := route.Spec.GetConf().GetSplitWithDestination(); len(defaultDestination) > 0 {
		cfs := clustersFromSplit(defaultDestination)
		routes = append(routes, envoy.Route{
			Match:    nil,
			Clusters: cfs,
		})
	}

	return routes
}
