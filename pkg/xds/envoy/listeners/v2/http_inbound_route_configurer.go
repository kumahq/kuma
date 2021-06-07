package v2

import (
	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	envoy_config_filter_network_local_rate_limit_v2alpha "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/local_rate_limit/v2alpha"
	envoy_type "github.com/envoyproxy/go-control-plane/envoy/type"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	v2 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v2"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"

	"github.com/kumahq/kuma/pkg/util/proto"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

type HttpInboundRouteConfigurer struct {
	Service   string
	Route     envoy_common.Route
	RateLimit *mesh_proto.RateLimit
}

func (c *HttpInboundRouteConfigurer) Configure(filterChain *envoy_listener.FilterChain) error {
	routeName := envoy_names.GetInboundRouteName(c.Service)
	routes, err := c.createRoutes()
	if err != nil {
		return err
	}

	routeConfig, err := envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV2).
		Configure(envoy_routes.CommonRouteConfiguration(routeName)).
		Configure(envoy_routes.ResetTagsHeader()).
		Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder(envoy_common.APIV2).
			Configure(envoy_routes.CommonVirtualHost(c.Service)).
			Configure(envoy_routes.Routes(routes)))).
		Build()
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig.(*envoy_api.RouteConfiguration),
		}
		return nil
	})
}

func (c *HttpInboundRouteConfigurer) createRoutes() (envoy_common.Routes, error) {
	route := c.Route

	if c.RateLimit != nil && c.RateLimit.GetConf().GetHttp() != nil {
		// Source
		if len(c.RateLimit.GetSources()) > 0 {
			if route.Match == nil {
				route.Match = &mesh_proto.TrafficRoute_Http_Match{}
			}

			if route.Match.Headers == nil {
				route.Match.Headers = make(map[string]*mesh_proto.TrafficRoute_Http_Match_StringMatcher)
			}

			var selectorRegexs []string
			for _, selector := range c.RateLimit.SourceTags() {
				selectorRegexs = append(selectorRegexs, tags.MatchingRegex(selector))
			}
			regexOR := tags.RegexOR(selectorRegexs...)

			route.Match.Headers[v2.TagsHeaderName] = &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
				MatcherType: &mesh_proto.TrafficRoute_Http_Match_StringMatcher_Regex{
					Regex: regexOR,
				},
			}
		}

		pbst, err := c.createRateLimit()
		if err != nil {
			return nil, err
		}

		if route.TypedPerFilterConfig == nil {
			route.TypedPerFilterConfig = map[string]*any.Any{}
		}

		route.TypedPerFilterConfig["envoy.filters.http.local_ratelimit"] = pbst
	}

	return envoy_common.Routes{route}, nil
}

func (c *HttpInboundRouteConfigurer) createRateLimit() (*any.Any, error) {
	rlHttp := c.RateLimit.GetConf().GetHttp()

	if rlHttp.GetOnRateLimit() != nil {
		return nil, errors.Errorf("APIv2 does not support the `onError` field")
	}

	config := &envoy_config_filter_network_local_rate_limit_v2alpha.LocalRateLimit{
		StatPrefix: "rate_limit",
		TokenBucket: &envoy_type.TokenBucket{
			MaxTokens: rlHttp.Requests.GetValue(),
			TokensPerFill: &wrappers.UInt32Value{
				Value: 1,
			},
			FillInterval: rlHttp.GetInterval(),
		},
	}

	return proto.MarshalAnyDeterministic(config)
}
