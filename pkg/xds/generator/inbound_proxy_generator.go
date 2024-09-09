package generator

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	defaults_mesh "github.com/kumahq/kuma/pkg/defaults/mesh"
	"github.com/kumahq/kuma/pkg/util/net"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

// OriginInbound is a marker to indicate by which ProxyGenerator resources were generated.
const OriginInbound = "inbound"

type InboundProxyGenerator struct{}

func (g InboundProxyGenerator) Generate(ctx context.Context, _ *core_xds.ResourceSet, xdsCtx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	for i, endpoint := range proxy.Dataplane.Spec.Networking.GetInboundInterfaces() {
		// we do not create inbounds for serviceless
		if endpoint.IsServiceLess() {
			continue
		}

		iface := proxy.Dataplane.Spec.Networking.Inbound[i]
		protocol := core_mesh.ParseProtocol(iface.GetProtocol())

		// generate CDS resource
		localClusterName := envoy_names.GetLocalClusterName(endpoint.WorkloadPort)
		clusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, localClusterName).
			Configure(envoy_clusters.ProvidedEndpointCluster(false, core_xds.Endpoint{Target: endpoint.WorkloadIP, Port: endpoint.WorkloadPort})).
			Configure(envoy_clusters.Timeout(defaults_mesh.DefaultInboundTimeout(), protocol))
		// localhost traffic is routed dirrectly to the application, in case of other interface we are going to set source address to
		// 127.0.0.6 to avoid redirections and thanks to first iptables rule just return fast
		if proxy.Dataplane.IsUsingTransparentProxy() && (endpoint.WorkloadIP != core_mesh.IPv4Loopback.String() && endpoint.WorkloadIP != core_mesh.IPv6Loopback.String()) {
			switch net.IsAddressIPv6(endpoint.WorkloadIP) {
			case true:
				clusterBuilder.Configure(envoy_clusters.UpstreamBindConfig(InPassThroughIPv6, 0))
			case false:
				clusterBuilder.Configure(envoy_clusters.UpstreamBindConfig(InPassThroughIPv4, 0))
			}
		}

		switch protocol {
		case core_mesh.ProtocolHTTP:
			clusterBuilder.Configure(envoy_clusters.Http())
		case core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
			clusterBuilder.Configure(envoy_clusters.Http2())
		}
		envoyCluster, err := clusterBuilder.Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate cluster %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), localClusterName)
		}
		resources.Add(&core_xds.Resource{
			Name:     localClusterName,
			Resource: envoyCluster,
			Origin:   OriginInbound,
		})

		cluster := envoy_common.NewCluster(envoy_common.WithService(localClusterName))
		routes := GenerateRoutes(proxy, endpoint, cluster)

		// generate LDS resource
		service := iface.GetService()
		inboundListenerName := envoy_names.GetInboundListenerName(endpoint.DataplaneIP, endpoint.DataplanePort)

		listenerBuilder := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, endpoint.DataplaneIP, endpoint.DataplanePort, core_xds.SocketAddressProtocolTCP).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Configure(envoy_listeners.TagsMetadata(iface.GetTags()))

		switch xdsCtx.Mesh.Resource.GetEnabledCertificateAuthorityBackend().GetMode() {
		case mesh_proto.CertificateAuthorityBackend_STRICT:
			listenerBuilder.
				Configure(envoy_listeners.FilterChain(FilterChainBuilder(true, protocol, proxy, localClusterName, xdsCtx, endpoint, service, &routes, nil, nil).Configure(
					envoy_listeners.NetworkRBAC(inboundListenerName, xdsCtx.Mesh.Resource.MTLSEnabled(), proxy.Policies.TrafficPermissions[endpoint]),
				)))
		case mesh_proto.CertificateAuthorityBackend_PERMISSIVE:
			listenerBuilder.
				Configure(envoy_listeners.TLSInspector()).
				Configure(envoy_listeners.FilterChain(
					FilterChainBuilder(false, protocol, proxy, localClusterName, xdsCtx, endpoint, service, &routes, nil, nil).Configure(
						envoy_listeners.MatchTransportProtocol("raw_buffer"))),
				).
				Configure(envoy_listeners.FilterChain(
					// we need to differentiate between just TLS and Kuma's TLS, because with permissive mode
					// the app itself might be protected by TLS.
					FilterChainBuilder(false, protocol, proxy, localClusterName, xdsCtx, endpoint, service, &routes, nil, nil).Configure(
						envoy_listeners.MatchTransportProtocol("tls"))),
				).
				Configure(envoy_listeners.FilterChain(
					FilterChainBuilder(true, protocol, proxy, localClusterName, xdsCtx, endpoint, service, &routes, nil, nil).Configure(
						envoy_listeners.MatchTransportProtocol("tls"),
						envoy_listeners.MatchApplicationProtocols(xds_tls.KumaALPNProtocols...),
						envoy_listeners.NetworkRBAC(inboundListenerName, xdsCtx.Mesh.Resource.MTLSEnabled(), proxy.Policies.TrafficPermissions[endpoint]),
					)),
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

func FilterChainBuilder(serverSideMTLS bool, protocol core_mesh.Protocol, proxy *core_xds.Proxy, localClusterName string, xdsCtx xds_context.Context, endpoint mesh_proto.InboundInterface, service string, routes *envoy_common.Routes, tlsVersion *tls.Version, ciphers tls.TlsCiphers) *envoy_listeners.FilterChainBuilder {
	filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource)
	switch protocol {
	// configuration for HTTP case
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
		filterChainBuilder.
			Configure(envoy_listeners.HttpConnectionManager(localClusterName, true)).
			Configure(envoy_listeners.FaultInjection(proxy.Policies.FaultInjections[endpoint]...)).
			Configure(envoy_listeners.RateLimit(proxy.Policies.RateLimitsInbound[endpoint])).
			Configure(envoy_listeners.Tracing(xdsCtx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace), service, envoy_common.TrafficDirectionInbound, "", false)).
			Configure(envoy_listeners.HttpInboundRoutes(service, *routes))
	case core_mesh.ProtocolGRPC:
		filterChainBuilder.
			Configure(envoy_listeners.HttpConnectionManager(localClusterName, true)).
			Configure(envoy_listeners.GrpcStats()).
			Configure(envoy_listeners.FaultInjection(proxy.Policies.FaultInjections[endpoint]...)).
			Configure(envoy_listeners.RateLimit(proxy.Policies.RateLimitsInbound[endpoint])).
			Configure(envoy_listeners.Tracing(xdsCtx.Mesh.GetTracingBackend(proxy.Policies.TrafficTrace), service, envoy_common.TrafficDirectionInbound, "", false)).
			Configure(envoy_listeners.HttpInboundRoutes(service, *routes))
	case core_mesh.ProtocolKafka:
		filterChainBuilder.
			Configure(envoy_listeners.Kafka(localClusterName)).
			Configure(envoy_listeners.TcpProxyDeprecated(localClusterName, envoy_common.NewCluster(envoy_common.WithService(localClusterName))))
	case core_mesh.ProtocolTCP:
		fallthrough
	default:
		// configuration for non-HTTP cases
		filterChainBuilder.Configure(envoy_listeners.TcpProxyDeprecated(localClusterName, envoy_common.NewCluster(envoy_common.WithService(localClusterName))))
	}
	if serverSideMTLS {
		filterChainBuilder.
			Configure(envoy_listeners.ServerSideMTLS(xdsCtx.Mesh.Resource, proxy.SecretsTracker, tlsVersion, ciphers))
	}
	return filterChainBuilder.
		Configure(envoy_listeners.Timeout(defaults_mesh.DefaultInboundTimeout(), protocol))
}

func GenerateRoutes(proxy *core_xds.Proxy, endpoint mesh_proto.InboundInterface, cluster *envoy_common.ClusterImpl) envoy_common.Routes {
	routes := envoy_common.Routes{}

	// Iterate over that RateLimits and generate the relevant Routes.
	// We do assume that the rateLimits resource is sorted, so the most
	// specific source matches come first.
	for _, rl := range proxy.Policies.RateLimitsInbound[endpoint] {
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

	return routes
}
