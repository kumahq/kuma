package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	"github.com/golang/protobuf/ptypes/wrappers"

	envoy_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls/v2"

	"github.com/kumahq/kuma/pkg/tls"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"

	"github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

type StaticEndpointsConfigurer struct {
	VirtualHostName string
	Paths           []*envoy_common.StaticEndpointPath
	KeyPair         *tls.KeyPair
}

var _ FilterChainConfigurer = &StaticEndpointsConfigurer{}

func (c *StaticEndpointsConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routes := []*envoy_route.Route{}
	for _, p := range c.Paths {
		route := &envoy_route.Route{
			Match: &envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{
					Prefix: p.Path,
				},
			},
			Action: &envoy_route.Route_Route{
				Route: &envoy_route.RouteAction{
					ClusterSpecifier: &envoy_route.RouteAction_Cluster{
						Cluster: p.ClusterName,
					},
					PrefixRewrite: p.RewritePath,
				},
			},
		}

		if p.HeaderExactMatch != "" {
			route.Match.Headers = []*envoy_route.HeaderMatcher{{
				Name: p.Header,
				HeaderMatchSpecifier: &envoy_route.HeaderMatcher_ExactMatch{
					ExactMatch: p.HeaderExactMatch,
				},
			}}
		}

		routes = append(routes, route)
	}

	sanitizedVirtualHostName := util_xds.SanitizeMetric(c.VirtualHostName)
	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix: sanitizedVirtualHostName,
		CodecType:  envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: []*envoy_hcm.HttpFilter{{
			Name: "envoy.filters.http.router",
		}},
		RouteSpecifier: &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &envoy_api.RouteConfiguration{
				VirtualHosts: []*envoy_route.VirtualHost{{
					Name:    sanitizedVirtualHostName,
					Domains: []string{"*"},
					Routes:  routes,
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

	if c.KeyPair != nil {
		tlsContext := envoy_tls.StaticDownstreamTlsContext(c.KeyPair)
		pbst, err = proto.MarshalAnyDeterministic(tlsContext)
		if err != nil {
			return err
		}

		filterChain.TransportSocket = &envoy_core.TransportSocket{
			Name: "envoy.transport_sockets.tls",
			ConfigType: &envoy_core.TransportSocket_TypedConfig{
				TypedConfig: pbst,
			},
		}
	}

	return nil
}
