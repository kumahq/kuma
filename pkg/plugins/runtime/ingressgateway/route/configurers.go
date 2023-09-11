package route

import (
	envoy_config_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_config_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/pkg/errors"

	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	envoy_virtual_hosts "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
)

// VirtualHostRoute creates an option to add the route builder to a
// virtual host. On execution, the builder will build the route and append
// it to the virtual host. Since Envoy evaluates route matches in order,
// route builders should be configured on virtual hosts in the intended
// match order.
func VirtualHostRoute(route *RouteBuilder) envoy_virtual_hosts.VirtualHostBuilderOpt {
	return envoy_virtual_hosts.AddVirtualHostConfigurer(
		envoy_virtual_hosts.VirtualHostConfigureFunc(func(vh *envoy_config_route.VirtualHost) error {
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

// RouteMatchPresentHeader appends a present match for the names HTTP request header (presentMatch makes absent)
func RouteMatchPresentHeader(name string, presentMatch bool) RouteConfigurer {
	if name == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		r.Match.Headers = append(r.Match.Headers,
			&envoy_config_route.HeaderMatcher{
				Name: name,
				HeaderMatchSpecifier: &envoy_config_route.HeaderMatcher_PresentMatch{
					PresentMatch: presentMatch,
				},
			},
		)
	})
}

func RouteActionClusterHeader(header string, tags envoy_tags.Tags) RouteConfigurer {
	if header == "" {
		return RouteConfigureFunc(nil)
	}

	return RouteMustConfigureFunc(func(r *envoy_config_route.Route) {
		rAction := &envoy_config_route.RouteAction{
			ClusterSpecifier: &envoy_config_route.RouteAction_ClusterHeader{
				ClusterHeader: header,
			},
		}
		if len(tags) != 0 {
			rAction.MetadataMatch = envoy_metadata.LbMetadata(tags)
		}

		r.Action = &envoy_config_route.Route_Route{
			Route: rAction,
		}
	})
}
