package route

import (
	"net/http"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/pkg/errors"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

func regexOf(regex string) *envoy_type_matcher.RegexMatcher {
	return &envoy_type_matcher.RegexMatcher{
		Regex: regex,
		EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
			GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
		},
	}
}

// RouteMatchExactPath updates the route to match the exact path. This
// replaces any previous path match specification.
func RouteMatchExactPath(path string) RouteConfigurer {
	if path == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.Match.PathSpecifier = &envoy_config_route.RouteMatch_Path{
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

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.Match.PathSpecifier = &envoy_config_route.RouteMatch_Prefix{
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

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.Match.PathSpecifier = &envoy_config_route.RouteMatch_SafeRegex{
			SafeRegex: regexOf(regex),
		}
	})
}

// RouteMatchExactHeader appends an exact match for the value of the named HTTP request header.
func RouteMatchExactHeader(name string, value string) RouteConfigurer {
	if name == "" || value == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route.HeaderMatcher_ExactMatch{
					ExactMatch: value,
				},
			},
		)
	})
}

// RouteMatchRegexHeader appends a regex match for the value of the named HTTP request header.
func RouteMatchRegexHeader(name string, regex string) RouteConfigurer {
	if name == "" || regex == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route.HeaderMatcher_SafeRegexMatch{
					SafeRegexMatch: regexOf(regex),
				},
			},
		)
	})
}

// RouteMatchExactQuery appends an exact match for the value of the named query parameter.
func RouteMatchExactQuery(name string, value string) RouteConfigurer {
	if name == "" || value == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: value,
			},
		}

		r.Match.QueryParameters = append(r.Match.QueryParameters,
			&envoy_config_route.QueryParameterMatcher{
				Name: name,
				QueryParameterMatchSpecifier: &envoy_config_route.QueryParameterMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		)
	})
}

// RouteMatchRegexQuery appends a regex match for the value of the named query parameter.
func RouteMatchRegexQuery(name string, regex string) RouteConfigurer {
	if name == "" || regex == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_SafeRegex{
				SafeRegex: regexOf(regex),
			},
		}

		r.Match.QueryParameters = append(r.Match.QueryParameters,
			&envoy_config_route.QueryParameterMatcher{
				Name: name,
				QueryParameterMatchSpecifier: &envoy_config_route.QueryParameterMatcher_StringMatch{
					StringMatch: &matcher,
				},
			},
		)
	})
}

// RouteAppendRequestHeader appends the given value to the existing values of the given header.
func RouteAppendRequestHeader(name string, value string) RouteConfigurer {
	if name == "" || value == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.RequestHeadersToAdd = append(r.RequestHeadersToAdd,
			&envoy_config_core.HeaderValueOption{
				Append: util_proto.Bool(true),
				Header: &envoy_config_core.HeaderValue{
					Key:   http.CanonicalHeaderKey(name),
					Value: value,
				},
			},
		)
	})
}

// RouteReplaceRequestHeader replaces all values of the given header with the given value.
func RouteReplaceRequestHeader(name string, value string) RouteConfigurer {
	if name == "" || value == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.RequestHeadersToAdd = append(r.RequestHeadersToAdd,
			&envoy_config_core.HeaderValueOption{
				Append: util_proto.Bool(false),
				Header: &envoy_config_core.HeaderValue{
					Key:   http.CanonicalHeaderKey(name),
					Value: value,
				},
			},
		)
	})
}

// RouteReplaceHostHeader replaces the Host header on the forwarded
// request. It is an error to rewrite the header if the route is not
// forwarding. The route action must be configured beforehand.
func RouteReplaceHostHeader(host string) RouteConfigurer {
	if host == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if r.GetAction() == nil {
			return errors.New("cannot configure the Host header before the route action")
		}

		if action := r.GetRoute(); action != nil {
			action.HostRewriteSpecifier = &envoy_config_route.RouteAction_HostRewriteLiteral{
				HostRewriteLiteral: host,
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

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.RequestHeadersToRemove = append(r.RequestHeadersToRemove, name)
	})
}

// RouteMirror enables traffic mirroring on the route. It is an error to enable
// mirroring if the route is not forwarding. The route action must be configured
// beforehand.
func RouteMirror(percent float64, destination Destination) RouteConfigurer {
	if percent <= 0.0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		clusterName, err := DestinationClusterName(destination)
		if err != nil {
			return errors.Wrap(err, "failed to generate mirror cluster name")
		}

		if r.GetAction() == nil {
			return errors.New("cannot configure mirroring before the route action")
		}

		action := r.GetRoute()

		// If we aren't forwarding on this route, we can't mirror.
		if action == nil {
			return errors.New("cannot configure mirroring on a non-forwarding route")
		}

		action.RequestMirrorPolicies = append(action.RequestMirrorPolicies,
			&envoy_config_route.RouteAction_RequestMirrorPolicy{
				Cluster: clusterName,
				RuntimeFraction: &envoy_config_core.RuntimeFractionalPercent{
					DefaultValue: envoy_listeners.ConvertPercentage(util_proto.Double(percent)),
				},
				TraceSampled: nil,
			},
		)

		return nil
	})
}

// RouteActionRedirect configures the route to automatically response
// with a HTTP redirection. This replaces any previous action specification.
func RouteActionRedirect(redirect *Redirection) RouteConfigurer {
	if redirect == nil {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		r.Action = &envoy_config_route.Route_Redirect{
			Redirect: &envoy_config_route.RedirectAction{
				SchemeRewriteSpecifier: &envoy_config_route.RedirectAction_SchemeRedirect{
					SchemeRedirect: redirect.Scheme,
				},
				HostRedirect: redirect.Host,
				PortRedirect: redirect.Port,
				StripQuery:   redirect.StripQuery,
			},
		}

		switch redirect.Status {
		case 301:
			r.GetRedirect().ResponseCode = envoy_config_route.RedirectAction_MOVED_PERMANENTLY
		case 302:
			r.GetRedirect().ResponseCode = envoy_config_route.RedirectAction_FOUND
		case 303:
			r.GetRedirect().ResponseCode = envoy_config_route.RedirectAction_SEE_OTHER
		case 307:
			r.GetRedirect().ResponseCode = envoy_config_route.RedirectAction_TEMPORARY_REDIRECT
		case 308:
			r.GetRedirect().ResponseCode = envoy_config_route.RedirectAction_PERMANENT_REDIRECT
		default:
			return errors.Errorf("redirect status code %d is not supported", redirect.Status)
		}

		return nil
	})
}

// RouteActionForward configures the route to forward traffic to the
// given destinations, with the appropriate weights. This replaces any
// previous action specification.
func RouteActionForward(destinations []Destination) RouteConfigurer {
	if len(destinations) == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		byName := map[string]Destination{}

		for _, d := range destinations {
			name, err := DestinationClusterName(d)
			if err != nil {
				return errors.Wrap(err, "failed to generate forwarding cluster name")
			}

			byName[name] = d

			// If there's only one destination, force the weight to 100%.
			if d.Weight == 0 && len(destinations) == 1 {
				d := byName[name]
				d.Weight = 1
				byName[name] = d
			}
		}

		var total uint32
		var weights []*envoy_config_route.WeightedCluster_ClusterWeight

		for n, d := range byName {
			total += d.Weight
			weights = append(weights, &envoy_config_route.WeightedCluster_ClusterWeight{
				Name:   n,
				Weight: util_proto.UInt32(d.Weight),
			})
		}

		r.Action = &envoy_config_route.Route_Route{
			Route: &envoy_config_route.RouteAction{
				ClusterSpecifier: &envoy_config_route.RouteAction_WeightedClusters{
					WeightedClusters: &envoy_config_route.WeightedCluster{
						Clusters:    weights,
						TotalWeight: util_proto.UInt32(total),
					}},
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
func VirtualHostRoute(route *RouteBuilder) envoy_routes.VirtualHostBuilderOpt {
	return envoy_routes.AddVirtualHostConfigurer(
		v3.VirtualHostConfigureFunc(func(vh *envoy_config_route.VirtualHost) error {
			resource, err := route.Build()
			if err != nil {
				return err
			}

			routeProto, ok := resource.(*envoy_config_route.Route)
			if !ok {
				return errors.Errorf("attempt to attach %T as type %q",
					resource, "envoy_config_route.Route")
			}

			vh.Routes = append(vh.Routes, routeProto)
			return nil
		}),
	)
}
