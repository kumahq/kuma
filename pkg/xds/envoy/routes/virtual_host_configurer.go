package routes

import (
	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
)

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
