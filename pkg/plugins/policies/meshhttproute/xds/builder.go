package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
	envoy_routes "github.com/kumahq/kuma/v3/pkg/xds/envoy/routes"
	envoy_virtual_hosts "github.com/kumahq/kuma/v3/pkg/xds/envoy/virtualhosts"
)

type OutboundRoute struct {
	Name    string
	Match   api.Match
	Filters []api.Filter
	// AllBackendRefsUnresolved is true when the rule declares backendRefs and
	// none of them resolve, which is the case the Gateway API answers with 500.
	AllBackendRefsUnresolved bool
	// MirrorSplits contains splits of the clusters created for RequestMirror
	// filters, keyed by the index of the filter in Filters
	MirrorSplits map[int]envoy_common.Split
	Split        []envoy_common.Split
}

type HttpOutboundRouteConfigurer struct {
	RouteConfigName string
	VirtualHostName string
	Routes          []OutboundRoute
	DpTags          mesh_proto.MultiValueTagSet
}

var _ envoy_listeners_v3.FilterChainConfigurer = &HttpOutboundRouteConfigurer{}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	virtualHostBuilder := envoy_virtual_hosts.NewVirtualHostBuilder(envoy_common.APIV3, c.VirtualHostName)
	for _, route := range c.Routes {
		route := envoy_virtual_hosts.AddVirtualHostConfigurer(
			&RoutesConfigurer{
				Name:                     route.Name,
				Match:                    route.Match,
				Filters:                  route.Filters,
				AllBackendRefsUnresolved: route.AllBackendRefsUnresolved,
				MirrorSplits:             route.MirrorSplits,
				Split:                    route.Split,
			})
		virtualHostBuilder = virtualHostBuilder.Configure(route)
	}
	static := envoy_listeners_v3.HttpStaticRouteConfigurer{
		Builder: envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3, c.RouteConfigName).
			Configure(envoy_routes.CommonRouteConfiguration()).
			Configure(envoy_routes.TagsHeader(c.DpTags)).
			Configure(envoy_routes.VirtualHost(virtualHostBuilder)),
	}

	return static.Configure(filterChain)
}
