package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_extensions_filters_network_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/local_ratelimit/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	"github.com/pkg/errors"
)

type Configurer struct {
	From core_xds.Rules
	Http *api.LocalHTTP
	Tcp *api.LocalTCP
	Dataplane          *core_mesh.DataplaneResource
	Protocol core_mesh.Protocol
}

func (c *Configurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if c.Http != nil && (c.Http.Enabled == nil || *c.Http.Enabled) {
		if err := c.configureHttpListener(filterChain); err != nil {
			return err
		}
		// route
	}
	if c.Tcp != nil && (c.Tcp.Enabled == nil || *c.Tcp.Enabled) {
		if err := c.configureTcpListener(filterChain); err != nil {
			return err
		}
	}

	return nil
}

// func (c *Configurer) configureRoutes(rc *envoy_route.RouteConfiguration) error{
// 	for _, vh := range rc.VirtualHosts {
// 		RouteConfig -> Name inbound:service
//		nastepnie
// 		VH -> Name 			service
//		Action
//		Lista route

// 	}

// 	return proto.MarshalAnyDeterministic(config)
// }

func (c *Configurer) configureHttpListener(filterChain *envoy_listener.FilterChain) error{
	config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
	}
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}
	httpRateLimit := func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.HttpFilters = append(hcm.HttpFilters,
			&envoy_hcm.HttpFilter{
				Name: "envoy.filters.http.local_ratelimit",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			})
		return nil
	}
	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpRateLimit); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	return nil
}

func (c *Configurer) configureTcpListener(filterChain *envoy_listener.FilterChain) error{
	config := &envoy_extensions_filters_network_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
		TokenBucket: &typev3.TokenBucket{
			MaxTokens: c.Tcp.Connections,
			TokensPerFill: proto.UInt32(c.Tcp.Connections),
			FillInterval: util_proto.Duration(c.Tcp.Interval.Duration),
		},
	}
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}
	filters := []*envoy_listener.Filter{}
	filters = append(filters,  &envoy_listener.Filter{
		Name: "envoy.extensions.filters.network.local_ratelimit.v3.LocalRateLimit",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	filterChain.Filters = append(filters, filterChain.Filters...)
	return nil
}
