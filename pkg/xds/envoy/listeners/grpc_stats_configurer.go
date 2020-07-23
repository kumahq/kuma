package listeners

import (
	envoy_grpc_stats "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/grpc_stats/v2alpha"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/kumahq/kuma/pkg/util/proto"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
)

type GrpcStatsConfigurer struct {
}

func GrpcStats() FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&GrpcStatsConfigurer{})
	})
}

func (g *GrpcStatsConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_grpc_stats.FilterConfig{
		EmitFilterState: true,
	}
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}
	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.HttpFilters = append([]*envoy_hcm.HttpFilter{
			{
				Name: envoy_wellknown.HTTPGRPCStats,
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			},
		}, manager.HttpFilters...)
		return nil
	})
}
