package listeners

import (
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
)

func HttpOutboundRoute(routeName string) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&HttpOutboundRouteConfigurer{
			routeName: routeName,
		})
	})
}

type HttpOutboundRouteConfigurer struct {
	routeName string
}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_Rds{
			Rds: &envoy_hcm.Rds{
				ConfigSource: &envoy_core.ConfigSource{
					ConfigSourceSpecifier: &envoy_core.ConfigSource_Ads{
						Ads: &envoy_core.AggregatedConfigSource{},
					},
				},
				RouteConfigName: c.routeName,
			},
		}
		return nil
	})
}
