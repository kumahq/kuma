package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type HttpInboundRouteConfigurer struct {
	Service string
	Routes  envoy_common.Routes
}

var _ FilterChainConfigurer = &HttpInboundRouteConfigurer{}

func (c *HttpInboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeName := envoy_names.GetInboundRouteName(c.Service)

	static := HttpStaticRouteConfigurer{
		Builder: envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3).
			Configure(envoy_routes.CommonRouteConfiguration(routeName)).
			Configure(envoy_routes.ResetTagsHeader()).
			Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder(envoy_common.APIV3).
				Configure(envoy_routes.CommonVirtualHost(c.Service)).
				Configure(envoy_routes.Routes(c.Routes)))),
	}

	return static.Configure(filterChain)
}
