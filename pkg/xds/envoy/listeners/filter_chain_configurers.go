package listeners

import (
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_extensions_compression_gzip_compressor_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/compression/gzip/compressor/v3"
	envoy_extensions_filters_http_compressor_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/compressor/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	common_tls "github.com/kumahq/kuma/api/common/v1alpha1/tls"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_routes_v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

func GrpcStats() FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.GrpcStatsConfigurer{})
}

func Kafka(statsName string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.KafkaConfigurer{
		StatsName: statsName,
	})
}

func Tracing(
	backend *mesh_proto.TracingBackend,
	service string,
	direction envoy_common.TrafficDirection,
	destination string,
	spawnUpstreamSpan bool,
) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.TracingConfigurer{
		Backend:           backend,
		Service:           service,
		TrafficDirection:  direction,
		Destination:       destination,
		SpawnUpstreamSpan: spawnUpstreamSpan,
	})
}

func StaticEndpoints(virtualHostName string, paths []*envoy_common.StaticEndpointPath) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.StaticEndpointsConfigurer{
		VirtualHostName: virtualHostName,
		Paths:           paths,
	})
}

func DirectResponse(virtualHostName string, endpoints []v3.DirectResponseEndpoints) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.DirectResponseConfigurer{
		VirtualHostName: virtualHostName,
		Endpoints:       endpoints,
	})
}

func NetworkDirectResponse(response string) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.NetworkDirectResponseConfigurer{
		Response: []byte(response),
	})
}

func ServerSideMTLS(mesh *core_mesh.MeshResource, secrets core_xds.SecretsTracker, tlsVersion *common_tls.Version, tlsCiphers common_tls.TlsCiphers) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.ServerSideMTLSConfigurer{
		Mesh:           mesh,
		SecretsTracker: secrets,
		TlsVersion:     tlsVersion,
		TlsCiphers:     tlsCiphers,
	})
}

func ServerSideStaticMTLS(mtlsCerts core_xds.ServerSideMTLSCerts) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.ServerSideStaticMTLSConfigurer{
		MTLSCerts: mtlsCerts,
	})
}

func ServerSideStaticTLS(tlsCerts core_xds.ServerSideTLSCertPaths) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.ServerSideStaticTLSConfigurer{
		CertPath: tlsCerts.CertPath,
		KeyPath:  tlsCerts.KeyPath,
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

type splitAdapter struct {
	clusterName string
	weight      uint32
	lbMetadata  tags.Tags

	hasExternalService bool
}

func (s *splitAdapter) ClusterName() string      { return s.clusterName }
func (s *splitAdapter) Weight() uint32           { return s.weight }
func (s *splitAdapter) LBMetadata() tags.Tags    { return s.lbMetadata }
func (s *splitAdapter) HasExternalService() bool { return s.hasExternalService }

func TcpProxyDeprecated(statsName string, clusters ...envoy_common.Cluster) FilterChainBuilderOpt {
	var splits []envoy_common.Split
	for _, cluster := range clusters {
		cluster := cluster.(*envoy_common.ClusterImpl)
		splits = append(splits, &splitAdapter{
			clusterName:        cluster.Name(),
			weight:             cluster.Weight(),
			lbMetadata:         cluster.Tags(),
			hasExternalService: cluster.IsExternalService(),
		})
	}
	return AddFilterChainConfigurer(&v3.TcpProxyConfigurer{
		StatsName:   statsName,
		Splits:      splits,
		UseMetadata: false,
	})
}

func TcpProxyDeprecatedWithMetadata(statsName string, clusters ...envoy_common.Cluster) FilterChainBuilderOpt {
	var splits []envoy_common.Split
	for _, cluster := range clusters {
		cluster := cluster.(*envoy_common.ClusterImpl)
		splits = append(splits, &splitAdapter{
			clusterName:        cluster.Name(),
			weight:             cluster.Weight(),
			lbMetadata:         cluster.Tags(),
			hasExternalService: cluster.IsExternalService(),
		})
	}
	return AddFilterChainConfigurer(&v3.TcpProxyConfigurer{
		StatsName:   statsName,
		Splits:      splits,
		UseMetadata: true,
	})
}

func TCPProxy(statsName string, splits ...envoy_common.Split) FilterChainBuilderOpt {
	return AddFilterChainConfigurer(&v3.TcpProxyConfigurer{
		StatsName:   statsName,
		Splits:      splits,
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
	sourceService string,
	destinationService string,
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

	return AddFilterChainConfigurer(
		v3.HttpConnectionManagerMustConfigureFunc(func(hcm *envoy_hcm.HttpConnectionManager) {
			for _, virtualHost := range hcm.GetRouteConfig().VirtualHosts {
				virtualHost.RetryPolicy = envoy_routes_v3.RetryConfig(retry, protocol)
			}
		}))
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
			hcm.PathWithEscapedSlashesAction = envoy_hcm.HttpConnectionManager_UNESCAPE_AND_REDIRECT
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
