package generator

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/user"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	util_protocol "github.com/kumahq/kuma/pkg/util/protocol"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

var outboundLog = core.Log.WithName("xds").WithName("outbound-proxy-generator")

// OriginOutbound is a marker to indicate by which ProxyGenerator resources were generated.
const OriginOutbound = "outbound"

type OutboundProxyGenerator struct{}

func (g OutboundProxyGenerator) Generate(ctx context.Context, _ *model.ResourceSet, xdsCtx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	if len(xdsCtx.Mesh.Resources.TrafficRoutes().Items) == 0 {
		return nil, nil
	}

	outbounds := proxy.Dataplane.Spec.Networking.GetOutbounds(mesh_proto.NonBackendRefFilter) // backend refs work only with new policies.
	resources := model.NewResourceSet()
	if len(outbounds) == 0 {
		return resources, nil
	}

	tlsReady := xdsCtx.Mesh.GetTLSReadiness()
	servicesAcc := envoy_common.NewServicesAccumulator(tlsReady)

	// ClusterCache (cluster hash -> cluster name) protects us from creating excessive amount of caches.
	// For one outbound we pick one traffic route so LB and Timeout are the same.
	// If we have same split in many HTTP matches we can use the same cluster with different weight
	clusterCache := map[string]string{}

	outboundsMultipleIPs := buildOutboundsWithMultipleIPs(proxy.Dataplane, outbounds, xdsCtx.Mesh.VIPDomains)
	for _, outbound := range outboundsMultipleIPs {
		// Determine the list of destination subsets
		// For one outbound listener it may contain many subsets (ex. TrafficRoute to many destinations)
		routes := g.determineRoutes(proxy, outbound.Addresses[0], clusterCache, xdsCtx.Mesh.Resource.ZoneEgressEnabled())
		clusters := routes.Clusters()

		protocol := inferProtocol(xdsCtx.Mesh, clusters)

		servicesAcc.Add(clusters...)

		// Generate listener
		listener, err := g.generateLDS(xdsCtx, proxy, routes, outbound, protocol)
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
	cdsResources, err := g.generateCDS(xdsCtx, services, proxy)
	if err != nil {
		return nil, err
	}
	resources.AddSet(cdsResources)

	edsResources, err := g.generateEDS(ctx, xdsCtx, services, proxy)
	if err != nil {
		return nil, err
	}
	resources.AddSet(edsResources)
	return resources, nil
}

func (OutboundProxyGenerator) generateLDS(ctx xds_context.Context, proxy *model.Proxy, routes envoy_common.Routes, outbound OutboundWithMultipleIPs, protocol core_mesh.Protocol) (envoy_common.NamedResource, error) {
	oface := outbound.Addresses[0]
	rateLimits := []*core_mesh.RateLimitResource{}
	if rateLimit, exists := proxy.Policies.RateLimitsOutbound[oface]; exists {
		rateLimits = append(rateLimits, rateLimit)
	}
	meshName := proxy.Dataplane.Meta.GetMesh()
	sourceService := proxy.Dataplane.Spec.GetIdentifyingService()
	serviceName := outbound.Tags[mesh_proto.ServiceTag]
	outboundListenerName := envoy_names.GetOutboundListenerName(oface.DataplaneIP, oface.DataplanePort)
	retryPolicy := proxy.Policies.Retries[serviceName]
	var timeoutPolicyConf *mesh_proto.Timeout_Conf
	if timeoutPolicy := proxy.Policies.Timeouts[oface]; timeoutPolicy != nil {
		timeoutPolicyConf = timeoutPolicy.Spec.GetConf()
	}
	filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
		filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource)
		switch protocol {
		case core_mesh.ProtocolGRPC:
			filterChainBuilder.
				Configure(envoy_listeners.HttpConnectionManager(serviceName, false)).
				Configure(envoy_listeners.Tracing(
					ctx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace),
					sourceService,
					envoy_common.TrafficDirectionOutbound,
					serviceName,
					false,
				)).
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
				Configure(envoy_listeners.Tracing(
					ctx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace),
					sourceService,
					envoy_common.TrafficDirectionOutbound,
					serviceName,
					false,
				)).
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
				Configure(envoy_listeners.TcpProxyDeprecated(serviceName, routes.Clusters()...)).
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
				Configure(envoy_listeners.TcpProxyDeprecated(serviceName, routes.Clusters()...)).
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
	listener, err := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, oface.DataplaneIP, oface.DataplanePort, model.SocketAddressProtocolTCP).
		Configure(envoy_listeners.FilterChain(filterChainBuilder)).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(outbound.Tags).WithoutTags(mesh_proto.MeshTag))).
		Configure(envoy_listeners.AdditionalAddresses(outbound.AdditionalAddresses())).
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
		protocol := ctx.Mesh.GetServiceProtocol(serviceName)
		tlsReady := service.TLSReady()

		for _, c := range service.Clusters() {
			cluster := c.(*envoy_common.ClusterImpl)
			clusterName := cluster.Name()
			edsClusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, clusterName).
				Configure(envoy_clusters.Timeout(cluster.Timeout(), protocol)).
				Configure(envoy_clusters.CircuitBreaker(circuitBreaker)).
				Configure(envoy_clusters.OutlierDetection(circuitBreaker)).
				Configure(envoy_clusters.HealthCheck(protocol, healthCheck))

			clusterTags := []envoy_tags.Tags{cluster.Tags()}

			if service.HasExternalService() {
				if ctx.Mesh.Resource.ZoneEgressEnabled() {
					edsClusterBuilder.
						Configure(envoy_clusters.EdsCluster()).
						Configure(envoy_clusters.ClientSideMTLS(
							proxy.SecretsTracker,
							ctx.Mesh.Resource,
							mesh_proto.ZoneEgressServiceName,
							tlsReady,
							clusterTags,
						))
				} else {
					endpoints := proxy.Routing.ExternalServiceOutboundTargets[serviceName]
					isIPv6 := proxy.Dataplane.IsIPv6()

					edsClusterBuilder.
						Configure(envoy_clusters.ProvidedEndpointCluster(isIPv6, endpoints...)).
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
					Configure(envoy_clusters.EdsCluster()).
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
	ctx context.Context,
	xdsCtx xds_context.Context,
	services envoy_common.Services,
	proxy *model.Proxy,
) (*model.ResourceSet, error) {
	apiVersion := proxy.APIVersion
	resources := model.NewResourceSet()

	for _, serviceName := range services.Sorted() {
		// When no zone egress is present in a mesh Endpoints for ExternalServices
		// are specified in load assignment in DNS Cluster.
		// We are not allowed to add endpoints with DNS names through EDS.
		if !services[serviceName].HasExternalService() || xdsCtx.Mesh.Resource.ZoneEgressEnabled() {
			for _, c := range services[serviceName].Clusters() {
				cluster := c.(*envoy_common.ClusterImpl)
				var endpoints model.EndpointMap
				if cluster.Mesh() != "" {
					endpoints = xdsCtx.Mesh.CrossMeshEndpoints[cluster.Mesh()]
				} else {
					endpoints = xdsCtx.Mesh.EndpointMap
				}

				loadAssignment, err := xdsCtx.ControlPlane.CLACache.GetCLA(user.Ctx(ctx, user.ControlPlane), xdsCtx.Mesh.Resource.Meta.GetName(), xdsCtx.Mesh.Hash, cluster, apiVersion, endpoints)
				if err != nil {
					return nil, errors.Wrapf(err, "could not get ClusterLoadAssignment for %s", serviceName)
				}

				resources.Add(&model.Resource{
					Name:     cluster.Name(),
					Origin:   OriginOutbound,
					Resource: loadAssignment,
				})
			}
		}
	}

	return resources, nil
}

func inferProtocol(meshCtx xds_context.MeshContext, clusters []envoy_common.Cluster) core_mesh.Protocol {
	var protocol core_mesh.Protocol = core_mesh.ProtocolUnknown
	for idx, cluster := range clusters {
		serviceName := cluster.Tags()[mesh_proto.ServiceTag]
		serviceProtocol := meshCtx.GetServiceProtocol(serviceName)
		if idx == 0 {
			protocol = serviceProtocol
			continue
		}
		protocol = util_protocol.GetCommonProtocol(serviceProtocol, protocol)
	}
	return protocol
}

func (OutboundProxyGenerator) determineRoutes(
	proxy *model.Proxy,
	oface mesh_proto.OutboundInterface,
	clusterCache map[string]string,
	hasEgress bool,
) envoy_common.Routes {
	var routes envoy_common.Routes

	route := proxy.Routing.TrafficRoutes[oface]
	if route == nil {
		if len(proxy.Policies.TrafficRoutes) > 0 {
			outboundLog.Info("there is no selected TrafficRoute for the outbound interface, which means that the traffic won't be routed. Visit https://kuma.io/docs/latest/policies/traffic-route/ to check how to introduce the routing.", "dataplane", proxy.Dataplane.Meta.GetName(), "mesh", proxy.Dataplane.Meta.GetMesh(), "outbound", oface)
		}
		return nil
	}

	var timeoutConf *mesh_proto.Timeout_Conf
	if timeout := proxy.Policies.Timeouts[oface]; timeout != nil {
		timeoutConf = timeout.Spec.GetConf()
	}

	rateLimit := proxy.Policies.RateLimitsOutbound[oface]

	clustersFromSplit := func(splits []*mesh_proto.TrafficRoute_Split) []envoy_common.Cluster {
		var clusters []envoy_common.Cluster
		for _, destination := range splits {
			service := destination.Destination[mesh_proto.ServiceTag]
			if destination.GetWeight().GetValue() == 0 {
				// 0 assumes no traffic is passed there. Envoy doesn't support 0 weight, so instead of passing it to Envoy we just skip such cluster.
				continue
			}

			name, _ := tags.Tags(destination.Destination).DestinationClusterName(nil)

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
			if endpoints := proxy.Routing.ExternalServiceOutboundTargets[service]; len(endpoints) > 0 {
				isExternalService = true
			}

			allTags := envoy_tags.Tags(destination.Destination)
			cluster := envoy_common.NewCluster(
				envoy_common.WithService(service),
				envoy_common.WithName(name),
				envoy_common.WithWeight(destination.GetWeight().GetValue()),
				// The mesh tag is set here if this destination is generated
				// from a MeshGateway virtual outbound and is not part of the
				// service tags
				envoy_common.WithTags(allTags.WithoutTags(mesh_proto.MeshTag)),
				envoy_common.WithTimeout(timeoutConf),
				envoy_common.WithLB(route.Spec.GetConf().GetLoadBalancer()),
				envoy_common.WithExternalService(isExternalService),
			)

			if mesh, ok := destination.Destination[mesh_proto.MeshTag]; ok {
				cluster.SetMesh(mesh)
			}

			if name, ok := clusterCache[allTags.String()]; ok {
				cluster.SetName(name)
			} else {
				clusterCache[allTags.String()] = cluster.Name()
			}

			clusters = append(clusters, cluster)
		}
		return clusters
	}

	appendRoute := func(routes envoy_common.Routes, match *mesh_proto.TrafficRoute_Http_Match, modify *mesh_proto.TrafficRoute_Http_Modify,
		clusters []envoy_common.Cluster,
	) envoy_common.Routes {
		if len(clusters) == 0 {
			return routes
		}

		hasExternal := false
		for _, cluster := range clusters {
			if cluster.IsExternalService() {
				hasExternal = true
				break
			}
		}

		var rlSpec *mesh_proto.RateLimit
		if hasExternal && !hasEgress && rateLimit != nil {
			rlSpec = rateLimit.Spec
		} // otherwise rate limit is applied on the inbound side

		return append(routes, envoy_common.Route{
			Match:     match,
			Modify:    modify,
			RateLimit: rlSpec,
			Clusters:  clusters,
		})
	}

	for _, http := range route.Spec.GetConf().GetHttp() {
		routes = appendRoute(routes, http.Match, http.Modify, clustersFromSplit(http.GetSplitWithDestination()))
	}

	if defaultDestination := route.Spec.GetConf().GetSplitWithDestination(); len(defaultDestination) != 0 {
		routes = appendRoute(routes, nil, nil, clustersFromSplit(defaultDestination))
	}

	return routes
}

type OutboundWithMultipleIPs struct {
	Tags      map[string]string
	Addresses []mesh_proto.OutboundInterface
}

func (o OutboundWithMultipleIPs) AdditionalAddresses() []mesh_proto.OutboundInterface {
	if len(o.Addresses) > 1 {
		return o.Addresses[1:]
	}
	return nil
}

func buildOutboundsWithMultipleIPs(dataplane *core_mesh.DataplaneResource, outbounds []*mesh_proto.Dataplane_Networking_Outbound, meshVIPDomains []xds_types.VIPDomains) []OutboundWithMultipleIPs {
	kumaVIPs := map[string]bool{}
	for _, vipDomain := range meshVIPDomains {
		kumaVIPs[vipDomain.Address] = true
	}

	tagsToOutbounds := map[string]OutboundWithMultipleIPs{}
	for _, outbound := range outbounds {
		tags := util_maps.Clone(outbound.GetTags())
		tags[mesh_proto.ServiceTag] = outbound.GetService()
		tagsStr := mesh_proto.SingleValueTagSet(tags).String()
		owmi := tagsToOutbounds[tagsStr]
		owmi.Tags = tags
		address := dataplane.Spec.Networking.ToOutboundInterface(outbound)
		// add Kuma VIPs down the list, so if there is a non Kuma VIP (i.e. Kube Cluster IP), it goes as primary address.
		if kumaVIPs[address.DataplaneIP] {
			owmi.Addresses = append(owmi.Addresses, address)
		} else {
			owmi.Addresses = append([]mesh_proto.OutboundInterface{address}, owmi.Addresses...)
		}
		tagsToOutbounds[tagsStr] = owmi
	}

	// return sorted outbounds for a stable XDS config
	var result []OutboundWithMultipleIPs
	for _, key := range util_maps.SortedKeys(tagsToOutbounds) {
		result = append(result, tagsToOutbounds[key])
	}
	return result
}
