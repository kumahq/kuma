package xds

import (
	"strings"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_extensions_filters_network_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/local_ratelimit/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"

	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/kumahq/kuma/pkg/util/proto"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	rate_limit "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

func RateLimitConfigurationFromPolicy(rl *api.LocalHTTP) *rate_limit.RateLimitConfiguration {
	if pointer.Deref(rl.Disabled) || rl.RequestRate == nil {
		return nil
	}

	onRateLimit := &rate_limit.OnRateLimit{}
	if rl.OnRateLimit != nil {
		for _, h := range pointer.Deref(rl.OnRateLimit.Headers).Add {
			onRateLimit.Headers = append(onRateLimit.Headers, &rate_limit.Headers{
				Key:    string(h.Name),
				Value:  string(h.Value),
				Append: true,
			})
		}
		for _, header := range pointer.Deref(rl.OnRateLimit.Headers).Set {
			for _, val := range strings.Split(string(header.Value), ",") {
				onRateLimit.Headers = append(onRateLimit.Headers, &rate_limit.Headers{
					Key:    string(header.Name),
					Value:  val,
					Append: false,
				})
			}
		}
		onRateLimit.Status = pointer.Deref(rl.OnRateLimit.Status)
	}

	return &rate_limit.RateLimitConfiguration{
		Interval:    rl.RequestRate.Interval.Duration,
		Requests:    rl.RequestRate.Num,
		OnRateLimit: onRateLimit,
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

	rlConf := RateLimitConfigurationFromPolicy(c.Http)
	if rlConf == nil {
		return nil
	}

	rateLimit, err := rate_limit.NewRateLimitConfiguration(rlConf)
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
	rlConf := RateLimitConfigurationFromPolicy(c.Http)
	if rlConf == nil {
		// MeshRateLimit policy is matched for the DPP, but rate limit either disabled
		// or not configured. Potentially we can return errors that bubble up to GUI from here.
		return nil
	}

	rateLimit, err := rate_limit.NewRateLimitConfiguration(rlConf)
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
		return policies_xds.InsertHTTPFiltersBeforeRouter(hcm, &envoy_hcm.HttpFilter{
			Name: "envoy.filters.http.local_ratelimit",
			ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: pbstListener,
			},
		})
	}
	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpRoutes); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	return nil
}

func (c *Configurer) configureTcpListener(filterChain *envoy_listener.FilterChain) error {
	if pointer.Deref(c.Tcp.Disabled) || c.Tcp.ConnectionRate == nil {
		// MeshRateLimit policy is matched for the DPP, but rate limit either disabled
		// or not configured. Potentially we can return errors that bubble up to GUI from here.
		return nil
	}

	config := &envoy_extensions_filters_network_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "tcp_rate_limit",
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens:     c.Tcp.ConnectionRate.Num,
			TokensPerFill: proto.UInt32(c.Tcp.ConnectionRate.Num),
			FillInterval:  proto.Duration(c.Tcp.ConnectionRate.Interval.Duration),
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
