package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type HttpOutboundRouteConfigurer struct {
	Service string
	Subsets []envoy_common.ClusterSubset
	DpTags  mesh_proto.MultiValueTagSet
}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeConfig, err := envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV2).
		Configure(envoy_routes.CommonRouteConfiguration(envoy_names.GetOutboundRouteName(c.Service))).
		Configure(envoy_routes.TagsHeader(c.DpTags)).
		Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder(envoy_common.APIV2).
			Configure(envoy_routes.CommonVirtualHost(c.Service)).
			Configure(envoy_routes.DefaultRoute(c.Subsets...)))).
		Build()
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig.(*envoy_api.RouteConfiguration),
		}
		return nil
	})
}
