package v3_test

import (
	envoy_config_route_v3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_routes "github.com/kumahq/kuma/pkg/xds/envoy/routes/v3"
)

var _ = Describe("RoutesConfigurer", func() {
	type testCase struct {
		routes   envoy_common.Routes
		expected string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			virtualHost := &envoy_config_route_v3.VirtualHost{}
			err := envoy_routes.RoutesConfigurer{Routes: given.routes}.
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
			routes: []envoy_common.Route{
				envoy_common.NewRouteFromCluster(envoy_common.NewCluster(envoy_common.WithName("backend"))),
			},
			expected: `
routes:
  - match:
      prefix: "/"
    route:
      timeout: "0s"
      cluster: backend`,
		}),
	)
})
