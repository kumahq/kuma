package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("gRPCStatsConfigurer", func() {
	type testCase struct {
		expected string
	}
	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			filterChain, err := NewFilterChainBuilder(envoy.APIV3).
				Configure(HttpConnectionManager("stats", false)).
				Configure(GrpcStats()).
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
			expected: `
            filters:
            - name: envoy.filters.network.http_connection_manager
              typedConfig:
                '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                httpFilters:
                - name: envoy.filters.http.grpc_stats
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.grpc_stats.v3.FilterConfig
                    emitFilterState: true
                    statsForAllMethods: true
                - name: envoy.filters.http.router
                  typedConfig:
                    '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                statPrefix: stats`,
		}),
	)
})
