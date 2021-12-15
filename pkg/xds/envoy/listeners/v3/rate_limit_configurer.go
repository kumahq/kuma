package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/util/proto"
)

type RateLimitConfigurer struct {
	RateLimits []*core_mesh.RateLimitResource
}

func (r *RateLimitConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if !r.hasHttpRateLimit() {
		return nil
	}

	config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
	}

	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(manager *envoy_hcm.HttpConnectionManager) error {
		manager.HttpFilters = append(manager.HttpFilters,
			&envoy_hcm.HttpFilter{
				Name: "envoy.filters.http.local_ratelimit",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			})
		return nil
	})
}

func (r *RateLimitConfigurer) hasHttpRateLimit() bool {
	return len(r.RateLimits) > 0
}
