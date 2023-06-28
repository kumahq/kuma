package v3

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
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

// RouteMatchExactHeader appends an exact match for the value of the named HTTP request header.
func RouteMatchExactHeader(name string, value string) RouteConfigurer {
	if name == "" || value == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route_v3.Route) {
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
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
