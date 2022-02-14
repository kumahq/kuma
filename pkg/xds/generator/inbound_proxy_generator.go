package generator

import (
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

// OriginInbound is a marker to indicate by which ProxyGenerator resources were generated.
const OriginInbound = "inbound"

type InboundProxyGenerator struct {
}

func (g InboundProxyGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	endpoints, err := proxy.Dataplane.Spec.Networking.GetInboundInterfaces()
	if err != nil {
		return nil, err
	}
	resources := core_xds.NewResourceSet()
	for i, endpoint := range endpoints {
		// we do not create inbounds for serviceless
		if endpoint.IsServiceLess() {
			continue
		}

		iface := proxy.Dataplane.Spec.Networking.Inbound[i]
		protocol := core_mesh.ParseProtocol(iface.GetProtocol())

		// generate CDS resource
		localClusterName := envoy_names.GetLocalClusterName(endpoint.WorkloadPort)
		clusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
			Configure(envoy_clusters.ProvidedEndpointCluster(localClusterName, false, core_xds.Endpoint{Target: endpoint.WorkloadIP, Port: endpoint.WorkloadPort}))
		switch protocol {
		case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
			clusterBuilder.Configure(envoy_clusters.Http2())
		}
		cluster, err := clusterBuilder.Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate cluster %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), localClusterName)
		}
		resources.Add(&core_xds.Resource{
			Name:     localClusterName,
			Resource: cluster,
			Origin:   OriginInbound,
		})

		routes, err := g.buildInboundRoutes(
			envoy_common.NewCluster(envoy_common.WithService(localClusterName)),
			proxy.Policies.RateLimitsInbound[endpoint])
		if err != nil {
			return nil, err
		}

		// generate LDS resource
		service := iface.GetService()
		inboundListenerName := envoy_names.GetInboundListenerName(endpoint.DataplaneIP, endpoint.DataplanePort)
		filterChainBuilder := func(serverSideMTLS bool) *envoy_listeners.FilterChainBuilder {
			filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion)
			switch protocol {
			// configuration for HTTP case
			case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(localClusterName, true)).
					Configure(envoy_listeners.FaultInjection(proxy.Policies.FaultInjections[endpoint]...)).
					Configure(envoy_listeners.RateLimit(proxy.Policies.RateLimitsInbound[endpoint])).
					Configure(envoy_listeners.Tracing(ctx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace), service)).
					Configure(envoy_listeners.HttpInboundRoutes(service, routes))
			case core_mesh.ProtocolGRPC:
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(localClusterName, true)).
					Configure(envoy_listeners.GrpcStats()).
					Configure(envoy_listeners.FaultInjection(proxy.Policies.FaultInjections[endpoint]...)).
					Configure(envoy_listeners.RateLimit(proxy.Policies.RateLimitsInbound[endpoint])).
					Configure(envoy_listeners.Tracing(ctx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace), service)).
					Configure(envoy_listeners.HttpInboundRoutes(service, routes))
			case core_mesh.ProtocolKafka:
				filterChainBuilder.
					Configure(envoy_listeners.Kafka(localClusterName)).
					Configure(envoy_listeners.TcpProxy(localClusterName, envoy_common.NewCluster(envoy_common.WithService(localClusterName))))
			case core_mesh.ProtocolTCP:
				fallthrough
			default:
				// configuration for non-HTTP cases
				filterChainBuilder.Configure(envoy_listeners.TcpProxy(localClusterName, envoy_common.NewCluster(envoy_common.WithService(localClusterName))))
			}
			if serverSideMTLS {
				filterChainBuilder.
					Configure(envoy_listeners.ServerSideMTLS(ctx.Mesh.Resource))
			}
			return filterChainBuilder.
				Configure(envoy_listeners.NetworkRBAC(inboundListenerName, ctx.Mesh.Resource.MTLSEnabled(),
					proxy.Policies.TrafficPermissions[endpoint]))
		}

		listenerBuilder := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
			Configure(envoy_listeners.InboundListener(inboundListenerName, endpoint.DataplaneIP, endpoint.DataplanePort, core_xds.SocketAddressProtocolTCP)).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying()))

		switch ctx.Mesh.Resource.GetEnabledCertificateAuthorityBackend().GetMode() {
		case mesh_proto.CertificateAuthorityBackend_STRICT:
			listenerBuilder.
				Configure(envoy_listeners.FilterChain(filterChainBuilder(true)))
		case mesh_proto.CertificateAuthorityBackend_PERMISSIVE:
			listenerBuilder.
				Configure(envoy_listeners.TLSInspector()).
				Configure(envoy_listeners.FilterChain(
					filterChainBuilder(false).Configure(
						envoy_listeners.MatchTransportProtocol("raw_buffer"))),
				).
				Configure(envoy_listeners.FilterChain(
					filterChainBuilder(false).Configure(
						envoy_listeners.MatchTransportProtocol("tls"))),
				).
				Configure(envoy_listeners.FilterChain(
					filterChainBuilder(true).Configure(
						envoy_listeners.MatchTransportProtocol("tls"),
						envoy_listeners.MatchApplicationProtocols(xds_tls.KumaALPNProtocols...))),
				)
		default:
			return nil, errors.New("unknown mode for CA backend")
		}

		inboundListener, err := listenerBuilder.Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), inboundListenerName)
		}
		resources.Add(&core_xds.Resource{
			Name:     inboundListenerName,
			Resource: inboundListener,
			Origin:   OriginInbound,
		})
	}
	return resources, nil
}

func (g *InboundProxyGenerator) buildInboundRoutes(cluster envoy_common.Cluster, rateLimits []*core_mesh.RateLimitResource) (envoy_common.Routes, error) {
	routes := envoy_common.Routes{}

	// Iterate over that RateLimits and generate the relevant Routes.
	// We do assume that the rateLimits resource is sorted, so the most
	// specific source matches come first.
	for _, rateLimit := range rateLimits {
		if rateLimit.Spec.GetConf().GetHttp() != nil {
			route := envoy_common.NewRouteFromCluster(cluster)
			if len(rateLimit.Spec.GetSources()) > 0 {
				if route.Match == nil {
					route.Match = &mesh_proto.TrafficRoute_Http_Match{}
				}

				if route.Match.Headers == nil {
					route.Match.Headers = make(map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher)
				}

				var selectorRegexs []string
				for _, selector := range rateLimit.Spec.SourceTags() {
					selectorRegexs = append(selectorRegexs, tags.MatchingRegex(selector))
				}
				regexOR := tags.RegexOR(selectorRegexs...)

				route.Match.Headers[v3.TagsHeaderName] = &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
					MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex{
						Regex: regexOR,
					},
				}
			}

			route.RateLimit = rateLimit.Spec

			routes = append(routes, route)
		}
	}

	// Add the defaul fall-back route
	routes = append(routes, envoy_common.NewRouteFromCluster(cluster))

	return routes, nil
}
