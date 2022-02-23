package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

var _ = Describe("CommonVirtualHostConfigurer", func() {

	type testCase struct {
		virtualHostName string
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			routeConfiguration, err := NewVirtualHostBuilder(envoy.APIV3).
				Configure(CommonVirtualHost(given.virtualHostName)).
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
		Entry("basic VirtualHost", testCase{
			virtualHostName: "backend",
			expected: `
            name: backend
            domains:
            - '*'
`,
		}),
	)
})
