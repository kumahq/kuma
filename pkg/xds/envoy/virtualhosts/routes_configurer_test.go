package virtualhosts_test

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_virtual_hosts "github.com/kumahq/kuma/v3/pkg/xds/envoy/virtualhosts"
)

var _ = Describe("RoutesConfigurer", func() {
	type testCase struct {
		routes                envoy_common.Routes
		configureRouteTimeout bool
		expected              string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			virtualHost := &envoy_config_route_v3.VirtualHost{}
			err := envoy_virtual_hosts.RoutesConfigurer{
				Routes:                given.routes,
				ConfigureRouteTimeout: given.configureRouteTimeout,
			}.
				Configure(virtualHost)
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(virtualHost)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("routes without timeouts", testCase{
			configureRouteTimeout: true,
			routes: []envoy_common.Route{
				envoy_common.NewRouteFromCluster(plugins_xds.NewClusterBuilder().WithName("backend").Build()),
			},
			expected: `
routes:
  - match:
      prefix: "/"
    route:
      timeout: "0s"
      cluster: backend`,
		}),
		Entry("routes without configured request timeout", testCase{
			configureRouteTimeout: false,
			routes: []envoy_common.Route{
				envoy_common.NewRouteFromCluster(plugins_xds.NewClusterBuilder().WithName("backend").Build()),
			},
			expected: `
routes:
  - match:
      prefix: "/"
    route:
      cluster: backend`,
		}),
		Entry("routes with multiple clusters and timeouts", testCase{
			configureRouteTimeout: true,
			routes: []envoy_common.Route{
				envoy_common.NewRoute(
					envoy_common.WithCluster(plugins_xds.NewClusterBuilder().WithName("backend-1").Build()),
					envoy_common.WithCluster(plugins_xds.NewClusterBuilder().WithName("backend-2").Build()),
				),
			},
			expected: `
routes:
  - match:
      prefix: "/"
    route:
      timeout: "0s"
      weightedClusters:
        clusters:
          - name: backend-1
            weight: 1
          - name: backend-2
            weight: 1`,
		}),
		Entry("routes with multiple clusters without configured request timeout", testCase{
			configureRouteTimeout: false,
			routes: []envoy_common.Route{
				envoy_common.NewRoute(
					envoy_common.WithCluster(plugins_xds.NewClusterBuilder().WithName("backend-1").Build()),
					envoy_common.WithCluster(plugins_xds.NewClusterBuilder().WithName("backend-2").Build()),
				),
			},
			expected: `
routes:
  - match:
      prefix: "/"
    route:
      weightedClusters:
        clusters:
          - name: backend-1
            weight: 1
          - name: backend-2
            weight: 1`,
		}),
		Entry("routes without any clusters", testCase{
			configureRouteTimeout: true,
			routes: []envoy_common.Route{
				envoy_common.NewRoute(),
			},
			expected: `
routes:
  - match:
      prefix: "/"
    route: {}`,
		}),
	)
})
