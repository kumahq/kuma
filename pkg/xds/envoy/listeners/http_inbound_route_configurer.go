package listeners

import (
	"github.com/golang/protobuf/proto"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_wellknown "github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

func HttpInboundRoute(cluster ClusterInfo) ListenerBuilderOpt {
	return ListenerBuilderOptFunc(func(config *ListenerBuilderConfig) {
		config.Add(&HttpInboundRouteConfigurer{
			cluster: cluster,
		})
	})
}

type HttpInboundRouteConfigurer struct {
	// Cluster to forward traffic to.
	cluster ClusterInfo
}

func (c *HttpInboundRouteConfigurer) Configure(l *v2.Listener) error {
	config := &v2.RouteConfiguration{
		Name: "inbound",
		VirtualHosts: []*envoy_route.VirtualHost{{
			Name:    "local_service",
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

	return UpdateFilterConfig(l, envoy_wellknown.HTTPConnectionManager, func(filterConfig proto.Message) error {
		hcm, ok := filterConfig.(*envoy_hcm.HttpConnectionManager)
		if !ok {
			return NewUnexpectedFilterConfigTypeError(filterConfig, &envoy_hcm.HttpConnectionManager{})
		}
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: config,
		}
		return nil
	})
}
