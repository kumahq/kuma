package xds

import (
	"time"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	rules_inbound "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

var _ = Describe("MeshRateLimit configurer", func() {
	It("should skip matched-route config when there is no base or matched rate limit", func() {
		filterChain := httpFilterChainWithSingleRoute()
		rules := []*rules_inbound.Rule{{
			Match: &common_api.Match{
				SpiffeID: &common_api.SpiffeIDMatch{
					Type:  common_api.ExactMatchType,
					Value: "spiffe://default/client",
				},
			},
			Conf: api.Conf{
				Local: &api.Local{
					HTTP: &api.LocalHTTP{
						OnRateLimit: &api.OnRateLimit{
							Status: pointer.To(uint32(429)),
						},
					},
				},
			},
		}}

		Expect(ConfigureMatchedRoutesOnFilterChain(filterChain, api.Conf{}, rules)).To(Succeed())

		Expect(filterChain.Filters).To(HaveLen(1))
		Expect(filterChain.Filters[0].GetName()).To(Equal("envoy.filters.network.http_connection_manager"))

		hcm := httpConnectionManagerFromFilterChain(filterChain)
		Expect(hcm.GetHttpFilters()).To(HaveLen(1))
		Expect(hcm.GetHttpFilters()[0].GetName()).To(Equal("envoy.filters.http.router"))

		routes := hcm.GetRouteConfig().GetVirtualHosts()[0].GetRoutes()
		Expect(routes).To(HaveLen(1))
		Expect(routes[0].GetMatch().GetFilterState()).To(BeEmpty())
		Expect(routes[0].GetTypedPerFilterConfig()).To(BeNil())
	})

	It("should add the HTTP local rate limit filter for gateway route-specific rate limits", func() {
		filterChain := httpFilterChainWithSingleRouteNamed("route-1")
		hcm := httpConnectionManagerFromFilterChain(filterChain)
		routeConfig := hcm.GetRouteConfig()

		configurer := &Configurer{
			Element: subsetutils.MeshElement(),
			Rules: core_rules.Rules{{
				Subset: subsetutils.Subset{{
					Key:   core_rules.RuleMatchesHashTag,
					Value: "route-1",
				}},
				Conf: api.Conf{
					Local: &api.Local{
						HTTP: &api.LocalHTTP{
							RequestRate: &api.Rate{Num: 1, Interval: k8s.Duration{Duration: 10 * time.Second}},
							OnRateLimit: &api.OnRateLimit{
								Status: pointer.To(uint32(428)),
							},
						},
					},
				},
			}},
		}

		Expect(configurer.ConfigureGatewayRoute(routeConfig, filterChain)).To(Succeed())

		hcm = httpConnectionManagerFromFilterChain(filterChain)
		Expect(hcm.GetHttpFilters()).To(HaveLen(2))
		Expect(hcm.GetHttpFilters()[0].GetName()).To(Equal(httpLocalRateLimitFilterName))
		Expect(hcm.GetHttpFilters()[1].GetName()).To(Equal("envoy.filters.http.router"))

		routes := routeConfig.GetVirtualHosts()[0].GetRoutes()
		Expect(routes).To(HaveLen(1))
		Expect(routes[0].GetTypedPerFilterConfig()).To(HaveKey(httpLocalRateLimitFilterName))
	})
})

func httpFilterChainWithSingleRoute() *envoy_listener.FilterChain {
	return httpFilterChainWithSingleRouteNamed("")
}

func httpFilterChainWithSingleRouteNamed(name string) *envoy_listener.FilterChain {
	routerConfig, err := util_proto.MarshalAnyDeterministic(&envoy_router.Router{})
	Expect(err).ToNot(HaveOccurred())

	hcmConfig, err := util_proto.MarshalAnyDeterministic(&envoy_hcm.HttpConnectionManager{
		HttpFilters: []*envoy_hcm.HttpFilter{{
			Name: "envoy.filters.http.router",
			ConfigType: &envoy_hcm.HttpFilter_TypedConfig{
				TypedConfig: routerConfig,
			},
		}},
		RouteSpecifier: &envoy_hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &envoy_route.RouteConfiguration{
				VirtualHosts: []*envoy_route.VirtualHost{{
					Routes: []*envoy_route.Route{{
						Name: name,
						Match: &envoy_route.RouteMatch{
							PathSpecifier: &envoy_route.RouteMatch_Prefix{Prefix: "/"},
						},
						Action: &envoy_route.Route_Route{
							Route: &envoy_route.RouteAction{
								ClusterSpecifier: &envoy_route.RouteAction_Cluster{
									Cluster: "backend",
								},
							},
						},
					}},
				}},
			},
		},
	})
	Expect(err).ToNot(HaveOccurred())

	return &envoy_listener.FilterChain{
		Filters: []*envoy_listener.Filter{{
			Name: "envoy.filters.network.http_connection_manager",
			ConfigType: &envoy_listener.Filter_TypedConfig{
				TypedConfig: hcmConfig,
			},
		}},
	}
}

func httpConnectionManagerFromFilterChain(filterChain *envoy_listener.FilterChain) *envoy_hcm.HttpConnectionManager {
	hcm := &envoy_hcm.HttpConnectionManager{}
	Expect(util_proto.UnmarshalAnyTo(filterChain.Filters[0].GetTypedConfig(), hcm)).To(Succeed())
	return hcm
}
