package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

type OutboundRoute struct {
	Hash                    common_api.MatchesHash
	Match                   api.Match
	Filters                 []api.Filter
	Split                   []envoy_common.Split
	BackendRefToClusterName map[common_api.BackendRefHash]string
}

type HttpOutboundRouteConfigurer struct {
	Name    string
	Service string
	Routes  []OutboundRoute
	DpTags  mesh_proto.MultiValueTagSet
}

var _ envoy_listeners_v3.FilterChainConfigurer = &HttpOutboundRouteConfigurer{}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	var name string
	if c.Name == "" {
		name = envoy_names.GetOutboundRouteName(c.Service)
	} else {
		name = c.Name
	}

	virtualHostBuilder := envoy_virtual_hosts.NewVirtualHostBuilder(envoy_common.APIV3, c.Service)
	for _, route := range c.Routes {
		route := envoy_virtual_hosts.AddVirtualHostConfigurer(
			&RoutesConfigurer{
				Hash:                    route.Hash,
				Match:                   route.Match,
				Filters:                 route.Filters,
				Split:                   route.Split,
				BackendRefToClusterName: route.BackendRefToClusterName,
			})
		virtualHostBuilder = virtualHostBuilder.Configure(route)
	}
	static := envoy_listeners_v3.HttpStaticRouteConfigurer{
		Builder: envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3, name).
			Configure(envoy_routes.CommonRouteConfiguration()).
			Configure(envoy_routes.TagsHeader(c.DpTags)).
			Configure(envoy_routes.VirtualHost(virtualHostBuilder)),
	}

	return static.Configure(filterChain)
}
