package v2

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_config_filter_network_local_rate_limit_v2alpha "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/local_rate_limit/v2alpha"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type RateLimitConfigurer struct {
	RateLimit *mesh_proto.RateLimit
}

func (r *RateLimitConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if !r.hasHttpRateLimit() {
		return nil
	}

	config := &envoy_config_filter_network_local_rate_limit_v2alpha.LocalRateLimit{
		StatPrefix: "rate_limit",
	}

	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.HttpFilters = append([]*envoy_hcm.HttpFilter{
			{
				Name: "envoy.filters.http.local_ratelimit",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			},
		}, manager.HttpFilters...)
		return nil
	})
}

func (r *RateLimitConfigurer) hasHttpRateLimit() bool {
	return r.RateLimit != nil && r.RateLimit.GetConf().GetHttp() != nil
}
