package virtualhosts_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	"github.com/kumahq/kuma/v3/pkg/xds/envoy"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/routes"
	. "github.com/kumahq/kuma/v3/pkg/xds/envoy/virtualhosts"
)

var _ = Describe("RouteConfigurationVirtualHostConfigurer", func() {
	type Opt = VirtualHostBuilderOpt
	type testCase struct {
		expected        string
		opts            []Opt
		virtualHostName string
	}

	Context("V3", func() {
		DescribeTable("should generate proper Envoy config",
			func(given testCase) {
				// when
				routeConfiguration, err := NewRouteConfigurationBuilder(envoy.APIV3, "route_configuration").
					Configure(VirtualHost(NewVirtualHostBuilder(envoy.APIV3, given.virtualHostName).
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
				virtualHostName: "backend",
				opts:            []Opt{},
				expected: `
            name: route_configuration
            virtualHosts:
            - domains:
              - '*'
              name: backend
`,
			}),
			Entry("virtual host with domains", testCase{
				virtualHostName: "backend",
				opts: []Opt{
					DomainNames("foo.example.com", "bar.example.com"),
				},
				expected: `
            name: route_configuration
            virtualHosts:
            - domains:
              - foo.example.com
              - bar.example.com
              name: backend
`,
			}),
			Entry("virtual host with empty domains", testCase{
				virtualHostName: "backend",
				opts: []Opt{
					DomainNames(),
				},
				expected: `
            name: route_configuration
            virtualHosts:
            - domains:
              - '*'
              name: backend
`,
			}),
		)
	})
})
