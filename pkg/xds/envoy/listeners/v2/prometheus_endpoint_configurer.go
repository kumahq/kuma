package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/golang/protobuf/ptypes/wrappers"

	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type PrometheusEndpointConfigurer struct {
	StatsName   string
	Path        string
	ClusterName string
}

var _ FilterChainConfigurer = &PrometheusEndpointConfigurer{}

func (c *PrometheusEndpointConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix: util_xds.SanitizeMetric(c.StatsName),
		CodecType:  envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: []*envoy_hcm.HttpFilter{{
			Name: "envoy.filters.http.router",
		}},
		RouteSpecifier: &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &envoy_api.RouteConfiguration{
				VirtualHosts: []*envoy_route.VirtualHost{{
					Name:    "envoy_admin",
					Domains: []string{"*"},
					Routes: []*envoy_route.Route{{
						Match: &envoy_route.RouteMatch{
							PathSpecifier: &envoy_route.RouteMatch_Prefix{
								Prefix: c.Path,
							},
						},
						Action: &envoy_route.Route_Route{
							Route: &envoy_route.RouteAction{
								ClusterSpecifier: &envoy_route.RouteAction_Cluster{
									Cluster: c.ClusterName,
								},
								PrefixRewrite: "/stats/prometheus", // well-known Admin API endpoint
							},
						},
					}},
				}},
				ValidateClusters: &wrappers.BoolValue{
					Value: false,
				},
			},
		},
	}
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_listener.Filter{
		Name: "envoy.filters.network.http_connection_manager",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}
