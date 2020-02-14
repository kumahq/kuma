package routes_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/routes"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("CommonRouteConfigurationConfigurer", func() {

	type testCase struct {
		routeName string
		expected  string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			routeConfiguration, err := NewRouteConfigurationBuilder().
				Configure(CommonRouteConfiguration(given.routeName)).
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
		Entry("basic RouteConfiguration", testCase{
			routeName: "outbound:backend",
			expected: `
            name: outbound:backend
            validateClusters: true
`,
		}),
	)
})
