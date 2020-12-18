package v2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
			filterChain, err := NewFilterChainBuilder(envoy.APIV2).
				Configure(HttpConnectionManager("stats")).
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
                '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                httpFilters:
                - name: envoy.filters.http.grpc_stats
                  typedConfig:
                    '@type': type.googleapis.com/envoy.config.filter.http.grpc_stats.v2alpha.FilterConfig
                    emitFilterState: true
                    statsForAllMethods: true
                - name: envoy.filters.http.router
                statPrefix: stats`,
		}),
	)
})
