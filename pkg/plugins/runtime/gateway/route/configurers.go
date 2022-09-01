package route

import (
	"net/http"
	"strings"
	"time"

	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	v3 "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
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
		matcher := envoy_type_matcher.StringMatcher{
			MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
				Exact: value,
			},
		}
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route.HeaderMatcher_StringMatch{
					StringMatch: &matcher,
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
				Cluster: destination.Name,
				RuntimeFraction: &envoy_config_core.RuntimeFractionalPercent{
					DefaultValue: envoy_listeners.ConvertPercentage(util_proto.Double(percent)),
				},
				TraceSampled: nil,
			},
		)

		return nil
	})
}

// RoutePerFilterConfig sets an optional per-filter configuration message
// for this route. filterName is the name of the filter that should receive
// the configuration that is specified in filterConfig
func RoutePerFilterConfig(filterName string, filterConfig *anypb.Any) RouteConfigurer {
	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
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

func RouteRewrite(rewrite *Rewrite) RouteConfigurer {
	if rewrite == nil {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if r.GetAction() == nil {
			return errors.New("cannot configure rewrite before the route action")
		}

		action := r.GetRoute()

		if action == nil {
			return errors.New("cannot configure rewrite on a non-forwarding route")
		}

		if rewrite.ReplaceFullPath != nil {
			action.RegexRewrite = &envoy_type_matcher.RegexMatchAndSubstitute{
				Pattern: &envoy_type_matcher.RegexMatcher{
					EngineType: &envoy_type_matcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &envoy_type_matcher.RegexMatcher_GoogleRE2{},
					},
					Regex: `.*`,
				},
				Substitution: *rewrite.ReplaceFullPath,
			}
		}

		if rewrite.ReplacePrefixMatch != nil {
			action.PrefixRewrite = *rewrite.ReplacePrefixMatch
		}

		return nil
	})
}

// RouteActionForward configures the route to forward traffic to the
// given destinations, with the appropriate weights. This replaces any
// previous action specification.
func RouteActionForward(mesh *core_mesh.MeshResource, endpoints core_xds.EndpointMap, proxyTags mesh_proto.MultiValueTagSet, destinations []Destination) RouteConfigurer {
	if len(destinations) == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		byName := map[string]Destination{}

		for _, d := range destinations {
			// If there's only one destination, force the weight to 100%.
			if d.Weight == 0 && len(destinations) == 1 {
				d.Weight = 1
			}

			byName[d.Name] = d
		}

		var total uint32
		var weights []*envoy_config_route.WeightedCluster_ClusterWeight

		for n, d := range byName {
			total += d.Weight

			var requestHeadersToAdd []*envoy_config_core.HeaderValueOption

			isMeshCluster := mesh.ZoneEgressEnabled() || !HasExternalServiceEndpoint(mesh, endpoints, d)

			if isMeshCluster {
				requestHeadersToAdd = []*envoy_config_core.HeaderValueOption{{
					Header: &envoy_config_core.HeaderValue{Key: v3.TagsHeaderName, Value: tags.Serialize(proxyTags)},
				}}
			}

			weights = append(weights, &envoy_config_route.WeightedCluster_ClusterWeight{
				RequestHeadersToAdd: requestHeadersToAdd,
				Name:                n,
				Weight:              util_proto.UInt32(d.Weight),
			})
		}

		r.Action = &envoy_config_route.Route_Route{
			Route: &envoy_config_route.RouteAction{
				Timeout: nil, // TODO(jpeach) support request timeout from the Timeout policy, but which one?
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

// RouteActionRetryDefault initializes the retry policy with defaults appropriate for the protocol.
func RouteActionRetryDefault(protocol core_mesh.Protocol) RouteConfigurer {
	// The retry policy only supports HTTP and GRPC.
	switch protocol {
	case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
	default:
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		route := r.GetRoute()
		if route == nil {
			return nil
		}

		p := &envoy_config_route.RetryPolicy{}

		switch protocol {
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2:
			p.RetryOn = envoy_listeners.HttpRetryOnDefault
		case core_mesh.ProtocolGRPC:
			p.RetryOn = envoy_listeners.GrpcRetryOnDefault
		}

		route.RetryPolicy = p
		return nil
	})
}

// RouteActionRetryTimeout sets the per-try retry timeout.
func RouteActionRetryTimeout(perTryTimeout time.Duration) RouteConfigurer {
	if perTryTimeout == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if p := r.GetRoute().GetRetryPolicy(); p != nil {
			p.PerTryTimeout = util_proto.Duration(perTryTimeout)
		}

		return nil
	})
}

// RouteActionRetryCount sets the number of retries to attempt.
func RouteActionRetryCount(numRetries uint32) RouteConfigurer {
	if numRetries == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if p := r.GetRoute().GetRetryPolicy(); p != nil {
			p.NumRetries = util_proto.UInt32(numRetries)
		}

		return nil
	})
}

// RouteActionRetryBackoff sets the backoff policy for retries.
func RouteActionRetryBackoff(interval time.Duration, max time.Duration) RouteConfigurer {
	// the base interval is required, but Envoy will default the max.
	if interval == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if p := r.GetRoute().GetRetryPolicy(); p != nil {
			p.RetryBackOff = &envoy_config_route.RetryPolicy_RetryBackOff{
				BaseInterval: util_proto.Duration(interval),
			}

			if max > 0 {
				p.RetryBackOff.MaxInterval = util_proto.Duration(max)
			}
		}

		return nil
	})
}

// RouteActionRetryMethods sets the HTTP methods that should trigger retries.
func RouteActionRetryMethods(httpMethod ...string) RouteConfigurer {
	if len(httpMethod) == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		p := r.GetRoute().GetRetryPolicy()
		if p != nil {
			return nil
		}

		for _, m := range httpMethod {
			matcher := envoy_type_matcher.StringMatcher{
				MatchPattern: &envoy_type_matcher.StringMatcher_Exact{
					Exact: m,
				},
			}
			p.RetriableRequestHeaders = append(p.RetriableRequestHeaders,
				&envoy_config_route.HeaderMatcher{
					Name: ":method",
					HeaderMatchSpecifier: &envoy_config_route.HeaderMatcher_StringMatch{
						StringMatch: &matcher,
					},
				})
		}

		return nil
	})
}

// RouteActionRetryOnStatus sets the HTTP status codes for triggering retries.
func RouteActionRetryOnStatus(httpStatus ...uint32) RouteConfigurer {
	if len(httpStatus) == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if p := r.GetRoute().GetRetryPolicy(); p != nil {
			p.RetryOn = envoy_listeners.HttpRetryOnRetriableStatusCodes
			p.RetriableStatusCodes = make([]uint32, len(httpStatus))
			copy(p.RetriableStatusCodes, httpStatus)
		}

		return nil
	})
}

// RouteActionRetryOnConditions sets the Envoy condition names for triggering retries.
func RouteActionRetryOnConditions(conditionNames ...string) RouteConfigurer {
	if len(conditionNames) == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if p := r.GetRoute().GetRetryPolicy(); p != nil {
			var conditions []string
			for _, c := range conditionNames {
				conditions = append(conditions, strings.ReplaceAll(c, "_", "-"))
			}
			p.RetryOn = strings.Join(conditions, ",")
		}

		return nil
	})
}

// RouteActionRequestTimeout sets the total timeout for an upstream request.
func RouteActionRequestTimeout(timeout time.Duration) RouteConfigurer {
	if timeout == 0 {
		return RouteConfigureFunc(nil)
	}

	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		if p := r.GetRoute(); p != nil {
			p.Timeout = util_proto.Duration(timeout)
		}

		return nil
	})
}

// RouteActionDirectResponse sets the direct response for a route
func RouteActionDirectResponse(status uint32, respStr string) RouteConfigurer {
	return RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		r.Action = &envoy_config_route.Route_DirectResponse{
			DirectResponse: &envoy_config_route.DirectResponseAction{
				Status: status,
				Body: &envoy_config_core.DataSource{
					Specifier: &envoy_config_core.DataSource_InlineString{
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
