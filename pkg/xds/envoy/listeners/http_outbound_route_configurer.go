package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

func HttpOutboundRoute(service string, subsets []envoy_common.ClusterSubset, dpTags mesh_proto.MultiValueTagSet) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&HttpOutboundRouteConfigurer{
			service: service,
			subsets: subsets,
			dpTags:  dpTags,
		})
	})
}

type HttpOutboundRouteConfigurer struct {
	service string
	subsets []envoy_common.ClusterSubset
	dpTags  mesh_proto.MultiValueTagSet
}

func (c *HttpOutboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeConfig, err := envoy_routes.NewRouteConfigurationBuilder().
		Configure(envoy_routes.CommonRouteConfiguration(envoy_names.GetOutboundRouteName(c.service))).
		Configure(envoy_routes.TagsHeader(c.dpTags)).
		Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder().
			Configure(envoy_routes.CommonVirtualHost(c.service)).
			Configure(envoy_routes.DefaultRoute(c.subsets...)))).
		Build()
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		}
		return nil
	})
}
