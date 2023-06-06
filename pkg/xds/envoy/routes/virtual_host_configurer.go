package routes

import (
	envoy_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

type RouteConfigurationVirtualHostConfigurerV3 struct {
	builder *envoy_virtual_hosts.VirtualHostBuilder
}

func (c RouteConfigurationVirtualHostConfigurerV3) Configure(routeConfiguration *envoy_route_v3.RouteConfiguration) error {
	virtualHost, err := c.builder.Build()
	if err != nil {
		return err
	}
	routeConfiguration.VirtualHosts = append(routeConfiguration.VirtualHosts, virtualHost.(*envoy_route_v3.VirtualHost))
	return nil
}
