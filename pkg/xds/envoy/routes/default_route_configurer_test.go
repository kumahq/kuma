package routes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/routes"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
	envoy_common "github.com/Kong/kuma/pkg/xds/envoy"
)

var _ = Describe("DefaultRouteConfigurer", func() {

	type testCase struct {
		clusters []envoy_common.ClusterInfo
		expected string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			routeConfiguration, err := NewVirtualHostBuilder().
				Configure(DefaultRoute(given.clusters...)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(routeConfiguration)
			// then
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic VirtualHost with a single destination cluster", testCase{
			clusters: []envoy_common.ClusterInfo{
				{Name: "backend", Weight: 200},
			},
			expected: `
            routes:
            - match:
                prefix: /
              route:
                cluster: backend
`,
		}),
		Entry("basic VirtualHost with weighted destination clusters", testCase{
			clusters: []envoy_common.ClusterInfo{
				{Name: "backend{version=v1}", Weight: 30},
				{Name: "backend{version=v2}", Weight: 70},
			},
			expected: `
            routes:
            - match:
                prefix: /
              route:
                weightedClusters:
                  clusters:
                  - name: backend{version=v1}
                    weight: 30
                  - name: backend{version=v2}
                    weight: 70
`,
		}),
	)
})
