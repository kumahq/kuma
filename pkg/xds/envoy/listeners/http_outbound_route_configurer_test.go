package listeners_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("HttpOutboundRouteConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		routeName       string
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder().
				Configure(OutboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
				Configure(FilterChain(NewFilterChainBuilder().
					Configure(HttpConnectionManager(given.statsName)).
					Configure(HttpOutboundRoute(given.routeName)))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager with an outbound route", testCase{
			listenerName:    "outbound:127.0.0.1:18080",
			listenerAddress: "127.0.0.1",
			listenerPort:    18080,
			statsName:       "127.0.0.1:18080",
			routeName:       "outbound:backend",
			expected: `
            name: outbound:127.0.0.1:18080
            trafficDirection: OUTBOUND
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 18080
            filterChains:
            - filters:
              - name: envoy.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  httpFilters:
                  - name: envoy.router
                  rds:
                    configSource:
                      ads: {}
                    routeConfigName: outbound:backend
                  statPrefix: "127_0_0_1_18080"
`,
		}),
	)
})
