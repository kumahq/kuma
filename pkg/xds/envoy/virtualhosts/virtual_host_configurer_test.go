package virtualhosts_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/routes"
	. "github.com/kumahq/kuma/pkg/xds/envoy/virtualhosts"
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
			Entry("virtual host with retry", testCase{
				opts: []Opt{
					CommonVirtualHost("backend"),
					DomainNames(),
					Retry(
						&core_mesh.RetryResource{
							Spec: &mesh_proto.Retry{
								Conf: &mesh_proto.Retry_Conf{
									Http: &mesh_proto.Retry_Conf_Http{
										NumRetries:       util_proto.UInt32(7),
										RetriableMethods: []mesh_proto.HttpMethod{mesh_proto.HttpMethod_GET, mesh_proto.HttpMethod_POST},
									},
								},
							},
						},
						mesh.ProtocolHTTP,
					),
				},
				expected: `
            virtualHosts:
            - domains:
              - '*'
              name: backend
              retryPolicy:
                numRetries: 7
                retriableRequestHeaders:
                    - name: :method
                      stringMatch:
                        exact: GET
                    - name: :method
                      stringMatch:
                        exact: POST
                retryOn: gateway-error,connect-failure,refused-stream
`,
			}),
		)
	})
})
