package listeners

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
	envoy_routes "github.com/Kong/kuma/pkg/xds/envoy/routes"
)

func HttpInboundRoute(service string, cluster envoy_common.ClusterInfo) FilterChainBuilderOpt {
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
	cluster envoy_common.ClusterInfo
}

func (c *HttpInboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeName := fmt.Sprintf("inbound:%s", c.service)
	routeConfig, err := envoy_routes.NewRouteConfigurationBuilder().
		Configure(envoy_routes.CommonRouteConfiguration(routeName)).
		Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder().
			Configure(envoy_routes.CommonVirtualHost(c.service)).
			Configure(envoy_routes.DefaultRoute(c.cluster)))).
		Build()
	if err != nil {
		return err
	}

	return UpdateFilterConfig(filterChain, envoy_wellknown.HTTPConnectionManager, func(filterConfig proto.Message) error {
		hcm, ok := filterConfig.(*envoy_hcm.HttpConnectionManager)
		if !ok {
			return NewUnexpectedFilterConfigTypeError(filterConfig, &envoy_hcm.HttpConnectionManager{})
		}
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig,
		}
		return nil
	})
}
