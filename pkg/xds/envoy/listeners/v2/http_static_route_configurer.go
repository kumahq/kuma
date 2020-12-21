package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

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
			RouteConfig: routeConfig.(*envoy_api.RouteConfiguration),
		}
		return nil
	})
}
