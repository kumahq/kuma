package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

type HttpOutboundRouteConfigurer struct {
	VirtualHostName string
	RouteConfigName string
	Routes          envoy_common.Routes
	DpTags          mesh_proto.MultiValueTagSet
}

var _ FilterChainConfigurer = &HttpOutboundRouteConfigurer{}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	virtualHost := envoy_virtual_hosts.NewVirtualHostBuilder(envoy_common.APIV3, c.VirtualHostName).
		Configure(envoy_virtual_hosts.Routes(c.Routes))

	routeConfig := envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3, c.RouteConfigName).
		Configure(envoy_routes.CommonRouteConfiguration()).
		Configure(envoy_routes.TagsHeader(c.DpTags)).
		Configure(envoy_routes.VirtualHost(virtualHost))

	return NewHttpStaticRouteConfigurer(routeConfig).Configure(filterChain)
}
