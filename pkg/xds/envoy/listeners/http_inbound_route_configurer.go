package listeners

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
	envoy_names "github.com/Kong/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/Kong/kuma/pkg/xds/envoy/routes"
)

func HttpInboundRoute(service string, cluster envoy_common.ClusterSubset) FilterChainBuilderOpt {
	return FilterChainBuilderOptFunc(func(config *FilterChainBuilderConfig) {
		config.Add(&HttpInboundRouteConfigurer{
			service: service,
			cluster: cluster,
		})
	})
}

type HttpInboundRouteConfigurer struct {
	service string
	// Cluster to forward traffic to.
	cluster envoy_common.ClusterSubset
}

func (c *HttpInboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeName := envoy_names.GetInboundRouteName(c.service)
	routeConfig, err := envoy_routes.NewRouteConfigurationBuilder().
		Configure(envoy_routes.CommonRouteConfiguration(routeName)).
		Configure(envoy_routes.ResetTagsHeader()).
		Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder().
			Configure(envoy_routes.CommonVirtualHost(c.service)).
			Configure(envoy_routes.DefaultRoute(c.cluster)))).
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
