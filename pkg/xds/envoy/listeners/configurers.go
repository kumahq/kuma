package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/tls"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

func GrpcStats() FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.GrpcStatsConfigurer{})
}

func Kafka(statsName string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.KafkaConfigurer{
		StatsName: statsName,
	})
}

func Tracing(backend *mesh_proto.TracingBackend, service string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.TracingConfigurer{
		Backend: backend,
		Service: service,
	})
}

func TLSInspector() ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.TLSInspectorConfigurer{})
}

func OriginalDstForwarder() ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.OriginalDstForwarderConfigurer{})
}

func StaticEndpoints(virtualHostName string, paths []*envoy_common.StaticEndpointPath) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.StaticEndpointsConfigurer{
		VirtualHostName: virtualHostName,
		Paths:           paths,
	})
}

func StaticTlsEndpoints(virtualHostName string, keyPair *tls.KeyPair, paths []*envoy_common.StaticEndpointPath) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.StaticEndpointsConfigurer{
		VirtualHostName: virtualHostName,
		Paths:           paths,
		KeyPair:         keyPair,
	})
}

func ServerSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.ServerSideMTLSConfigurer{
		Ctx:      ctx,
		Metadata: metadata,
	})
}

func HttpConnectionManager(statsName string, forwardClientCertDetails bool) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.HttpConnectionManagerConfigurer{
		StatsName:                statsName,
		ForwardClientCertDetails: forwardClientCertDetails,
	})
}

func FilterChainMatch(transport string, serverNames ...string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.FilterChainMatchConfigurer{
		ServerNames:       serverNames,
		TransportProtocol: transport,
	})
}

func SourceMatcher(address string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.SourceMatcherConfigurer{
		Address: address,
	})
}

func InboundListener(listenerName string, address string, port uint32, protocol core_xds.SocketAddressProtocol) ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.InboundListenerConfigurer{
		Protocol:     protocol,
		ListenerName: listenerName,
		Address:      address,
		Port:         port,
	})
}

func NetworkRBAC(statsName string, rbacEnabled bool, permission *mesh_core.TrafficPermissionResource) FilterChainBuilderOpt {
	if !rbacEnabled {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.NetworkRBACConfigurer{
		StatsName:  statsName,
		Permission: permission,
	})
}

func OutboundListener(listenerName string, address string, port uint32, protocol core_xds.SocketAddressProtocol) ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.OutboundListenerConfigurer{
		Protocol:     protocol,
		ListenerName: listenerName,
		Address:      address,
		Port:         port,
	})
}

func TransparentProxying(transparentProxying *mesh_proto.Dataplane_Networking_TransparentProxying) ListenerBuilderOpt {
	virtual := transparentProxying.GetRedirectPortOutbound() != 0 && transparentProxying.GetRedirectPortInbound() != 0
	if virtual {
		return AddListenerConfigurer(&v3.TransparentProxyingConfigurer{})
	}

	return ListenerBuilderOptFunc(nil)
}

func NoBindToPort() ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.TransparentProxyingConfigurer{})
}

func TcpProxy(statsName string, clusters ...envoy_common.Cluster) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.TcpProxyConfigurer{
		StatsName:   statsName,
		Clusters:    clusters,
		UseMetadata: false,
	})
}

func TcpProxyWithMetadata(statsName string, clusters ...envoy_common.Cluster) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.TcpProxyConfigurer{
		StatsName:   statsName,
		Clusters:    clusters,
		UseMetadata: true,
	})
}

func FaultInjection(faultInjection *mesh_proto.FaultInjection) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.FaultInjectionConfigurer{
		FaultInjection: faultInjection,
	})
}

func RateLimit(rateLimits []*mesh_proto.RateLimit) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.RateLimitConfigurer{
		RateLimits: rateLimits,
	})
}

func NetworkAccessLog(
	mesh string,
	trafficDirection envoy_common.TrafficDirection,
	sourceService string,
	destinationService string,
	backend *mesh_proto.LoggingBackend,
	proxy *core_xds.Proxy,
) FilterChainBuilderOpt {
	if backend == nil {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.NetworkAccessLogConfigurer{
		AccessLogConfigurer: v3.AccessLogConfigurer{
			Mesh:               mesh,
			TrafficDirection:   trafficDirection,
			SourceService:      sourceService,
			DestinationService: destinationService,
			Backend:            backend,
			Proxy:              proxy,
		},
	})
}

func HttpAccessLog(
	mesh string,
	trafficDirection envoy_common.TrafficDirection,
	sourceService string, destinationService string,
	backend *mesh_proto.LoggingBackend,
	proxy *core_xds.Proxy,
) FilterChainBuilderOpt {
	if backend == nil {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.HttpAccessLogConfigurer{
		AccessLogConfigurer: v3.AccessLogConfigurer{
			Mesh:               mesh,
			TrafficDirection:   trafficDirection,
			SourceService:      sourceService,
			DestinationService: destinationService,
			Backend:            backend,
			Proxy:              proxy,
		},
	})
}

func HttpStaticRoute(builder *envoy_routes.RouteConfigurationBuilder) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.HttpStaticRouteConfigurer{
		Builder: builder,
	})
}

// HttpDynamicRoute configures the listener filter chain to dynamically request
// the named RouteConfiguration.
func HttpDynamicRoute(name string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.HttpDynamicRouteConfigurer{
		RouteName: name,
	})
}

func HttpInboundRoutes(service string, routes envoy_common.Routes) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.HttpInboundRouteConfigurer{
		Service: service,
		Routes:  routes,
	})
}

func HttpOutboundRoute(service string, routes envoy_common.Routes, dpTags mesh_proto.MultiValueTagSet) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.HttpOutboundRouteConfigurer{
		Service: service,
		Routes:  routes,
		DpTags:  dpTags,
	})
}

func FilterChain(builder *FilterChainBuilder) ListenerBuilderOpt {
	return AddListenerConfigurer(&ListenerFilterChainConfigurerV3{
		builder: builder,
	})
}

func MaxConnectAttempts(retry *mesh_core.RetryResource) FilterChainBuilderOpt {
	if retry == nil || retry.Spec.Conf.GetTcp() == nil {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.MaxConnectAttemptsConfigurer{
		Retry: retry,
	})
}

func Retry(
	retry *mesh_core.RetryResource,
	protocol mesh_core.Protocol,
) FilterChainBuilderOpt {
	if retry == nil {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.RetryConfigurer{
		Retry:    retry,
		Protocol: protocol,
	})
}

func Timeout(timeout *mesh_proto.Timeout_Conf, protocol mesh_core.Protocol) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.TimeoutConfigurer{
		Conf:     timeout,
		Protocol: protocol,
	})
}

func DNS(vips map[string]string, emptyDnsPort uint32) ListenerBuilderOpt {
	return AddListenerConfigurer(&v3.DNSConfigurer{
		VIPs:         vips,
		EmptyDNSPort: emptyDnsPort,
	})
}

func ConnectionBufferLimit(bytes uint32) ListenerBuilderOpt {
	return AddListenerConfigurer(
		v3.ListenerMustConfigureFunc(func(l *envoy_listener.Listener) {
			l.PerConnectionBufferLimitBytes = wrapperspb.UInt32(bytes)
		}))
}

func EnableReusePort(enable bool) ListenerBuilderOpt {
	return AddListenerConfigurer(
		v3.ListenerMustConfigureFunc(func(l *envoy_listener.Listener) {
			// TODO(jpeach) in Envoy 1.20, this field is deprecated in favor of EnableReusePort.
			l.ReusePort = enable
		}))
}

func EnableFreebind(enable bool) ListenerBuilderOpt {
	return AddListenerConfigurer(
		v3.ListenerMustConfigureFunc(func(l *envoy_listener.Listener) {
			l.Freebind = wrapperspb.Bool(enable)
		}))
}
