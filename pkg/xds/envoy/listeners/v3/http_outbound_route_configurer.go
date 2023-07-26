package v3

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

type HttpOutboundRouteConfigurer struct {
	Service string
	Routes  envoy_common.Routes
	DpTags  mesh_proto.MultiValueTagSet
}

var _ FilterChainConfigurer = &HttpOutboundRouteConfigurer{}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	static := HttpStaticRouteConfigurer{
		Builder: envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3, envoy_names.GetOutboundRouteName(c.Service)).
			Configure(envoy_routes.CommonRouteConfiguration()).
			Configure(envoy_routes.TagsHeader(c.DpTags)).
			Configure(envoy_routes.VirtualHost(envoy_virtual_hosts.NewVirtualHostBuilder(envoy_common.APIV3, c.Service).
				Configure(envoy_virtual_hosts.Routes(c.Routes)))),
	}

	return static.Configure(filterChain)
}
