package routes

import (
	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_route_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

type RouteConfigurationVirtualHostConfigurerV2 struct {
	builder *VirtualHostBuilder
}

func (c RouteConfigurationVirtualHostConfigurerV2) Configure(routeConfiguration *envoy_api_v2.RouteConfiguration) error {
	virtualHost, err := c.builder.Build()
	if err != nil {
		return err
	}
	routeConfiguration.VirtualHosts = append(routeConfiguration.VirtualHosts, virtualHost.(*envoy_route_v2.VirtualHost))
	return nil
}

type RouteConfigurationVirtualHostConfigurerV3 struct {
	builder *VirtualHostBuilder
}

func (c RouteConfigurationVirtualHostConfigurerV3) Configure(routeConfiguration *envoy_route_v3.RouteConfiguration) error {
	virtualHost, err := c.builder.Build()
	if err != nil {
		return err
	}
	routeConfiguration.VirtualHosts = append(routeConfiguration.VirtualHosts, virtualHost.(*envoy_route_v3.VirtualHost))
	return nil
}
