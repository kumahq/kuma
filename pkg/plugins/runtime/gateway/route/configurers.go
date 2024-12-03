package route

import (
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_type_matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

// RouteMirror enables traffic mirroring on the route. It is an error to enable
// mirroring if the route is not forwarding. The route action must be configured
// beforehand.
func RouteMirror(percent float64, destination Destination) envoy_routes.RouteConfigurer {
	if percent <= 0.0 {
		return envoy_routes.RouteConfigureFunc(nil)
	}

	return envoy_routes.RouteConfigureFunc(func(r *envoy_config_route.Route) error {
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

// RouteActionRedirect configures the route to automatically response
// with an HTTP redirection. This replaces any previous action specification.
func RouteActionRedirect(redirect *Redirection, port uint32) envoy_routes.RouteConfigurer {
	if redirect == nil {
		return envoy_routes.RouteConfigureFunc(nil)
	}

	return envoy_routes.RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		envoyRedirect := &envoy_config_route.RedirectAction{
			StripQuery: redirect.StripQuery,
		}
		if redirect.Scheme != "" {
			envoyRedirect.SchemeRewriteSpecifier = &envoy_config_route.RedirectAction_SchemeRedirect{
				SchemeRedirect: redirect.Scheme,
			}
		}
		if redirect.Host != "" {
			envoyRedirect.HostRedirect = redirect.Host
		}
		if redirect.Port > 0 {
			envoyRedirect.PortRedirect = redirect.Port
		} else {
			envoyRedirect.PortRedirect = port
		}
		if rewrite := redirect.PathRewrite; rewrite != nil {
			if rewrite.ReplaceFullPath != nil {
				envoyRedirect.PathRewriteSpecifier = &envoy_config_route.RedirectAction_RegexRewrite{
					RegexRewrite: &envoy_type_matcher.RegexMatchAndSubstitute{
						Pattern: &envoy_type_matcher.RegexMatcher{
							Regex: `.*`,
						},
						Substitution: *rewrite.ReplaceFullPath,
					},
				}
			}

			if rewrite.ReplacePrefixMatch != nil {
				envoyRedirect.PathRewriteSpecifier = &envoy_config_route.RedirectAction_PrefixRewrite{
					PrefixRewrite: *rewrite.ReplacePrefixMatch,
				}
			}
		}

		switch redirect.Status {
		case 301:
			envoyRedirect.ResponseCode = envoy_config_route.RedirectAction_MOVED_PERMANENTLY
		case 302:
			envoyRedirect.ResponseCode = envoy_config_route.RedirectAction_FOUND
		case 303:
			envoyRedirect.ResponseCode = envoy_config_route.RedirectAction_SEE_OTHER
		case 307:
			envoyRedirect.ResponseCode = envoy_config_route.RedirectAction_TEMPORARY_REDIRECT
		case 308:
			envoyRedirect.ResponseCode = envoy_config_route.RedirectAction_PERMANENT_REDIRECT
		default:
			return errors.Errorf("redirect status code %d is not supported", redirect.Status)
		}

		r.Action = &envoy_config_route.Route_Redirect{
			Redirect: envoyRedirect,
		}

		return nil
	})
}

func RouteRewrite(rewrite *Rewrite) envoy_routes.RouteConfigurer {
	if rewrite == nil {
		return envoy_routes.RouteConfigureFunc(nil)
	}

	return envoy_routes.RouteConfigureFunc(func(r *envoy_config_route.Route) error {
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
func RouteActionForward(xdsCtx xds_context.Context, endpoints core_xds.EndpointMap, proxyTags mesh_proto.MultiValueTagSet, destinations []Destination) envoy_routes.RouteConfigurer {
	if len(destinations) == 0 {
		return envoy_routes.RouteConfigureFunc(nil)
	}

	return envoy_routes.RouteConfigureFunc(func(r *envoy_config_route.Route) error {
		byName := map[string]Destination{}

		for _, d := range destinations {
			// If there's only one destination, force the weight to 100%.
			if d.Weight == 0 && len(destinations) == 1 {
				d.Weight = 1
			}

			byName[d.Name] = d
		}

		var weights []*envoy_config_route.WeightedCluster_ClusterWeight

		names := util_maps.SortedKeys(byName)
		for _, name := range names {
			destination := byName[name]
			var requestHeadersToAdd []*envoy_config_core.HeaderValueOption

			isMeshCluster := xdsCtx.Mesh.Resource.ZoneEgressEnabled() || !xdsCtx.Mesh.IsExternalService(destination.Destination[mesh_proto.ServiceTag])
			if len(xdsCtx.Mesh.Resources.TrafficPermissions().Items) > 0 {
				isMeshCluster = xdsCtx.Mesh.Resource.ZoneEgressEnabled() || !HasExternalServiceEndpoint(xdsCtx.Mesh.Resource, endpoints, destination)
			}

			if isMeshCluster {
				requestHeadersToAdd = []*envoy_config_core.HeaderValueOption{{
					Header: &envoy_config_core.HeaderValue{Key: tags.TagsHeaderName, Value: tags.Serialize(proxyTags)},
				}}
			}

			weights = append(weights, &envoy_config_route.WeightedCluster_ClusterWeight{
				RequestHeadersToAdd: requestHeadersToAdd,
				Name:                name,
				Weight:              util_proto.UInt32(destination.Weight),
			})
		}

		r.Action = &envoy_config_route.Route_Route{
			Route: &envoy_config_route.RouteAction{
				ClusterNotFoundResponseCode: envoy_config_route.RouteAction_INTERNAL_SERVER_ERROR,
				Timeout:                     nil, // TODO(jpeach) support request timeout from the Timeout policy, but which one?
				ClusterSpecifier: &envoy_config_route.RouteAction_WeightedClusters{
					WeightedClusters: &envoy_config_route.WeightedCluster{
						Clusters: weights,
					},
				},
			},
		}

		return nil
	})
}
