package listeners

import (
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/tls"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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

func ServerSideMTLS(ctx xds_context.Context) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.ServerSideMTLSConfigurer{
		Ctx: ctx,
	})
}

func HttpConnectionManager(statsName string, forwardClientCertDetails bool) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.HttpConnectionManagerConfigurer{
		StatsName:                statsName,
		ForwardClientCertDetails: forwardClientCertDetails,
	})
}

func FilterChainMatch(transport string, serverNames, applicationProtocols []string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.FilterChainMatchConfigurer{
		ServerNames:          serverNames,
		TransportProtocol:    transport,
		ApplicationProtocols: applicationProtocols,
	})
}

func SourceMatcher(address string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.SourceMatcherConfigurer{
		Address: address,
	})
}

func NetworkRBAC(statsName string, rbacEnabled bool, permission *core_mesh.TrafficPermissionResource) FilterChainBuilderOpt {
	if !rbacEnabled {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.NetworkRBACConfigurer{
		StatsName:  statsName,
		Permission: permission,
	})
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

func FaultInjection(faultInjections ...*mesh_proto.FaultInjection) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.FaultInjectionConfigurer{
		FaultInjections: faultInjections,
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

func MaxConnectAttempts(retry *core_mesh.RetryResource) FilterChainBuilderOpt {
	if retry == nil || retry.Spec.Conf.GetTcp() == nil {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.MaxConnectAttemptsConfigurer{
		Retry: retry,
	})
}

func Retry(
	retry *core_mesh.RetryResource,
	protocol core_mesh.Protocol,
) FilterChainBuilderOpt {
	if retry == nil {
		return FilterChainBuilderOptFunc(nil)
	}

	return AddFilterChainConfigurer(&v3.RetryConfigurer{
		Retry:    retry,
		Protocol: protocol,
	})
}

func Timeout(timeout *mesh_proto.Timeout_Conf, protocol core_mesh.Protocol) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.TimeoutConfigurer{
		Conf:     timeout,
		Protocol: protocol,
	})
}

// ServerHeader sets the value that the HttpConnectionManager will write
// to the "Server" header in HTTP responses.
func ServerHeader(name string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(
		v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
			hcm.ServerName = name
		}),
	)
}

// EnablePathNormalization enables HTTP request path normalization.
func EnablePathNormalization() FilterChainBuilderOpt {
	return AddFilterChainConfigurer(
		v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
			hcm.NormalizePath = util_proto.Bool(true)
			hcm.MergeSlashes = true

			// TODO(jpeach) set path_with_escaped_slashes_action when we upgrade to Envoy v1.19.
		}),
	)
}

// StripHostPort strips the port component before matching the HTTP host
// header (authority) to the available virtual hosts.
func StripHostPort() FilterChainBuilderOpt {
	return AddFilterChainConfigurer(
		v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
			hcm.StripPortMode = &envoy_hcm.HttpConnectionManager_StripAnyHostPort{
				StripAnyHostPort: true,
			}
		}),
	)
}
