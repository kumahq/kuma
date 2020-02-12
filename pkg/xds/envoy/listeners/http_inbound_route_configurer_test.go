package listeners_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/listeners"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("HttpInboundRouteConfigurer", func() {

	type testCase struct {
		listenerName    string
		listenerAddress string
		listenerPort    uint32
		statsName       string
		cluster         ClusterInfo
		expected        string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder().
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort)).
				Configure(HttpConnectionManager(given.statsName)).
				Configure(HttpInboundRoute(given.cluster)).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("basic http_connection_manager with a single destination cluster", testCase{
			listenerName:    "inbound:192.168.0.1:8080",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			statsName:       "localhost:8080",
			cluster:         ClusterInfo{Name: "localhost:8080", Weight: 200},
			expected: `
            name: inbound:192.168.0.1:8080
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            filterChains:
            - filters:
              - name: envoy.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.config.filter.network.http_connection_manager.v2.HttpConnectionManager
                  httpFilters:
                  - name: envoy.router
                  routeConfig:
                    name: inbound
                    validateClusters: true
                    virtualHosts:
                    - domains:
                      - '*'
                      name: local_service
                      routes:
                      - match:
                          prefix: /
                        route:
                          cluster: localhost:8080
                  statPrefix: localhost_8080
`,
		}),
	)
})
