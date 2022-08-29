package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("RateLimitConfigurer", func() {
	type testCase struct {
		input    []*core_mesh.RateLimitResource
		expected string
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			filterChain, err := NewFilterChainBuilder(envoy.APIV3).
				Configure(HttpConnectionManager("stats", false)).
				Configure(RateLimit(given.input)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())
			// when
			actual, err := util_proto.ToYAML(filterChain)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic input", testCase{
			input: []*core_mesh.RateLimitResource{
				{
					Spec: &mesh_proto.RateLimit{
						Sources: []*mesh_proto.Selector{
							{
								Match: map[string]string{
									"tag1": "value1",
									"tag2": "value2",
								},
							},
						},
						Conf: &mesh_proto.RateLimit_Conf{
							Http: &mesh_proto.RateLimit_Conf_Http{
								Requests: 100,
							},
						},
					},
				},
			},

			expected: `
            filters:
            - name: envoy.filters.network.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                httpFilters:
                - name: envoy.filters.http.local_ratelimit
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.local_ratelimit.v3.LocalRateLimit
                    statPrefix: rate_limit
                - name: envoy.filters.http.router
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                statPrefix: stats`,
		}),
	)
})
