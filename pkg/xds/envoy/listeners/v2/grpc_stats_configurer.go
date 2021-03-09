package v2

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_grpc_stats "github.com/envoyproxy/go-control-plane/envoy/config/filter/http/grpc_stats/v2alpha"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/kumahq/kuma/pkg/util/proto"
)

type GrpcStatsConfigurer struct {
}

var _ FilterChainConfigurer = &GrpcStatsConfigurer{}

func (g *GrpcStatsConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_grpc_stats.FilterConfig{
		EmitFilterState: true,
		PerMethodStatSpecifier: &envoy_grpc_stats.FilterConfig_StatsForAllMethods{
			StatsForAllMethods: &wrappers.BoolValue{Value: true},
		},
	}
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}
	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.HttpFilters = append([]*envoy_hcm.HttpFilter{
			{
				Name: "envoy.filters.http.grpc_stats",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			},
		}, manager.HttpFilters...)
		return nil
	})
}
