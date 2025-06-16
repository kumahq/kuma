package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

type HttpInboundRouteConfigurer struct {
	Name    string
	Service string
	Routes  envoy_common.Routes
}

var _ FilterChainConfigurer = &HttpInboundRouteConfigurer{}

func (c *HttpInboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeName := envoy_names.GetInboundRouteName(c.Service)
	vhName := c.Service
	if c.Name != "" {
		routeName = c.Name
		vhName = c.Name
	}

	static := HttpStaticRouteConfigurer{
		Builder: envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3, routeName).
			Configure(envoy_routes.CommonRouteConfiguration()).
			Configure(envoy_routes.ResetTagsHeader()).
			Configure(envoy_routes.VirtualHost(envoy_virtual_hosts.NewVirtualHostBuilder(envoy_common.APIV3, vhName).
				Configure(envoy_virtual_hosts.Routes(c.Routes)))),
	}

	return static.Configure(filterChain)
}
