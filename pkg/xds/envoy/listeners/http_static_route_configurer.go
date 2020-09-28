package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

func HttpStaticRoute(builder *envoy_routes.RouteConfigurationBuilder) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&HttpStaticRouteConfigurer{
			builder: builder,
		})
	})
}

type HttpStaticRouteConfigurer struct {
	builder *envoy_routes.RouteConfigurationBuilder
}

func (c *HttpStaticRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeConfig, err := c.builder.Build()
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		}
		return nil
	})
}
