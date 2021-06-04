package v3

import (
	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_type_v3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"

	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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

	routeConfig, err := envoy_routes.NewRouteConfigurationBuilder(envoy_common.APIV3).
		Configure(envoy_routes.CommonRouteConfiguration(routeName)).
		Configure(envoy_routes.ResetTagsHeader()).
		Configure(envoy_routes.VirtualHost(envoy_routes.NewVirtualHostBuilder(envoy_common.APIV3).
			Configure(envoy_routes.CommonVirtualHost(c.Service)).
			Configure(envoy_routes.Routes(routes)))).
		Build()
	if err != nil {
		return err
	}

	return UpdateHTTPConnectionManager(filterChain, func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.RouteSpecifier = &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: routeConfig.(*envoy_route.RouteConfiguration),
		}

		// Enable the Local Rate Limit at the HttpConnectionManager level
		if c.hasHttpRateLimit() {
			config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
				StatPrefix: "rate_limit",
			}

			pbst, err := proto.MarshalAnyDeterministic(config)
			if err != nil {
				return err
			}

			hcm.HttpFilters = append([]*envoy_hcm.HttpFilter{
				{
					Name: "envoy.filters.http.local_ratelimit",
					ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
						TypedConfig: pbst,
					},
				},
			}, hcm.HttpFilters...)
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

			route.Match.Headers[v3.TagsHeaderName] = &mesh_proto.TrafficRoute_Http_Match_StringMatcher{
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

	var status *envoy_type_v3.HttpStatus
	var responseHeaders []*envoy_config_core_v3.HeaderValueOption
	if rlHttp.GetOnError() != nil {
		status = &envoy_type_v3.HttpStatus{
			Code: envoy_type_v3.StatusCode(rlHttp.GetOnError().GetStatus().GetValue()),
		}
		responseHeaders = []*envoy_config_core_v3.HeaderValueOption{}
		for _, h := range rlHttp.GetOnError().GetHeaders() {
			responseHeaders = append(responseHeaders, &envoy_config_core_v3.HeaderValueOption{
				Header: &envoy_config_core_v3.HeaderValue{
					Key:   h.GetKey(),
					Value: h.GetValue(),
				},
				Append: h.GetAppend(),
			})
		}
	}

	config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
		Status:     status,
		TokenBucket: &envoy_type_v3.TokenBucket{
			MaxTokens: rlHttp.Connections.GetValue(),
			TokensPerFill: &wrappers.UInt32Value{
				Value: 1,
			},
			FillInterval: rlHttp.GetInterval(),
		},
		FilterEnabled: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enabled",
		},
		FilterEnforced: &envoy_config_core_v3.RuntimeFractionalPercent{
			DefaultValue: &envoy_type_v3.FractionalPercent{
				Numerator:   100,
				Denominator: envoy_type_v3.FractionalPercent_HUNDRED,
			},
			RuntimeKey: "local_rate_limit_enforced",
		},
		ResponseHeadersToAdd: responseHeaders,
	}

	return proto.MarshalAnyDeterministic(config)
}

func (c *HttpInboundRouteConfigurer) hasHttpRateLimit() bool {
	return c.RateLimit != nil && c.RateLimit.GetConf().GetHttp() != nil
}
