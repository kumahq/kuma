package endpoints_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/pkg/xds/envoy/endpoints"

	core_xds "github.com/Kong/kuma/pkg/core/xds"

	util_proto "github.com/Kong/kuma/pkg/util/proto"
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
					{Target: "192.168.0.1", Port: 8081},
					{Target: "192.168.0.2", Port: 8082},
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
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.2
                          portValue: 8082
`,
			}),
			Entry("with tags", testCase{
				cluster: "127.0.0.1:8080",
				endpoints: []core_xds.Endpoint{
					{Target: "192.168.0.1", Port: 8081, Tags: map[string]string{"service": "backend", "region": "us"}},
					{Target: "192.168.0.2", Port: 8082, Tags: map[string]string{"service": "backend", "region": "eu"}},
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
                          service: backend
                  - endpoint:
                      address:
                        socketAddress:
                          address: 192.168.0.2
                          portValue: 8082
                    metadata:
                      filterMetadata:
                        envoy.lb:
                          region: eu
                          service: backend
`,
			}),
		)
	})

	Describe("CreateLbMetadata()", func() {

		It("should handle `nil` map of tags", func() {
			// when
			metadata := CreateLbMetadata(nil)
			// then
			Expect(metadata).To(BeNil())
		})

		It("should handle empty map of tags", func() {
			// when
			metadata := CreateLbMetadata(map[string]string{})
			// then
			Expect(metadata).To(BeNil())
		})

		type testCase struct {
			tags     map[string]string
			expected string
		}
		DescribeTable("should generate Envoy metadata",
			func(given testCase) {
				// when
				metadata := CreateLbMetadata(given.tags)
				// and
				actual, err := util_proto.ToYAML(metadata)
				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(actual).To(MatchYAML(given.expected))
			},
			Entry("map with 1 tag", testCase{
				tags: map[string]string{
					"service": "redis",
				},
				expected: `
                filterMetadata:
                  envoy.lb:
                    service: redis
`,
			}),
			Entry("map with multiple tags", testCase{
				tags: map[string]string{
					"service": "redis",
					"version": "v1",
					"region":  "eu",
				},
				expected: `
                filterMetadata:
                  envoy.lb:
                    service: redis
                    version: v1
                    region: eu
`,
			}),
		)
	})
})
