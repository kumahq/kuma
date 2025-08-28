package routes

import (
	"net/http"
	"time"

	envoy_config_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher_v3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"

	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

// RouteMatchExactPath updates the route to match the exact path. This
// replaces any previous path match specification.
func RouteMatchExactPath(path string) RouteConfigurer {
	if path == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.Match.PathSpecifier = &envoy_config_route_v3.RouteMatch_Path{
			Path: path,
		}
	})
}

// RouteMatchPrefixPath updates the route to match the given path
// prefix. This is a byte-wise prefix, so it just checks that the request
// path begins with the given string. This replaces any previous path match
// specification.
func RouteMatchPrefixPath(prefix string) RouteConfigurer {
	if prefix == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.Match.PathSpecifier = &envoy_config_route_v3.RouteMatch_Prefix{
			Prefix: prefix,
		}
	})
}

// RouteMatchRegexPath updates the route to match the path using the
// given regex. This replaces any previous path match specification.
func RouteMatchRegexPath(regex string) RouteConfigurer {
	if regex == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.Match.PathSpecifier = &envoy_config_route_v3.RouteMatch_SafeRegex{
			SafeRegex: &envoy_type_matcher_v3.RegexMatcher{Regex: regex},
		}
	})
}

// RouteMatchExactHeader appends an exact match for the value of the named HTTP request header.
func RouteMatchExactHeader(name, value string) RouteConfigurer {
	if name == "" || value == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		matcher := envoy_type_matcher_v3.StringMatcher{
			MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{
				Exact: value,
			},
		}
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route_v3.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route_v3.HeaderMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		)
	})
}

// RouteMatchRegexHeader appends a regex match for the value of the named HTTP request header.
func RouteMatchRegexHeader(name, regex string) RouteConfigurer {
	if name == "" || regex == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route_v3.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route_v3.HeaderMatcher_StringMatch{
					StringMatch: &envoy_type_matcher_v3.StringMatcher{
						MatchPattern: &envoy_type_matcher_v3.StringMatcher_SafeRegex{
							SafeRegex: &envoy_type_matcher_v3.RegexMatcher{
								Regex: regex,
							},
						},
					},
				},
			},
		)
	})
}

// RouteMatchPresentHeader appends a present match for the names HTTP request header (presentMatch makes absent)
func RouteMatchPresentHeader(name string, presentMatch bool) RouteConfigurer {
	if name == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route_v3.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route_v3.HeaderMatcher_PresentMatch{
					PresentMatch: presentMatch,
				},
			},
		)
	})
}

// RouteMatchPrefixHeader appends a prefix match for the names HTTP request header
func RouteMatchPrefixHeader(name, match string) RouteConfigurer {
	if name == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route_v3.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route_v3.HeaderMatcher_PrefixMatch{
					PrefixMatch: match,
				},
			},
		)
	})
}

// RouteMatchExactQuery appends an exact match for the value of the named query parameter.
func RouteMatchExactQuery(name, value string) RouteConfigurer {
	if name == "" || value == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		matcher := envoy_type_matcher_v3.StringMatcher{
			MatchPattern: &envoy_type_matcher_v3.StringMatcher_Exact{
				Exact: value,
			},
		}

		r.Match.QueryParameters = append(r.Match.QueryParameters,
			&envoy_config_route_v3.QueryParameterMatcher{
				Name: name,
				QueryParameterMatchSpecifier: &envoy_config_route_v3.QueryParameterMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		)
	})
}

// RouteMatchRegexQuery appends a regex match for the value of the named query parameter.
func RouteMatchRegexQuery(name, regex string) RouteConfigurer {
	if name == "" || regex == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		matcher := envoy_type_matcher_v3.StringMatcher{
			MatchPattern: &envoy_type_matcher_v3.StringMatcher_SafeRegex{
				SafeRegex: &envoy_type_matcher_v3.RegexMatcher{Regex: regex},
			},
		}

		r.Match.QueryParameters = append(r.Match.QueryParameters,
			&envoy_config_route_v3.QueryParameterMatcher{
				Name: name,
				QueryParameterMatchSpecifier: &envoy_config_route_v3.QueryParameterMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		)
	})
}

func RouteAppendHeader(name, value string) *envoy_config_core_v3.HeaderValueOption {
	return &envoy_config_core_v3.HeaderValueOption{
		AppendAction: envoy_config_core_v3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
		Header: &envoy_config_core_v3.HeaderValue{
			Key:   http.CanonicalHeaderKey(name),
			Value: value,
		},
	}
}

func RouteReplaceHeader(name, value string) *envoy_config_core_v3.HeaderValueOption {
	return &envoy_config_core_v3.HeaderValueOption{
		AppendAction: envoy_config_core_v3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
		Header: &envoy_config_core_v3.HeaderValue{
			Key:   http.CanonicalHeaderKey(name),
			Value: value,
		},
	}
}

// RouteAddRequestHeader alters the given request header value.
func RouteAddRequestHeader(option *envoy_config_core_v3.HeaderValueOption) RouteConfigurer {
	if option == nil {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.RequestHeadersToAdd = append(r.RequestHeadersToAdd, option)
	})
}

// RouteAddResponseHeader alters the given response header value.
func RouteAddResponseHeader(option *envoy_config_core_v3.HeaderValueOption) RouteConfigurer {
	if option == nil {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.ResponseHeadersToAdd = append(r.ResponseHeadersToAdd, option)
	})
}

// RouteReplaceHostHeader replaces the Host header on the forwarded
// request. It is an error to rewrite the header if the route is not
// forwarding. The route action must be configured beforehand.
func RouteReplaceHostHeader(host string) RouteConfigurer {
	if host == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		if r.GetAction() == nil {
			return errors.New("cannot configure the Host header before the route action")
		}

		if action := r.GetRoute(); action != nil {
			action.HostRewriteSpecifier = &envoy_config_route_v3.RouteAction_HostRewriteLiteral{
				HostRewriteLiteral: host,
			}
		}

		return nil
	})
}

func RouteSetRewriteHostToBackendHostname(value bool) RouteConfigurer {
	if !value {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		if r.GetAction() == nil {
			return errors.New("cannot set the 'auto_host_rewrite' before the route action")
		}

		if action := r.GetRoute(); action != nil {
			action.HostRewriteSpecifier = &envoy_config_route_v3.RouteAction_AutoHostRewrite{
				AutoHostRewrite: util_proto.Bool(value),
			}
		}

		return nil
	})
}

// RouteDeleteRequestHeader deletes the given header from the HTTP request.
func RouteDeleteRequestHeader(name string) RouteConfigurer {
	if name == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.RequestHeadersToRemove = append(r.RequestHeadersToRemove, name)
	})
}

// RouteDeleteResponseHeader deletes the given header from the HTTP response.
func RouteDeleteResponseHeader(name string) RouteConfigurer {
	if name == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		r.ResponseHeadersToRemove = append(r.ResponseHeadersToRemove, name)
	})
}

// RoutePerFilterConfig sets an optional per-filter configuration message
// for this route. filterName is the name of the filter that should receive
// the configuration that is specified in filterConfig
func RoutePerFilterConfig(filterName string, filterConfig *anypb.Any) RouteConfigurer {
	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		if r.GetTypedPerFilterConfig() == nil {
			r.TypedPerFilterConfig = map[string]*anypb.Any{}
		}

		m := r.GetTypedPerFilterConfig()

		if _, ok := m[filterName]; ok {
			return errors.Errorf("duplicate %q per-filter config for %s",
				filterConfig.GetTypeUrl(), filterName)
		}

		m[filterName] = filterConfig
		return nil
	})
}

// RouteActionRetryDefault initializes the retry policy with defaults appropriate for the protocol.
func RouteActionRetryDefault(protocol core_meta.Protocol) RouteConfigurer {
	// The retry policy only supports HTTP and GRPC.
	switch protocol {
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
	default:
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		route := r.GetRoute()
		if route == nil {
			return nil
		}

		p := &envoy_config_route_v3.RetryPolicy{}

		switch protocol {
		case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2:
			p.RetryOn = envoy_routes.HttpRetryOnDefault
		case core_meta.ProtocolGRPC:
			p.RetryOn = envoy_routes.GrpcRetryOnDefault
		}

		route.RetryPolicy = p
		return nil
	})
}

func RouteActionRetry(retry *core_mesh.RetryResource, protocol core_meta.Protocol) RouteConfigurer {
	// The retry policy only supports HTTP and GRPC.
	switch protocol {
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
	default:
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		route := r.GetRoute()
		route.RetryPolicy = envoy_routes.RetryConfig(retry, protocol)
		return nil
	})
}

// RouteActionRequestTimeout sets the total timeout for an upstream request.
func RouteActionRequestTimeout(timeout time.Duration) RouteConfigurer {
	if timeout == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		if p := r.GetRoute(); p != nil {
			p.Timeout = util_proto.Duration(timeout)
		}

		return nil
	})
}

func RouteActionIdleTimeout(timeout time.Duration) RouteConfigurer {
	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		if p := r.GetRoute(); p != nil {
			p.IdleTimeout = util_proto.Duration(timeout)
		}

		return nil
	})
}

// RouteActionDirectResponse sets the direct response for a route
func RouteActionDirectResponse(status uint32, respStr string) RouteConfigurer {
	return RouteConfigureFunc(func(r *envoy_config_route_v3.Route) error {
		r.Action = &envoy_config_route_v3.Route_DirectResponse{
			DirectResponse: &envoy_config_route_v3.DirectResponseAction{
				Status: status,
				Body: &envoy_config_core_v3.DataSource{
					Specifier: &envoy_config_core_v3.DataSource_InlineString{
						InlineString: respStr,
					},
				},
			},
		}
		return nil
	})
}

// VirtualHostRoute creates an option to add the route builder to a
// virtual host. On execution, the builder will build the route and append
// it to the virtual host. Since Envoy evaluates route matches in order,
// route builders should be configured on virtual hosts in the intended
// match order.
func VirtualHostRoute(route *RouteBuilder) envoy_virtual_hosts.VirtualHostBuilderOpt {
	return envoy_virtual_hosts.AddVirtualHostConfigurer(
		envoy_virtual_hosts.VirtualHostConfigureFunc(func(vh *envoy_config_route_v3.VirtualHost) error {
			resource, err := route.Build()
			if err != nil {
				return err
			}

			routeProto, ok := resource.(*envoy_config_route_v3.Route)
			if !ok {
				return errors.Errorf("attempt to attach %T as type %q",
					resource, "envoy_config_route_v3.Route")
			}

			vh.Routes = append(vh.Routes, routeProto)
			return nil
		}),
	)
}
