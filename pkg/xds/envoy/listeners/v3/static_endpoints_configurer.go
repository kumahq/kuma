package v3

import (
	envoy_config_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type StaticEndpointsConfigurer struct {
	VirtualHostName string
	Paths           []*envoy_common.StaticEndpointPath
}

var _ FilterChainConfigurer = &StaticEndpointsConfigurer{}

func (c *StaticEndpointsConfigurer) Configure(filterChain *envoy_config_listener.FilterChain) error {
	routes := []*envoy_config_route.Route{}
	for _, p := range c.Paths {
		rb := envoy_routes.NewRouteBuilder(envoy_common.APIV3, envoy_common.AnonymousResource).
			Configure(envoy_routes.RouteMatchPrefixPath(p.Path)).
			Configure(envoy_routes.RouteMustConfigureFunc(func(envoyRoute *envoy_config_route.Route) {
				envoyRoute.Action = &envoy_config_route.Route_Route{
					Route: &envoy_config_route.RouteAction{
						ClusterSpecifier: &envoy_config_route.RouteAction_Cluster{
							Cluster: p.ClusterName,
						},
						PrefixRewrite: p.RewritePath,
					},
				}
			}))

		if p.HeaderExactMatch != "" {
			rb.Configure(envoy_routes.RouteMatchExactHeader(p.Header, p.HeaderExactMatch))
		}
		route, err := rb.Build()
		if err != nil {
			return err
		}
		routes = append(routes, route)
	}

	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix:  util_xds.SanitizeMetric(c.VirtualHostName),
		CodecType:   envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: []*envoy_hcm.HttpFilter{},
		RouteSpecifier: &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &envoy_config_route.RouteConfiguration{
				VirtualHosts: []*envoy_config_route.VirtualHost{{
					Name:    c.VirtualHostName,
					Domains: []string{"*"},
					Routes:  routes,
				}},
				ValidateClusters: util_proto.Bool(false),
			},
		},
	}
	pbst, err := util_proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_config_listener.Filter{
		Name: "envoy.filters.network.http_connection_manager",
		ConfigType: &envoy_config_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}
