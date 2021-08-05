package generator

import (
	"context"

	"github.com/pkg/errors"

	kuma_mesh "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

var outboundLog = core.Log.WithName("outbound-proxy-generator")

// OriginOutbound is a marker to indicate by which ProxyGenerator resources were generated.
const OriginOutbound = "outbound"

type OutboundProxyGenerator struct {
}

// Whenever `split` is specified in the TrafficRoute which has more than kuma.io/service tag
// We generate a separate Envoy cluster with _X_ suffix. SplitCounter ensures that we have different X for every split in one Dataplane
// Each split is distinct for the whole Dataplane so we can avoid accidental cluster overrides.
type splitCounter struct {
	counter int
}

func (s *splitCounter) getAndIncrement() int {
	counter := s.counter
	s.counter++
	return counter
}

func (g OutboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	outbounds := proxy.Dataplane.Spec.Networking.GetOutbound()
	resources := model.NewResourceSet()
	if len(outbounds) == 0 {
		return resources, nil
	}
	services := envoy_common.Services{}
	splitCounter := &splitCounter{}

	for _, outbound := range outbounds {
		// Determine the list of destination subsets
		// For one outbound listener it may contain many subsets (ex. TrafficRoute to many destinations)
		routes, err := g.determineRoutes(proxy, outbound, splitCounter)
		if err != nil {
			return nil, err
		}
		clusters := routes.Clusters()
		services.Add(clusters...)

		protocol := g.inferProtocol(proxy, clusters)

		// Generate listener
		listener, err := g.generateLDS(proxy, routes, outbound, protocol)
		if err != nil {
			return nil, err
		}
		resources.Add(&model.Resource{
			Name:     listener.GetName(),
			Origin:   OriginOutbound,
			Resource: listener,
		})
	}

	// Generate clusters. It cannot be generated on the fly with outbound loop because we need to know all subsets of the cluster for every service.
	cdsResources, err := g.generateCDS(ctx, proxy, services)
	if err != nil {
		return nil, err
	}
	resources.AddSet(cdsResources)

	edsResources, err := g.generateEDS(ctx, services, proxy.APIVersion)
	if err != nil {
		return nil, err
	}
	resources.AddSet(edsResources)

	return resources, nil
}

func (_ OutboundProxyGenerator) generateLDS(proxy *model.Proxy, routes envoy_common.Routes, outbound *kuma_mesh.Dataplane_Networking_Outbound, protocol mesh_core.Protocol) (envoy_common.NamedResource, error) {
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
	meshName := proxy.Dataplane.Meta.GetMesh()
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	serviceName := outbound.GetTagsIncludingLegacy()[kuma_mesh.ServiceTag]
	outboundListenerName := envoy_names.GetOutboundListenerName(oface.DataplaneIP, oface.DataplanePort)
	retryPolicy := proxy.Policies.Retries[serviceName]
	var timeoutPolicyConf *kuma_mesh.Timeout_Conf
	if timeoutPolicy := proxy.Policies.Timeouts[oface]; timeoutPolicy != nil {
		timeoutPolicyConf = timeoutPolicy.Spec.GetConf()
	}
	filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
		filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion)
		switch protocol {
		case mesh_core.ProtocolGRPC:
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName, false)).
				Configure(envoy_listeners.Tracing(proxy.Policies.TracingBackend, sourceService)).
				Configure(envoy_listeners.HttpAccessLog(meshName, envoy_common.TrafficDirectionOutbound, sourceService, serviceName, proxy.Policies.Logs[serviceName], proxy)).
				Configure(envoy_listeners.HttpOutboundRoute(serviceName, routes, proxy.Dataplane.Spec.TagSet())).
				Configure(envoy_listeners.Retry(retryPolicy, protocol)).
				Configure(envoy_listeners.GrpcStats())
		case mesh_core.ProtocolHTTP, mesh_core.ProtocolHTTP2:
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName, false)).
				Configure(envoy_listeners.Tracing(proxy.Policies.TracingBackend, sourceService)).
				Configure(envoy_listeners.HttpAccessLog(
					meshName,
					envoy_common.TrafficDirectionOutbound,
					sourceService,
					serviceName,
					proxy.Policies.Logs[serviceName],
					proxy,
				)).
				Configure(envoy_listeners.HttpOutboundRoute(serviceName, routes, proxy.Dataplane.Spec.TagSet())).
				Configure(envoy_listeners.Retry(retryPolicy, protocol))
		case mesh_core.ProtocolKafka:
			filterChainBuilder.
				Configure(envoy_listeners.Kafka(serviceName)).
				Configure(envoy_listeners.TcpProxy(serviceName, routes.Clusters()...)).
				Configure(envoy_listeners.NetworkAccessLog(
					meshName,
					envoy_common.TrafficDirectionOutbound,
					sourceService,
					serviceName,
					proxy.Policies.Logs[serviceName],
					proxy,
				)).
				Configure(envoy_listeners.MaxConnectAttempts(retryPolicy))

		case mesh_core.ProtocolTCP:
			fallthrough
		default:
			// configuration for non-HTTP cases
			filterChainBuilder.
				Configure(envoy_listeners.TcpProxy(serviceName, routes.Clusters()...)).
				Configure(envoy_listeners.NetworkAccessLog(
					meshName,
					envoy_common.TrafficDirectionOutbound,
					sourceService,
					serviceName,
					proxy.Policies.Logs[serviceName],
					proxy,
				)).
				Configure(envoy_listeners.MaxConnectAttempts(retryPolicy))
		}

		filterChainBuilder.
			Configure(envoy_listeners.Timeout(timeoutPolicyConf, protocol))
		return filterChainBuilder
	}()
	listener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
		Configure(envoy_listeners.OutboundListener(outboundListenerName, oface.DataplaneIP, oface.DataplanePort, model.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.FilterChain(filterChainBuilder)).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener %s for service %s", outboundListenerName, serviceName)
	}
	return listener, nil
}

func (o OutboundProxyGenerator) generateCDS(ctx xds_context.Context, proxy *model.Proxy, services envoy_common.Services) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	for _, serviceName := range services.Names() {
		service := services[serviceName]
		healthCheck := proxy.Policies.HealthChecks[serviceName]
		circuitBreaker := proxy.Policies.CircuitBreakers[serviceName]
		protocol := o.inferProtocol(proxy, service.Clusters())

		for _, cluster := range service.Clusters() {
			edsClusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
				Configure(envoy_clusters.Timeout(protocol, cluster.Timeout())).
				Configure(envoy_clusters.CircuitBreaker(circuitBreaker)).
				Configure(envoy_clusters.OutlierDetection(circuitBreaker)).
				Configure(envoy_clusters.HealthCheck(protocol, healthCheck))

			if service.HasExternalService() {
				edsClusterBuilder.
					Configure(envoy_clusters.StrictDNSCluster(cluster.Name(), proxy.Routing.OutboundTargets[serviceName],
						proxy.Dataplane.IsIPv6())).
					Configure(envoy_clusters.ClientSideTLS(proxy.Routing.OutboundTargets[serviceName]))
				switch protocol {
				case mesh_core.ProtocolHTTP:
					edsClusterBuilder.Configure(envoy_clusters.Http())
				case mesh_core.ProtocolHTTP2, mesh_core.ProtocolGRPC:
					edsClusterBuilder.Configure(envoy_clusters.Http2())
				default:
				}
			} else {
				edsClusterBuilder.
					Configure(envoy_clusters.EdsCluster(cluster.Name())).
					Configure(envoy_clusters.LB(cluster.LB())).
					Configure(envoy_clusters.ClientSideMTLS(ctx, proxy.Metadata, serviceName, []envoy_common.Tags{cluster.Tags()})).
					Configure(envoy_clusters.Http2())
			}
			edsCluster, err := edsClusterBuilder.Build()
			if err != nil {
				return nil, errors.Wrapf(err, "build CDS for cluster %s failed", cluster.Name())
			}
			resources.Add(&model.Resource{
				Name:     cluster.Name(),
				Origin:   OriginOutbound,
				Resource: edsCluster,
			})
		}
	}

	return resources, nil
}

func (_ OutboundProxyGenerator) generateEDS(ctx xds_context.Context, services envoy_common.Services, apiVersion envoy_common.APIVersion) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()
	for _, serviceName := range services.Names() {
		// Endpoints for ExternalServices are specified in load assignment in DNS Cluster.
		// We are not allowed to add endpoints with DNS names through EDS.
		if !services[serviceName].HasExternalService() {
			for _, cluster := range services[serviceName].Clusters() {
				loadAssignment, err := ctx.ControlPlane.CLACache.GetCLA(context.Background(), ctx.Mesh.Resource.Meta.GetName(), ctx.Mesh.Hash, cluster, apiVersion)
				if err != nil {
					return nil, errors.Wrapf(err, "could not get ClusterLoadAssignment for %s", serviceName)
				}
				resources.Add(&model.Resource{
					Name:     cluster.Name(),
					Resource: loadAssignment,
				})
			}
		}
	}
	return resources, nil
}

// inferProtocol infers protocol for the destination listener. It will only return HTTP when all endpoints are tagged with HTTP.
func (_ OutboundProxyGenerator) inferProtocol(proxy *model.Proxy, clusters []envoy_common.Cluster) mesh_core.Protocol {
	var allEndpoints []model.Endpoint
	for _, cluster := range clusters {
		serviceName := cluster.Tags()[kuma_mesh.ServiceTag]
		endpoints := model.EndpointList(proxy.Routing.OutboundTargets[serviceName])
		allEndpoints = append(allEndpoints, endpoints...)
	}
	return InferServiceProtocol(allEndpoints)
}

func (_ OutboundProxyGenerator) determineRoutes(proxy *model.Proxy, outbound *kuma_mesh.Dataplane_Networking_Outbound, splitCounter *splitCounter) (routes envoy_common.Routes, err error) {
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)

	route := proxy.Routing.TrafficRoutes[oface]
	if route == nil {
		outboundLog.Info("there is no selected TrafficRoute for the outbound interface, which means that the traffic won't be routed. Visit https://kuma.io/docs/latest/policies/traffic-route/ to check how to introduce the routing.", "dataplane", proxy.Dataplane.Meta.GetName(), "mesh", proxy.Dataplane.Meta.GetMesh(), "outbound", oface)
		return nil, nil
	}

	var timeoutConf *kuma_mesh.Timeout_Conf
	if timeout := proxy.Policies.Timeouts[oface]; timeout != nil {
		timeoutConf = timeout.Spec.GetConf()
	}

	// ClusterCache (cluster hash -> cluster name) protects us from creating excessive amount of caches.
	// For one outbound we pick one traffic route so LB and Timeout are the same.
	// If we have same split in many HTTP matches we can use the same cluster with different weight
	clusterCache := map[string]string{}

	clustersFromSplit := func(splits []*kuma_mesh.TrafficRoute_Split) []envoy_common.Cluster {
		var clusters []envoy_common.Cluster
		for _, destination := range splits {
			service := destination.Destination[kuma_mesh.ServiceTag]
			if destination.GetWeight().GetValue() == 0 {
				// 0 assumes no traffic is passed there. Envoy doesn't support 0 weight, so instead of passing it to Envoy we just skip such cluster.
				continue
			}

			name := service
			if len(destination.GetDestination()) > 1 {
				name = envoy_names.GetSplitClusterName(service, splitCounter.getAndIncrement())
			}

			// We assume that all the targets are either ExternalServices or not
			// therefore we check only the first one
			isExternalService := false
			if endpoints := proxy.Routing.OutboundTargets[service]; len(endpoints) > 0 {
				ep := endpoints[0]
				if ep.IsExternalService() {
					isExternalService = true
				}
			}

			cluster := envoy_common.NewCluster(
				envoy_common.WithService(service),
				envoy_common.WithName(name),
				envoy_common.WithWeight(destination.GetWeight().GetValue()),
				envoy_common.WithTags(destination.Destination),
				envoy_common.WithTimeout(timeoutConf),
				envoy_common.WithLB(route.Spec.GetConf().GetLoadBalancer()),
				envoy_common.WithExternalService(isExternalService),
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

	for _, http := range route.Spec.GetConf().GetHttp() {
		route := envoy_common.Route{
			Match:    http.Match,
			Modify:   http.Modify,
			Clusters: clustersFromSplit(http.GetSplitWithDestination()),
		}
		routes = append(routes, route)
	}

	if defaultDestination := route.Spec.GetConf().GetSplitWithDestination(); len(defaultDestination) > 0 {
		cfs := clustersFromSplit(defaultDestination)
		route := envoy_common.Route{
			Match:    nil,
			Clusters: cfs,
		}
		routes = append(routes, route)
	}
	return
}
