package generator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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

	servicesAcc := envoy_common.NewServicesAccumulator(ctx.Mesh.ServiceTLSReadiness)

	// ClusterCache (cluster hash -> cluster name) protects us from creating excessive amount of caches.
	// For one outbound we pick one traffic route so LB and Timeout are the same.
	// If we have same split in many HTTP matches we can use the same cluster with different weight
	clusterCache := map[string]string{}
	splitCounter := &splitCounter{}

	for _, outbound := range outbounds {
		// Determine the list of destination subsets
		// For one outbound listener it may contain many subsets (ex. TrafficRoute to many destinations)
		routes := g.determineRoutes(proxy, outbound, clusterCache, splitCounter, ctx.Mesh.Resource.ZoneEgressEnabled())
		clusters := routes.Clusters()
		servicesAcc.Add(clusters...)

		protocol := g.inferProtocol(proxy, clusters)

		// Generate listener
		listener, err := g.generateLDS(ctx, proxy, routes, outbound, protocol)
		if err != nil {
			return nil, err
		}
		resources.Add(&model.Resource{
			Name:     listener.GetName(),
			Origin:   OriginOutbound,
			Resource: listener,
		})
	}

	services := servicesAcc.Services()

	// Generate clusters. It cannot be generated on the fly with outbound loop because we need to know all subsets of the cluster for every service.
	cdsResources, err := g.generateCDS(ctx, services, proxy)
	if err != nil {
		return nil, err
	}
	resources.AddSet(cdsResources)

	edsResources, err := g.generateEDS(ctx, services, proxy)
	if err != nil {
		return nil, err
	}
	resources.AddSet(edsResources)

	return resources, nil
}

func (OutboundProxyGenerator) generateLDS(ctx xds_context.Context, proxy *model.Proxy, routes envoy_common.Routes, outbound *mesh_proto.Dataplane_Networking_Outbound, protocol core_mesh.Protocol) (envoy_common.NamedResource, error) {
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
	rateLimits := []*core_mesh.RateLimitResource{}
	if rateLimit, exists := proxy.Policies.RateLimitsOutbound[oface]; exists {
		rateLimits = append(rateLimits, rateLimit)
	}
	meshName := proxy.Dataplane.Meta.GetMesh()
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
	outboundListenerName := envoy_names.GetOutboundListenerName(oface.DataplaneIP, oface.DataplanePort)
	retryPolicy := proxy.Policies.Retries[serviceName]
	var timeoutPolicyConf *mesh_proto.Timeout_Conf
	if timeoutPolicy := proxy.Policies.Timeouts[oface]; timeoutPolicy != nil {
		timeoutPolicyConf = timeoutPolicy.Spec.GetConf()
	}
	filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
		filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion)
		switch protocol {
		case core_mesh.ProtocolGRPC:
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName, false)).
				Configure(envoy_listeners.Tracing(ctx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace), sourceService)).
				Configure(envoy_listeners.HttpAccessLog(meshName, envoy_common.TrafficDirectionOutbound, sourceService, serviceName,
					ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[serviceName]), proxy)).
				Configure(envoy_listeners.HttpOutboundRoute(serviceName, routes, proxy.Dataplane.Spec.TagSet())).
				// backwards compatibility to support RateLimit for ExternalServices without ZoneEgress
				ConfigureIf(!ctx.Mesh.Resource.ZoneEgressEnabled(), envoy_listeners.RateLimit(rateLimits)).
				Configure(envoy_listeners.Retry(retryPolicy, protocol)).
				Configure(envoy_listeners.GrpcStats())
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName, false)).
				Configure(envoy_listeners.Tracing(ctx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace), sourceService)).
				// backwards compatibility to support RateLimit for ExternalServices without ZoneEgress
				ConfigureIf(!ctx.Mesh.Resource.ZoneEgressEnabled(), envoy_listeners.RateLimit(rateLimits)).
				Configure(envoy_listeners.HttpAccessLog(
					meshName,
					envoy_common.TrafficDirectionOutbound,
					sourceService,
					serviceName,
					ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[serviceName]),
					proxy,
				)).
				Configure(envoy_listeners.HttpOutboundRoute(serviceName, routes, proxy.Dataplane.Spec.TagSet())).
				Configure(envoy_listeners.Retry(retryPolicy, protocol))
		case core_mesh.ProtocolKafka:
			filterChainBuilder.
				Configure(envoy_listeners.Kafka(serviceName)).
				Configure(envoy_listeners.TcpProxy(serviceName, routes.Clusters()...)).
				Configure(envoy_listeners.NetworkAccessLog(
					meshName,
					envoy_common.TrafficDirectionOutbound,
					sourceService,
					serviceName,
					ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[serviceName]),
					proxy,
				)).
				Configure(envoy_listeners.MaxConnectAttempts(retryPolicy))

		case core_mesh.ProtocolTCP:
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
					ctx.Mesh.GetLoggingBackend(proxy.Policies.TrafficLogs[serviceName]),
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
		Configure(envoy_listeners.TagsMetadata(envoy_common.Tags(outbound.GetTagsIncludingLegacy()).WithoutTags(mesh_proto.MeshTag))).
		Build()
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate listener %s for service %s", outboundListenerName, serviceName)
	}
	return listener, nil
}

func (g OutboundProxyGenerator) generateCDS(ctx xds_context.Context, services envoy_common.Services, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		service := services[serviceName]
		healthCheck := proxy.Policies.HealthChecks[serviceName]
		circuitBreaker := proxy.Policies.CircuitBreakers[serviceName]
		protocol := g.inferProtocol(proxy, service.Clusters())
		tlsReady := service.TLSReady()

		for _, cluster := range service.Clusters() {
			edsClusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
				Configure(envoy_clusters.Timeout(cluster.Timeout(), protocol)).
				Configure(envoy_clusters.CircuitBreaker(circuitBreaker)).
				Configure(envoy_clusters.OutlierDetection(circuitBreaker)).
				Configure(envoy_clusters.HealthCheck(protocol, healthCheck))

			clusterName := cluster.Name()
			clusterTags := []envoy_common.Tags{cluster.Tags()}

			if service.HasExternalService() {
				if ctx.Mesh.Resource.ZoneEgressEnabled() {
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster(clusterName)).
						Configure(envoy_clusters.ClientSideMTLS(
							proxy.SecretsTracker,
							ctx.Mesh.Resource,
							mesh_proto.ZoneEgressServiceName,
							tlsReady,
							clusterTags,
						))
				} else {
					endpoints := proxy.Routing.OutboundTargets[serviceName]
					isIPv6 := proxy.Dataplane.IsIPv6()

					edsClusterBuilder.
						Configure(envoy_clusters.ProvidedEndpointCluster(clusterName, isIPv6, endpoints...)).
						Configure(envoy_clusters.ClientSideTLS(endpoints))
				}

				switch protocol {
				case core_mesh.ProtocolHTTP:
					edsClusterBuilder.Configure(envoy_clusters.Http())
				case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
					edsClusterBuilder.Configure(envoy_clusters.Http2())
				default:
				}
			} else {
				edsClusterBuilder.
					Configure(envoy_clusters.EdsCluster(clusterName)).
					Configure(envoy_clusters.LB(cluster.LB())).
					Configure(envoy_clusters.Http2())

				if upstreamMeshName := cluster.Mesh(); upstreamMeshName != "" {
					for _, otherMesh := range append(ctx.Mesh.Resources.OtherMeshes().Items, ctx.Mesh.Resource) {
						if otherMesh.GetMeta().GetName() == upstreamMeshName {
							edsClusterBuilder.Configure(
								envoy_clusters.CrossMeshClientSideMTLS(
									proxy.SecretsTracker, ctx.Mesh.Resource, otherMesh, serviceName, tlsReady, clusterTags,
								),
							)
							break
						}
					}
				} else {
					edsClusterBuilder.Configure(envoy_clusters.ClientSideMTLS(
						proxy.SecretsTracker,
						ctx.Mesh.Resource, serviceName, tlsReady, clusterTags))
				}
			}

			edsCluster, err := edsClusterBuilder.Build()
			if err != nil {
				return nil, errors.Wrapf(err, "build CDS for cluster %s failed", clusterName)
			}

			resources.Add(&model.Resource{
				Name:     clusterName,
				Origin:   OriginOutbound,
				Resource: edsCluster,
			})
		}
	}

	return resources, nil
}

func (OutboundProxyGenerator) generateEDS(
	ctx xds_context.Context,
	services envoy_common.Services,
	proxy *model.Proxy,
) (*model.ResourceSet, error) {
	apiVersion := proxy.APIVersion
	resources := model.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		// When no zone egress is present in a mesh Endpoints for ExternalServices
		// are specified in load assignment in DNS Cluster.
		// We are not allowed to add endpoints with DNS names through EDS.
		if !services[serviceName].HasExternalService() || ctx.Mesh.Resource.ZoneEgressEnabled() {
			for _, cluster := range services[serviceName].Clusters() {
				var endpoints model.EndpointMap
				if cluster.Mesh() != "" {
					endpoints = ctx.Mesh.CrossMeshEndpoints[cluster.Mesh()]
				} else {
					endpoints = ctx.Mesh.EndpointMap
				}

				loadAssignment, err := ctx.ControlPlane.CLACache.GetCLA(context.Background(), ctx.Mesh.Resource.Meta.GetName(), ctx.Mesh.Hash, cluster, apiVersion, endpoints)
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
func (OutboundProxyGenerator) inferProtocol(proxy *model.Proxy, clusters []envoy_common.Cluster) core_mesh.Protocol {
	var allEndpoints []model.Endpoint
	for _, cluster := range clusters {
		serviceName := cluster.Tags()[mesh_proto.ServiceTag]
		endpoints := model.EndpointList(proxy.Routing.OutboundTargets[serviceName])
		allEndpoints = append(allEndpoints, endpoints...)
	}
	return InferServiceProtocol(allEndpoints)
}

func (OutboundProxyGenerator) determineRoutes(
	proxy *model.Proxy,
	outbound *mesh_proto.Dataplane_Networking_Outbound,
	clusterCache map[string]string,
	splitCounter *splitCounter,
	hasEgress bool,
) envoy_common.Routes {
	var routes envoy_common.Routes
	oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)

	route := proxy.Routing.TrafficRoutes[oface]
	if route == nil {
		outboundLog.Info("there is no selected TrafficRoute for the outbound interface, which means that the traffic won't be routed. Visit https://kuma.io/docs/latest/policies/traffic-route/ to check how to introduce the routing.", "dataplane", proxy.Dataplane.Meta.GetName(), "mesh", proxy.Dataplane.Meta.GetMesh(), "outbound", oface)
		return nil
	}

	var timeoutConf *mesh_proto.Timeout_Conf
	if timeout := proxy.Policies.Timeouts[oface]; timeout != nil {
		timeoutConf = timeout.Spec.GetConf()
	}

	// Return internal, external
	clustersFromSplit := func(splits []*mesh_proto.TrafficRoute_Split) ([]envoy_common.Cluster, []envoy_common.Cluster) {
		var clustersInternal []envoy_common.Cluster
		var clustersExternal []envoy_common.Cluster
		for _, destination := range splits {
			service := destination.Destination[mesh_proto.ServiceTag]
			if destination.GetWeight().GetValue() == 0 {
				// 0 assumes no traffic is passed there. Envoy doesn't support 0 weight, so instead of passing it to Envoy we just skip such cluster.
				continue
			}

			name := service

			if len(destination.GetDestination()) > 1 {
				name = envoy_names.GetSplitClusterName(service, splitCounter.getAndIncrement())
			}

			if mesh, ok := destination.Destination[mesh_proto.MeshTag]; ok {
				// The name should be distinct to the service & mesh combination
				name = fmt.Sprintf("%s_%s", name, mesh)
			}

			// We assume that all the targets are either ExternalServices or not
			// therefore we check only the first one
			var isExternalService bool
			if endpoints := proxy.Routing.OutboundTargets[service]; len(endpoints) > 0 {
				isExternalService = endpoints[0].IsExternalService()
			}

			cluster := envoy_common.NewCluster(
				envoy_common.WithService(service),
				envoy_common.WithName(name),
				envoy_common.WithWeight(destination.GetWeight().GetValue()),
				// The mesh tag is set here if this destination is generated
				// from a MeshGateway virtual outbound and is not part of the
				// service tags
				envoy_common.WithTags(envoy_common.Tags(destination.Destination).WithoutTags(mesh_proto.MeshTag)),
				envoy_common.WithTimeout(timeoutConf),
				envoy_common.WithLB(route.Spec.GetConf().GetLoadBalancer()),
				envoy_common.WithExternalService(isExternalService),
			)

			if mesh, ok := destination.Destination[mesh_proto.MeshTag]; ok {
				cluster.SetMesh(mesh)
			}

			if name, ok := clusterCache[cluster.Tags().String()]; ok {
				cluster.SetName(name)
			} else {
				clusterCache[cluster.Tags().String()] = cluster.Name()
			}

			if isExternalService {
				clustersExternal = append(clustersExternal, cluster)
			} else {
				clustersInternal = append(clustersInternal, cluster)
			}
		}
		return clustersInternal, clustersExternal
	}

	appendRoute := func(routes envoy_common.Routes, match *mesh_proto.TrafficRoute_Http_Match, modify *mesh_proto.TrafficRoute_Http_Modify,
		clusters []envoy_common.Cluster, rateLimit *core_mesh.RateLimitResource) envoy_common.Routes {
		if len(clusters) == 0 {
			return routes
		}

		// backwards compatibility to support RateLimit for ExternalServices without ZoneEgress
		if hasEgress {
			return append(routes, envoy_common.Route{
				Match:    match,
				Modify:   modify,
				Clusters: clusters,
			})
		} else {
			var rlSpec *mesh_proto.RateLimit
			if rateLimit != nil {
				rlSpec = rateLimit.Spec
			}
			return append(routes, envoy_common.Route{
				Match:     match,
				Modify:    modify,
				RateLimit: rlSpec,
				Clusters:  clusters,
			})
		}
	}

	for _, http := range route.Spec.GetConf().GetHttp() {
		clustersInternal, clustersExternal := clustersFromSplit(http.GetSplitWithDestination())
		routes = appendRoute(routes, http.Match, http.Modify, clustersInternal, nil)
		routes = appendRoute(routes, http.Match, http.Modify, clustersExternal, proxy.Policies.RateLimitsOutbound[oface])
	}

	if defaultDestination := route.Spec.GetConf().GetSplitWithDestination(); len(defaultDestination) != 0 {
		clustersInternal, clustersExternal := clustersFromSplit(defaultDestination)
		routes = appendRoute(routes, nil, nil, clustersInternal, nil)
		routes = appendRoute(routes, nil, nil, clustersExternal, proxy.Policies.RateLimitsOutbound[oface])
	}

	return routes
}
