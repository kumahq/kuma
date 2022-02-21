package listeners

import (
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_compression_gzip_compressor_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	envoy_extensions_filters_http_compressor_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
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

func ServerSideMTLS(mesh *core_mesh.MeshResource) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.ServerSideMTLSConfigurer{
		Mesh: mesh,
	})
}

func ServerSideMTLSWithCP(ctx xds_context.Context) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.ServerSideMTLSWithCPConfigurer{
		Ctx: ctx,
	})
}

func HttpConnectionManager(statsName string, forwardClientCertDetails bool) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.HttpConnectionManagerConfigurer{
		StatsName:                statsName,
		ForwardClientCertDetails: forwardClientCertDetails,
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

func FaultInjection(faultInjections ...*core_mesh.FaultInjectionResource) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.FaultInjectionConfigurer{
		FaultInjections: faultInjections,
	})
}

func RateLimit(rateLimits []*core_mesh.RateLimitResource) FilterChainBuilderOpt {
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

// DefaultCompressorFilter adds a gzip compressor filter in its default configuration.
func DefaultCompressorFilter() FilterChainBuilderOpt {
	return AddFilterChainConfigurer(
		v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
			c := envoy_extensions_filters_http_compressor_v3.Compressor{
				CompressorLibrary: &envoy_config_core.TypedExtensionConfig{
					Name:        "gzip",
					TypedConfig: util_proto.MustMarshalAny(&envoy_extensions_compression_gzip_compressor_v3.Gzip{}),
				},
				ResponseDirectionConfig: &envoy_extensions_filters_http_compressor_v3.Compressor_ResponseDirectionConfig{
					DisableOnEtagHeader: true,
				},
			}

			gzip := &envoy_hcm.HttpFilter{
				Name: "gzip-compress",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: util_proto.MustMarshalAny(&c),
				},
			}

			hcm.HttpFilters = append(hcm.HttpFilters, gzip)
		}),
	)
}
