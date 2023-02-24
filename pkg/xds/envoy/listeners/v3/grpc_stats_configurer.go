package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_grpc_stats "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_stats/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type GrpcStatsConfigurer struct{}

var _ FilterChainConfigurer = &GrpcStatsConfigurer{}

func (g *GrpcStatsConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_grpc_stats.FilterConfig{
		EmitFilterState: true,
	}
	pbst, err := util_proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}
	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.HttpFilters = append(manager.HttpFilters,
			&envoy_hcm.HttpFilter{
				Name: "envoy.filters.http.grpc_stats",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			})
		return nil
	})
}
