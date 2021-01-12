package endpoints_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("Endpoints", func() {

	Describe("CreateStaticEndpoint()", func() {
		It("should generate 'static' Endpoints", func() {
			// given
			expected := `
            clusterName: localhost:8080
            endpoints:
            - lbEndpoints:
              - endpoint:
                  address:
                    socketAddress:
                      address: 127.0.0.1
                      portValue: 8080
`
			// when
			resource := CreateStaticEndpoint("localhost:8080", "127.0.0.1", 8080)

			// then
			actual, err := util_proto.ToYAML(resource)

			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(expected))
		})
	})

	Describe("ClusterLoadAssignment()", func() {
		type testCase struct {
			cluster   string
			endpoints []core_xds.Endpoint
			expected  string
		}
		DescribeTable("should generate ClusterLoadAssignment",
			func(given testCase) {
				// when
				resource := CreateClusterLoadAssignment(given.cluster, given.endpoints)

				// then
				actual, err := util_proto.ToYAML(resource)

				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("without tags", testCase{
				cluster: "127.0.0.1:8080",
				endpoints: []core_xds.Endpoint{
					{
						Target: "192.168.0.1",
						Port:   8081,
						Weight: 2,
					},
					{
						Target: "192.168.0.2",
						Port:   8082,
						Weight: 1,
					},
				},
				expected: `
                clusterName: 127.0.0.1:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.1
                          portValue: 8081
                    loadBalancingWeight: 2
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.2
                          portValue: 8082
                    loadBalancingWeight: 1
`,
			}),
			Entry("stable output", testCase{
				cluster: "127.0.0.1:8080",
				endpoints: []core_xds.Endpoint{ // endpoints in different order, but result is the same as in previous test
					{
						Target: "192.168.0.2",
						Port:   8082,
						Weight: 1,
					},
					{
						Target: "192.168.0.1",
						Port:   8081,
						Weight: 2,
					},
				},
				expected: `
                clusterName: 127.0.0.1:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.1
                          portValue: 8081
                    loadBalancingWeight: 2
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.2
                          portValue: 8082
                    loadBalancingWeight: 1
`,
			}),
			Entry("with tags", testCase{
				cluster: "127.0.0.1:8080",
				endpoints: []core_xds.Endpoint{
					{
						Target: "192.168.0.1",
						Port:   8081,
						Tags:   map[string]string{"kuma.io/service": "backend", "region": "us"},
						Weight: 1,
					},
					{
						Target: "192.168.0.2",
						Port:   8082,
						Tags:   map[string]string{"kuma.io/service": "backend", "region": "eu"},
						Weight: 2,
					},
				},
				expected: `
                clusterName: 127.0.0.1:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.1
                          portValue: 8081
                    metadata:
                      filterMetadata:
                        envoy.lb:
                          region: us
                        envoy.transport_socket_match:
                          region: us
                    loadBalancingWeight: 1
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.2
                          portValue: 8082
                    metadata:
                      filterMetadata:
                        envoy.lb:
                          region: eu
                        envoy.transport_socket_match:
                          region: eu
                    loadBalancingWeight: 2
`,
			}),
			Entry("with locality tags", testCase{
				cluster: "127.0.0.1:8080",
				endpoints: []core_xds.Endpoint{
					{
						Target: "192.168.0.1",
						Port:   8081,
						Tags:   map[string]string{"kuma.io/service": "backend", "kuma.io/region": "us", "kuma.io/zone": "west"},
						Weight: 1,
					},
					{
						Target: "192.168.0.2",
						Port:   8082,
						Tags:   map[string]string{"kuma.io/service": "backend", "kuma.io/region": "eu", "kuma.io/zone": "west"},
						Weight: 2,
					},
				},
				expected: `
                clusterName: 127.0.0.1:8080
                endpoints:
                - lbEndpoints:
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.1
                          portValue: 8081
                    metadata:
                      filterMetadata:
                        envoy.lb:
                          kuma.io/region: us
                          kuma.io/zone: west
                        envoy.transport_socket_match:
                          kuma.io/region: us
                          kuma.io/zone: west
                    loadBalancingWeight: 1
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.2
                          portValue: 8082
                    metadata:
                      filterMetadata:
                        envoy.lb:
                          kuma.io/region: eu
                          kuma.io/zone: west
                        envoy.transport_socket_match:
                          kuma.io/region: eu
                          kuma.io/zone: west
                    loadBalancingWeight: 2
`,
			}),
		)
	})
})
