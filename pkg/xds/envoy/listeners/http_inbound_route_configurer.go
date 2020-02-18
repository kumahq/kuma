package listeners

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
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
	routeConfig := c.routeConfiguration()

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

func (c *HttpInboundRouteConfigurer) routeConfiguration() *v2.RouteConfiguration {
	return &v2.RouteConfiguration{
		Name: fmt.Sprintf("inbound:%s", c.service),
		VirtualHosts: []*envoy_route.VirtualHost{{
			Name:    c.service,
			Domains: []string{"*"},
			Routes: []*envoy_route.Route{{
				Match: &envoy_route.RouteMatch{
					PathSpecifier: &envoy_route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &envoy_route.Route_Route{
					Route: &envoy_route.RouteAction{
						ClusterSpecifier: &envoy_route.RouteAction_Cluster{
							Cluster: c.cluster.Name,
						},
					},
				},
			}},
		}},
		ValidateClusters: &wrappers.BoolValue{
			Value: true,
		},
	}
}
