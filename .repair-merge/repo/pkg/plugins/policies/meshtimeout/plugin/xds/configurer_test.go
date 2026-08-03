package xds

import (
	"time"

	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	rules_inbound "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
)

var _ = Describe("MeshTimeout configurer", func() {
	DescribeTable("matchedRouteSupportsTimeouts",
		func(routeAction *envoy_route.RouteAction, expected bool) {
			Expect(matchedRouteSupportsTimeouts(routeAction)).To(Equal(expected))
		},
		Entry("nil route action", nil, false),
		Entry("cluster route", &envoy_route.RouteAction{
			ClusterSpecifier: &envoy_route.RouteAction_Cluster{
				Cluster: "backend",
			},
		}, true),
		Entry("cluster header route", &envoy_route.RouteAction{
			ClusterSpecifier: &envoy_route.RouteAction_ClusterHeader{
				ClusterHeader: "x-backend",
			},
		}, true),
		Entry("weighted route with named clusters", &envoy_route.RouteAction{
			ClusterSpecifier: &envoy_route.RouteAction_WeightedClusters{
				WeightedClusters: &envoy_route.WeightedCluster{
					Clusters: []*envoy_route.WeightedCluster_ClusterWeight{{
						Name: "backend",
					}},
				},
			},
		}, true),
		Entry("weighted route with unnamed clusters", &envoy_route.RouteAction{
			ClusterSpecifier: &envoy_route.RouteAction_WeightedClusters{
				WeightedClusters: &envoy_route.WeightedCluster{
					Clusters: []*envoy_route.WeightedCluster_ClusterWeight{{}},
				},
			},
		}, false),
	)

	It("should keep same spiffeID matches with different SNI as separate matched routes", func() {
		routeConfiguration := routeConfigurationWithClusterRoute()
		rules := []*rules_inbound.Rule{
			matchedTimeoutRule("db.example.com", "2s", ""),
			matchedTimeoutRule("cache.example.com", "5s", ""),
		}

		Expect(ConfigureMatchedRoutes(routeConfiguration, rules)).To(Succeed())

		routes := routeConfiguration.VirtualHosts[0].Routes
		Expect(routes).To(HaveLen(3))

		dbRoute := findMatchedRouteBySNI(routes, "db.example.com")
		cacheRoute := findMatchedRouteBySNI(routes, "cache.example.com")

		Expect(dbRoute).ToNot(BeNil())
		Expect(cacheRoute).ToNot(BeNil())
		Expect(filterStateExact(dbRoute, matchSpiffeIDFilterStateKey)).To(Equal("spiffe://default/client"))
		Expect(filterStateExact(cacheRoute, matchSpiffeIDFilterStateKey)).To(Equal("spiffe://default/client"))
		Expect(dbRoute.GetRoute().GetTimeout().AsDuration()).To(Equal(2 * time.Second))
		Expect(cacheRoute.GetRoute().GetTimeout().AsDuration()).To(Equal(5 * time.Second))
		Expect(dbRoute.GetRoute().IdleTimeout).To(BeNil())
		Expect(cacheRoute.GetRoute().IdleTimeout).To(BeNil())
		Expect(routes[2].GetMatch().GetFilterState()).To(BeEmpty())
	})

	It("should merge a spiffe-only matched route into more specific SNI variants", func() {
		routeConfiguration := routeConfigurationWithClusterRoute()
		rules := []*rules_inbound.Rule{
			matchedTimeoutRule("", "2s", ""),
			matchedTimeoutRule("db.example.com", "", "4s"),
			matchedTimeoutRule("cache.example.com", "", "7s"),
		}

		Expect(ConfigureMatchedRoutes(routeConfiguration, rules)).To(Succeed())

		routes := routeConfiguration.VirtualHosts[0].Routes
		Expect(routes).To(HaveLen(4))

		genericRoute := findMatchedRouteWithoutSNI(routes)
		dbRoute := findMatchedRouteBySNI(routes, "db.example.com")
		cacheRoute := findMatchedRouteBySNI(routes, "cache.example.com")

		Expect(genericRoute).ToNot(BeNil())
		Expect(dbRoute).ToNot(BeNil())
		Expect(cacheRoute).ToNot(BeNil())
		Expect(genericRoute.GetRoute().GetTimeout().AsDuration()).To(Equal(2 * time.Second))
		Expect(genericRoute.GetRoute().IdleTimeout).To(BeNil())
		Expect(dbRoute.GetRoute().GetTimeout().AsDuration()).To(Equal(2 * time.Second))
		Expect(dbRoute.GetRoute().GetIdleTimeout().AsDuration()).To(Equal(4 * time.Second))
		Expect(cacheRoute.GetRoute().GetTimeout().AsDuration()).To(Equal(2 * time.Second))
		Expect(cacheRoute.GetRoute().GetIdleTimeout().AsDuration()).To(Equal(7 * time.Second))
	})
})

func routeConfigurationWithClusterRoute() *envoy_route.RouteConfiguration {
	return &envoy_route.RouteConfiguration{
		VirtualHosts: []*envoy_route.VirtualHost{{
			Routes: []*envoy_route.Route{{
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
	}
}

func matchedTimeoutRule(sni, requestTimeout, streamIdleTimeout string) *rules_inbound.Rule {
	match := &common_api.Match{
		SpiffeID: &common_api.SpiffeIDMatch{
			Type:  common_api.ExactMatchType,
			Value: "spiffe://default/client",
		},
	}
	if sni != "" {
		match.SNI = &common_api.SNIMatch{
			Type:  common_api.SNIExactMatchType,
			Value: sni,
		}
	}

	conf := api.Conf{
		Http: &api.Http{},
	}
	if requestTimeout != "" {
		conf.Http.RequestTimeout = parseDuration(requestTimeout)
	}
	if streamIdleTimeout != "" {
		conf.Http.StreamIdleTimeout = parseDuration(streamIdleTimeout)
	}

	return &rules_inbound.Rule{
		Match: match,
		Conf:  conf,
	}
}

func findMatchedRouteBySNI(routes []*envoy_route.Route, sni string) *envoy_route.Route {
	for _, route := range routes {
		if filterStateExact(route, matchSNIStateKey) == sni {
			return route
		}
	}
	return nil
}

func findMatchedRouteWithoutSNI(routes []*envoy_route.Route) *envoy_route.Route {
	for _, route := range routes {
		if filterStateExact(route, matchSpiffeIDFilterStateKey) != "" &&
			filterStateExact(route, matchSNIStateKey) == "" {
			return route
		}
	}
	return nil
}

func filterStateExact(route *envoy_route.Route, key string) string {
	if route == nil || route.GetMatch() == nil {
		return ""
	}

	for _, matcher := range route.GetMatch().GetFilterState() {
		if matcher.GetKey() == key && matcher.GetStringMatch() != nil {
			return matcher.GetStringMatch().GetExact()
		}
	}
	return ""
}

func parseDuration(value string) *kube_meta.Duration {
	duration, err := time.ParseDuration(value)
	Expect(err).ToNot(HaveOccurred())
	return &kube_meta.Duration{Duration: duration}
}
