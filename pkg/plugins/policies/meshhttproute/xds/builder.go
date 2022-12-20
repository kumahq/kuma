package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type HttpOutboundRouteConfigurer struct {
	Service  string
	Matches  []api.Match
	Filters  []api.Filter
	Clusters []envoy_common.Cluster
	DpTags   mesh_proto.MultiValueTagSet
}

var _ envoy_listeners_v3.FilterChainConfigurer = &HttpOutboundRouteConfigurer{}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routes := envoy_routes.AddVirtualHostConfigurer(
		&RoutesConfigurer{
			Matches:  c.Matches,
			Filters:  c.Filters,
			Clusters: c.Clusters,
		})

	static := envoy_listeners_v3.HttpStaticRouteConfigurer{
		Builder: envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3).
			Configure(envoy_routes.CommonRouteConfiguration(envoy_names.GetOutboundRouteName(c.Service))).
			Configure(envoy_routes.TagsHeader(c.DpTags)).
			Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder(envoy_common.APIV3).
				Configure(envoy_routes.CommonVirtualHost(c.Service)).
				Configure(routes))),
	}

	return static.Configure(filterChain)
}
