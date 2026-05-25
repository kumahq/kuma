package xds

import (
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
})
