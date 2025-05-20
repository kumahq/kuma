package v3

import (
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_dresp "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/direct_response/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"google.golang.org/protobuf/types/known/anypb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

type DirectResponseConfigurer struct {
	VirtualHostName   string
	Endpoints         []DirectResponseEndpoints
	InternalAddresses []core_xds.InternalAddress
}

type DirectResponseEndpoints struct {
	Path       string
	StatusCode uint32
	Response   string
}

var _ FilterChainConfigurer = &DirectResponseConfigurer{}

func (c *DirectResponseConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	httpFilters := []*envoy_hcm.HttpFilter{
		{
			Name: "envoy.filters.http.router",
			ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: &anypb.Any{
					TypeUrl: "type.googleapis.com/envoy.extensions.filters.http.router.v3.Router",
				},
			},
		},
	}

	var routes []*envoy_route.Route
	for _, endpoint := range c.Endpoints {
		routes = append(routes, &envoy_route.Route{
			Match: &envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{
					Prefix: endpoint.Path,
				},
			},
			Name: envoy_common.AnonymousResource,
			Action: &envoy_route.Route_DirectResponse{
				DirectResponse: &envoy_route.DirectResponseAction{
					Status: endpoint.StatusCode,
					Body: &envoy_core_v3.DataSource{
						Specifier: &envoy_core_v3.DataSource_InlineString{InlineString: endpoint.Response},
					},
				},
			},
		})
	}

	config := &envoy_hcm.HttpConnectionManager{
		StatPrefix:  util_xds.SanitizeMetric(c.VirtualHostName),
		CodecType:   envoy_hcm.HttpConnectionManager_AUTO,
		HttpFilters: httpFilters,
		RouteSpecifier: &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &envoy_route.RouteConfiguration{
				VirtualHosts: []*envoy_route.VirtualHost{{
					Name:    c.VirtualHostName,
					Domains: []string{"*"},
					Routes:  routes,
				}},
			},
		},
	}
	if len(c.InternalAddresses) == 0 {
		c.InternalAddresses = core_xds.LocalHostAddresses
	}
	config.InternalAddressConfig = &envoy_hcm.HttpConnectionManager_InternalAddressConfig{
		UnixSockets: false,
		CidrRanges:  core_xds.InternalAddressToEnvoyCIDRs(c.InternalAddresses),
	}
	pbst, err := util_proto.MarshalAnyDeterministic(config)
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

type NetworkDirectResponseConfigurer struct {
	Response []byte
}

var _ FilterChainConfigurer = &NetworkDirectResponseConfigurer{}

func (c *NetworkDirectResponseConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_dresp.Config{
		Response: &envoy_core_v3.DataSource{
			Specifier: &envoy_core_v3.DataSource_InlineBytes{
				InlineBytes: c.Response,
			},
		},
	}

	pbst, err := util_proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filterChain.Filters = append(filterChain.Filters, &envoy_listener.Filter{
		Name: "envoy.filters.network.direct_response",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	return nil
}
