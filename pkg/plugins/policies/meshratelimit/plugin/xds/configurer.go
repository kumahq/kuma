package xds

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_extensions_filters_http_local_ratelimit_v3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/xds"
	"github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
)

type Configurer struct {
	From        core_xds.Rules
	ClusterName string
	Dataplane   *core_mesh.DataplaneResource
}

func (c *Configurer) Configure(filterChain *envoy_listener.FilterChain) error {
	if err := c.configureRoutes(filterChain); err != nil {
		return err
	}
	// route
	if err := c.configureHttpListener(filterChain); err != nil {
		return err
	}
	return nil
}

func (c *Configurer) configureRoutes(filterChain *envoy_listener.FilterChain) error {
	httpRoutes := func(hcm *envoy_hcm.HttpConnectionManager) error {
		cluster := envoy_common.NewCluster(envoy_common.WithService(c.ClusterName))
		for _, vh := range hcm.GetRouteConfig().GetVirtualHosts() {
			existingRoutes := policies_xds.GatherRoutes(vh, true)
			routes := envoy_common.Routes{}
			// that means we have only one default route
			if len(existingRoutes) == 1 {
				for _, rule := range c.From {
					routes = append(routes, envoy_common.NewRoute(
						envoy_common.WithCluster(cluster),
						envoy_common.WithMatchHeaderRegex(envoy_routes.TagsHeaderName, tags.MatchRuleRegex(rule.Subset)),
						envoy_common.WithMeshRateLimit(rule.Conf.(policies_api.Conf).Local.HTTP),
					))
				}
			}
			if len(existingRoutes) > 1 {
				for _, rule := range c.From {
					for _, route := range existingRoutes {
						allTags := envoy_metadata.ExtractListOfTags(route.Metadata)
						for _, tagz := range allTags {
							isSubset := rule.Subset.IsSubset(core_xds.SubsetFromTags(tagz))
							if !isSubset && !rule.Conf.(policies_api.Conf).Local.HTTP.Disabled {
								routes = append(routes, envoy_common.NewRoute(
									envoy_common.WithCluster(cluster),
									envoy_common.WithMatchHeaderRegex(envoy_routes.TagsHeaderName, tags.MatchRuleRegex(rule.Subset)),
									envoy_common.WithMeshRateLimit(rule.Conf.(policies_api.Conf).Local.HTTP),
								))
							} else {
								if rule.Conf.(policies_api.Conf).Local.HTTP.Disabled {
									delete(route.TypedPerFilterConfig, "envoy.filters.http.local_ratelimit")
								}
							}
						}
					}
				}
			}
			configurer := envoy_routes.RoutesConfigurer{
				Routes: routes,
			}
			newVh := &envoy_route.VirtualHost{}
			// generate new routes with our defaults
			configurer.Configure(newVh)
			newRoutes := newVh.Routes
			newRoutes = append(newRoutes, vh.Routes...)
			// apply new routes to old object
			vh.Routes = newRoutes
		}
		return nil
	}
	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpRoutes); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	return nil
	// return proto.MarshalAnyDeterministic(config)
}

func (c *Configurer) configureHttpListener(filterChain *envoy_listener.FilterChain) error {
	config := &envoy_extensions_filters_http_local_ratelimit_v3.LocalRateLimit{
		StatPrefix: "rate_limit",
	}
	pbst, err := proto.MarshalAnyDeterministic(config)
	if err != nil {
		return err
	}
	httpRateLimit := func(hcm *envoy_hcm.HttpConnectionManager) error {
		hcm.HttpFilters = append(hcm.HttpFilters,
			&envoy_hcm.HttpFilter{
				Name: "envoy.filters.http.local_ratelimit",
				ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
					TypedConfig: pbst,
				},
			})
		return nil
	}
	if err := listeners_v3.UpdateHTTPConnectionManager(filterChain, httpRateLimit); err != nil && !errors.Is(err, &listeners_v3.UnexpectedFilterConfigTypeError{}) {
		return err
	}
	return nil
}
