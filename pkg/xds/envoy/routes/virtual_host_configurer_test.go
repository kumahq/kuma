package routes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/routes"
)

var _ = Describe("RouteConfigurationVirtualHostConfigurer", func() {

	type Opt = VirtualHostBuilderOpt
	type testCase struct {
		opts     []Opt
		expected string
	}

	Context("V3", func() {
		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// when
				routeConfiguration, err := NewRouteConfigurationBuilder(envoy.APIV3).
					Configure(VirtualHost(NewVirtualHostBuilder(envoy.APIV3).
						Configure(given.opts...))).
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
			Entry("basic virtual host", testCase{
				opts: []Opt{CommonVirtualHost("backend")},
				expected: `
            virtualHosts:
            - domains:
              - '*'
              name: backend
`,
			}),
			Entry("virtual host with domains", testCase{
				opts: []Opt{
					CommonVirtualHost("backend"),
					DomainNames("foo.example.com", "bar.example.com"),
				},
				expected: `
            virtualHosts:
            - domains:
              - foo.example.com
              - bar.example.com
              name: backend
`,
			}),
			Entry("virtual host with empty domains", testCase{
				opts: []Opt{
					CommonVirtualHost("backend"),
					DomainNames(),
				},
				expected: `
            virtualHosts:
            - domains:
              - '*'
              name: backend
`,
			}),
		)
	})
})
