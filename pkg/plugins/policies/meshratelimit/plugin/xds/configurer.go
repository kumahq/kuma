package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_extensions_filters_network_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/local_ratelimit/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"

	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	rate_limit "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

func RateLimitConfigurationFromPolicy(rl *api.LocalHTTP) *rate_limit.RateLimitConfiguration {
	headers := []*rate_limit.Headers{}
	if rl.OnRateLimit != nil {
		for _, h := range rl.OnRateLimit.Headers {
			header := &rate_limit.Headers{
				Key:   h.Key,
				Value: h.Value,
			}
			if h.Append != nil {
				header.Append = *h.Append
			}
			headers = append(headers, header)
		}
	}
	var status uint32
	if rl.OnRateLimit != nil && rl.OnRateLimit.Status != nil {
		status = *rl.OnRateLimit.Status
	}
	return &rate_limit.RateLimitConfiguration{
		Interval: rl.Interval.Duration,
		Requests: rl.Requests,
		OnRateLimit: &rate_limit.OnRateLimit{
			Status:  status,
			Headers: headers,
		},
	}
}

type Configurer struct {
	Http *api.LocalHTTP
	Tcp  *api.LocalTCP
}

func (c *Configurer) ConfigureFilterChain(filterChain *envoy_listener.FilterChain) error {
	if c.Http != nil {
		if err := c.configureHttpListener(filterChain); err != nil {
			return err
		}
	}
	if c.Tcp != nil {
		if err := c.configureTcpListener(filterChain); err != nil {
			return err
		}
	}
	return nil
}

func (c *Configurer) ConfigureRoute(route *envoy_route.RouteConfiguration) error {
	if route == nil {
		return nil
	}
	rateLimit, err := rate_limit.NewRateLimitConfiguration(RateLimitConfigurationFromPolicy(c.Http))
	if err != nil {
		return err
	}
	for _, vh := range route.VirtualHosts {
		for _, r := range vh.Routes {
			c.addRateLimitToRoute(r, rateLimit)
		}
	}
	return nil
}

func (c *Configurer) configureHttpListener(filterChain *envoy_listener.FilterChain) error {
	rateLimit, err := rate_limit.NewRateLimitConfiguration(RateLimitConfigurationFromPolicy(c.Http))
	if err != nil {
		return err
	}
	listenerConfig := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
	}
	pbstListener, err := proto.MarshalAnyDeterministic(listenerConfig)
	if err != nil {
		return err
	}

	httpRoutes := func(hcm *envoy_hcm.HttpConnectionManager) error {
		// gateway has dynamic routes so it shouldn't be changed here
		for _, vh := range hcm.GetRouteConfig().GetVirtualHosts() {
			routes := vh.GetRoutes()
			// when size is larger than 1 it means that old ratelimit is applied
			if len(routes) > 1 {
				return nil
			}
			for _, route := range routes {
				c.addRateLimitToRoute(route, rateLimit)
			}
		}

		for _, filter := range hcm.HttpFilters {
			if filter.Name == "envoy.filters.http.local_ratelimit" {
				return nil
			}
		}
		// envoy.filters.http.router has to be the last filter
		filters := []*envoy_hcm.HttpFilter{}
		for _, filter := range hcm.HttpFilters {
			if filter.Name == "envoy.filters.http.router" {
				filters = append(filters,
					&envoy_hcm.HttpFilter{
						Name: "envoy.filters.http.local_ratelimit",
						ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
							TypedConfig: pbstListener,
						},
					})
			}
			filters = append(filters, filter)
		}
		hcm.HttpFilters = filters
		return nil
	}
	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpRoutes); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	return nil
}

func (c *Configurer) configureTcpListener(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_extensions_filters_network_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "tcp_rate_limit",
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens:     c.Tcp.Connections,
			TokensPerFill: proto.UInt32(c.Tcp.Connections),
			FillInterval:  proto.Duration(c.Tcp.Interval.Duration),
		},
	}
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}

	filters := []*envoy_listener.Filter{}
	filters = append(filters, &envoy_listener.Filter{
		Name: "envoy.extensions.filters.network.local_ratelimit.v3.LocalRateLimit",
		ConfigType: &envoy_listener.Filter_TypedConfig{
			TypedConfig: pbst,
		},
	})
	filterChain.Filters = append(filters, filterChain.Filters...)
	return nil
}

func (c *Configurer) addRateLimitToRoute(route *envoy_route.Route, rateLimit *anypb.Any) {
	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = map[string]*anypb.Any{}
	}
	// if there is an old RateLimit policy
	if route.TypedPerFilterConfig["envoy.filters.http.local_ratelimit"] != nil {
		return
	}
	route.TypedPerFilterConfig["envoy.filters.http.local_ratelimit"] = rateLimit
}
