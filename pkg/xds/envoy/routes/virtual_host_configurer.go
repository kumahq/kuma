package routes

import (
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

func VirtualHost(builder *VirtualHostBuilder) RouteConfigurationBuilderOpt {
	return RouteConfigurationBuilderOptFunc(func(config *RouteConfigurationBuilderConfig) {
		config.Add(&RouteConfigurationVirtualHostConfigurer{
			builder: builder,
		})
	})
}

type RouteConfigurationVirtualHostConfigurer struct {
	builder *VirtualHostBuilder
}

func (c RouteConfigurationVirtualHostConfigurer) Configure(routeConfiguration *v2.RouteConfiguration) error {
	virtualHost, err := c.builder.Build()
	if err != nil {
		return err
	}
	routeConfiguration.VirtualHosts = append(routeConfiguration.VirtualHosts, virtualHost)
	return nil
}
