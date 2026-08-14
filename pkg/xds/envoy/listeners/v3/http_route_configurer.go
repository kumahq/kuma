package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	envoy_routes "github.com/kumahq/kuma/v3/pkg/xds/envoy/routes"
)

var _ FilterChainConfigurer = &HttpStaticRouteConfigurer{}

// HttpStaticRouteConfigurer configures a static set of routes into the
// HttpConnectionManager in the filter chain.
type HttpStaticRouteConfigurer struct {
	Builder *envoy_routes.RouteConfigurationBuilder
}

func (c *HttpStaticRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeConfig, err := c.Builder.Build()
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig.(*envoy_route.RouteConfiguration),
		}
		return nil
	})
}
