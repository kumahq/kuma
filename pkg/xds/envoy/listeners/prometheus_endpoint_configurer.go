package listeners

import (
	"github.com/golang/protobuf/ptypes"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"

	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

func PrometheusEndpoint(path string, clusterName string) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&PrometheusEndpointConfigurer{
			path:        path,
			clusterName: clusterName,
		})
	})
}

type PrometheusEndpointConfigurer struct {
	path        string
	clusterName string
}

func (c *PrometheusEndpointConfigurer) Configure(l *v2.Listener) error {
	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix: util_xds.SanitizeMetric(l.Name),
		CodecType:  envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: []*envoy_hcm.HttpFilter{{
			Name: envoy_wellknown.Router,
		}},
		RouteSpecifier: &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &v2.RouteConfiguration{
				VirtualHosts: []*envoy_route.VirtualHost{{
					Name:    "envoy_admin",
					Domains: []string{"*"},
					Routes: []*envoy_route.Route{{
						Match: &envoy_route.RouteMatch{
							PathSpecifier: &envoy_route.RouteMatch_Prefix{
								Prefix: c.path,
							},
						},
						Action: &envoy_route.Route_Route{
							Route: &envoy_route.RouteAction{
								ClusterSpecifier: &envoy_route.RouteAction_Cluster{
									Cluster: c.clusterName,
								},
								PrefixRewrite: "/stats/prometheus", // well-known Admin API endpoint
							},
						},
					}},
				}},
			},
		},
	}
	pbst, err := ptypes.MarshalAny(config)
	if err != nil {
		return err
	}

	for i := range l.FilterChains {
		l.FilterChains[i].Filters = append(l.FilterChains[i].Filters, &envoy_listener.Filter{
			Name: envoy_wellknown.HTTPConnectionManager,
			ConfigType: &envoy_listener.Filter_TypedConfig{
				TypedConfig: pbst,
			},
		})
	}

	return nil
}
