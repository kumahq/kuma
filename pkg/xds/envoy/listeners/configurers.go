package listeners

import (
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
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.GrpcStatsConfigurer{})
	})
}

func Kafka(statsName string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.KafkaConfigurer{
			StatsName: statsName,
		})
	})
}

func Tracing(backend *mesh_proto.TracingBackend, service string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.TracingConfigurer{
			Backend: backend,
			Service: service,
		})
	})
}

func TLSInspector() ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(&v3.TLSInspectorConfigurer{})
	})
}

func OriginalDstForwarder() ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(&v3.OriginalDstForwarderConfigurer{})
	})
}

func StaticEndpoints(virtualHostName string, paths []*envoy_common.StaticEndpointPath) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.StaticEndpointsConfigurer{
			VirtualHostName: virtualHostName,
			Paths:           paths,
		})
	})
}

func StaticTlsEndpoints(virtualHostName string, keyPair *tls.KeyPair, paths []*envoy_common.StaticEndpointPath) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.StaticEndpointsConfigurer{
			VirtualHostName: virtualHostName,
			Paths:           paths,
			KeyPair:         keyPair,
		})
	})
}

func ServerSideMTLS(ctx xds_context.Context, metadata *core_xds.DataplaneMetadata) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.ServerSideMTLSConfigurer{
			Ctx:      ctx,
			Metadata: metadata,
		})
	})
}

func HttpConnectionManager(statsName string, forwardClientCertDetails bool) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.HttpConnectionManagerConfigurer{
			StatsName:                statsName,
			ForwardClientCertDetails: forwardClientCertDetails,
		})
	})
}

func FilterChainMatch(transport string, serverNames ...string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.FilterChainMatchConfigurer{
			ServerNames:       serverNames,
			TransportProtocol: transport,
		})
	})
}

func SourceMatcher(address string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.SourceMatcherConfigurer{
			Address: address,
		})
	})
}

func InboundListener(listenerName string, address string, port uint32, protocol core_xds.SocketAddressProtocol) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(&v3.InboundListenerConfigurer{
			Protocol:     protocol,
			ListenerName: listenerName,
			Address:      address,
			Port:         port,
		})
	})
}

func NetworkRBAC(statsName string, rbacEnabled bool, permission *mesh_core.TrafficPermissionResource) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if rbacEnabled {
			config.AddV3(&v3.NetworkRBACConfigurer{
				StatsName:  statsName,
				Permission: permission,
			})
		}
	})
}

func OutboundListener(listenerName string, address string, port uint32, protocol core_xds.SocketAddressProtocol) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(&v3.OutboundListenerConfigurer{
			Protocol:     protocol,
			ListenerName: listenerName,
			Address:      address,
			Port:         port,
		})
	})
}

func TransparentProxying(transparentProxying *mesh_proto.Dataplane_Networking_TransparentProxying) ListenerBuilderOpt {
	virtual := transparentProxying.GetRedirectPortOutbound() != 0 && transparentProxying.GetRedirectPortInbound() != 0
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		if virtual {
			config.AddV3(&v3.TransparentProxyingConfigurer{})
		}
	})
}

func NoBindToPort() ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(&v3.TransparentProxyingConfigurer{})
	})
}

func TcpProxy(statsName string, clusters ...envoy_common.Cluster) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.TcpProxyConfigurer{
			StatsName:   statsName,
			Clusters:    clusters,
			UseMetadata: false,
		})
	})
}

func TcpProxyWithMetadata(statsName string, clusters ...envoy_common.Cluster) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.TcpProxyConfigurer{
			StatsName:   statsName,
			Clusters:    clusters,
			UseMetadata: true,
		})
	})
}

func FaultInjection(faultInjection *mesh_proto.FaultInjection) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.FaultInjectionConfigurer{
			FaultInjection: faultInjection,
		})
	})
}

func RateLimit(rateLimits []*mesh_proto.RateLimit) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.RateLimitConfigurer{
			RateLimits: rateLimits,
		})
	})
}

func NetworkAccessLog(mesh string, trafficDirection envoy_common.TrafficDirection, sourceService string, destinationService string, backend *mesh_proto.LoggingBackend, proxy *core_xds.Proxy) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if backend != nil {
			config.AddV3(&v3.NetworkAccessLogConfigurer{
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
	})
}

func HttpAccessLog(mesh string, trafficDirection envoy_common.TrafficDirection, sourceService string, destinationService string, backend *mesh_proto.LoggingBackend, proxy *core_xds.Proxy) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if backend != nil {
			config.AddV3(&v3.HttpAccessLogConfigurer{
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
	})
}

func HttpStaticRoute(builder *envoy_routes.RouteConfigurationBuilder) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.HttpStaticRouteConfigurer{
			Builder: builder,
		})
	})
}

func HttpInboundRoutes(service string, routes envoy_common.Routes) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.HttpInboundRoutesConfigurer{
			Service: service,
			Routes:  routes,
		})
	})
}

func HttpOutboundRoute(service string, routes envoy_common.Routes, dpTags mesh_proto.MultiValueTagSet) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.HttpOutboundRouteConfigurer{
			Service: service,
			Routes:  routes,
			DpTags:  dpTags,
		})
	})
}

func FilterChain(builder *FilterChainBuilder) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(&ListenerFilterChainConfigurerV3{
			builder: builder,
		})
	})
}

func MaxConnectAttempts(retry *mesh_core.RetryResource) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if retry != nil && retry.Spec.Conf.GetTcp() != nil {
			config.AddV3(&v3.MaxConnectAttemptsConfigurer{
				Retry: retry,
			})
		}
	})
}

func Retry(
	retry *mesh_core.RetryResource,
	protocol mesh_core.Protocol,
) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		if retry != nil {
			config.AddV3(&v3.RetryConfigurer{
				Retry:    retry,
				Protocol: protocol,
			})
		}
	})
}

func Timeout(timeout *mesh_proto.Timeout_Conf, protocol mesh_core.Protocol) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.AddV3(&v3.TimeoutConfigurer{
			Conf:     timeout,
			Protocol: protocol,
		})
	})
}

func DNS(vips map[string]string, emptyDnsPort uint32) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.AddV3(&v3.DNSConfigurer{
			VIPs:         vips,
			EmptyDNSPort: emptyDnsPort,
		})
	})
}
