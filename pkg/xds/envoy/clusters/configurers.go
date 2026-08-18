package clusters

import (
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters/v3"
)

func ClientSideTLS(endpoints []core_xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.ClientSideTLSConfigurer{
			Endpoints: endpoints,
		})
	})
}

func MeshExternalServiceClientSideTLS(endpoints []core_xds.Endpoint, systemCaPath string, useCommonTlsContext bool) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.ClientSideTLSConfigurer{
			Endpoints:           endpoints,
			SystemCaPath:        systemCaPath,
			UseCommonTlsContext: useCommonTlsContext,
		})
	})
}

func EdsCluster() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.EdsClusterConfigurer{})
		builder.AddConfigurer(&v3.AltStatNameConfigurer{})
	})
}

// ProvidedEndpointCluster sets the cluster with the defined endpoints, this is useful when endpoints are not discovered using EDS, so we don't use EdsCluster
func ProvidedEndpointCluster(hasIPv6 bool, endpoints ...core_xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.ProvidedEndpointClusterConfigurer{
			Name:      builder.name,
			Endpoints: endpoints,
			HasIPv6:   hasIPv6,
		})
		builder.AddConfigurer(&v3.AltStatNameConfigurer{})
	})
}

func StaticNoEndpointsCluster() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.StaticClusterConfigurer{
			Name: builder.name,
		})
		builder.AddConfigurer(&v3.AltStatNameConfigurer{})
	})
}

func ProvidedCustomEndpointCluster(hasIPv6 bool, allowsMixingEndpoints bool, endpoints ...core_xds.Endpoint) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.ProvidedEndpointClusterConfigurer{
			Name:                           builder.name,
			Endpoints:                      endpoints,
			HasIPv6:                        hasIPv6,
			AllowMixingIpAndNonIpEndpoints: allowsMixingEndpoints,
		})
		builder.AddConfigurer(&v3.AltStatNameConfigurer{})
	})
}

func Timeout(timeout envoy_common.Timeouts, protocol core_meta.Protocol) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.TimeoutConfigurer{
			Protocol: protocol,
			Conf:     timeout,
		})
	})
}

func DefaultTimeout() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.TimeoutConfigurer{
			Protocol: core_meta.ProtocolTCP,
		})
	})
}

func PassThroughCluster() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.PassThroughClusterConfigurer{})
		builder.AddConfigurer(&v3.AltStatNameConfigurer{})
	})
}

func UpstreamBindConfig(address string, port uint32) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.UpstreamBindConfigConfigurer{
			Address: address,
			Port:    port,
		})
	})
}

func Http2() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.Http2Configurer{})
	})
}

func Http() ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.HttpConfigurer{})
	})
}

func UpstreamTLSContext(config *envoy_tls.UpstreamTlsContext) ClusterBuilderOpt {
	return ClusterBuilderOptFunc(func(builder *ClusterBuilder) {
		builder.AddConfigurer(&v3.UpstreamTLSContextConfigure{
			Config: config,
		})
	})
}
