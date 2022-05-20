package v3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/core/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	. "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

var _ = Describe("StaticEndpointsConfigurer", func() {

	type testCase struct {
		listenerName     string
		listenerProtocol xds.SocketAddressProtocol
		listenerAddress  string
		listenerPort     uint32
		path             string
		clusterName      string
		expected         string
	}

	DescribeTable("should generate proper Envoy config",
		func(given testCase) {
			// when
			listener, err := NewListenerBuilder(envoy_common.APIV3).
				Configure(InboundListener(given.listenerName, given.listenerAddress, given.listenerPort, given.listenerProtocol)).
				Configure(FilterChain(NewFilterChainBuilder(envoy_common.APIV3).
					Configure(StaticEndpoints(given.listenerName,
						[]*envoy_common.StaticEndpointPath{
							{
								ClusterName: given.clusterName,
								Path:        given.path,
								RewritePath: "/stats/prometheus",
							},
						})))).
				Build()
			// then
			Expect(err).ToNot(HaveOccurred())

			// when
			actual, err := util_proto.ToYAML(listener)
			Expect(err).ToNot(HaveOccurred())
			// and
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("prometheus endpoint without transparent proxying", testCase{
			listenerName:    "kuma:metrics:prometheus",
			listenerAddress: "192.168.0.1",
			listenerPort:    8080,
			path:            "/non-standard-path",
			clusterName:     "kuma:envoy:admin",
			expected: `
            name: kuma:metrics:prometheus
            trafficDirection: INBOUND
            address:
              socketAddress:
                address: 192.168.0.1
                portValue: 8080
            enableReusePort: false
            filterChains:
            - filters:
              - name: envoy.filters.network.http_connection_manager
                typedConfig:
                  '@type': type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                  httpFilters:
                  - name: envoy.filters.http.router
                    typedConfig:
                      '@type': type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
                  routeConfig:
                    validateClusters: false
                    virtualHosts:
                    - domains:
                      - '*'
                      name: kuma:metrics:prometheus
                      routes:
                      - match:
                          prefix: /non-standard-path
                        route:
                          cluster: kuma:envoy:admin
                          prefixRewrite: /stats/prometheus
                  statPrefix: kuma_metrics_prometheus
`,
		}),
	)

})
