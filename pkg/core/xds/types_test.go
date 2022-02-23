package xds_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

var _ = Describe("xDS", func() {

	Describe("ParseProxyId(..)", func() {

		Context("valid input", func() {
			type testCase struct {
				nodeID   string
				expected core_xds.ProxyId
			}

			DescribeTable("should successfully parse",
				func(given testCase) {
					// when
					proxyId, err := core_xds.ParseProxyIdFromString(given.nodeID)

					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(*proxyId).To(Equal(given.expected))
				},
				Entry("mesh and name without namespace", testCase{
					nodeID:   "demo.example",
					expected: *core_xds.BuildProxyId("demo", "example"),
				}),
				Entry("name with namespace and mesh", testCase{
					nodeID:   "demo.example.sample",
					expected: *core_xds.BuildProxyId("demo", "example.sample"),
				}),
				Entry("mesh and name without namespace and proxy type", testCase{
					nodeID:   "demo.example",
					expected: *core_xds.BuildProxyId("demo", "example"),
				}),
				Entry("name with namespace and mesh and proxy type", testCase{
					nodeID:   "demo.example.sample",
					expected: *core_xds.BuildProxyId("demo", "example.sample"),
				}),
			)
		})

		Context("invalid input", func() {
			type testCase struct {
				nodeID      string
				expectedErr interface{}
			}

			DescribeTable("should fail to parse",
				func(given testCase) {
					// when
					key, err := core_xds.ParseProxyIdFromString(given.nodeID)

					// then
					Expect(err).To(MatchError(given.expectedErr))
					// and
					Expect(key).To(BeNil())
				},
				Entry("empty", testCase{
					nodeID:      "",
					expectedErr: "Envoy ID must not be nil",
				}),
				Entry("mesh without name and namespace", testCase{
					nodeID:      "demo",
					expectedErr: "the name should be provided after the dot",
				}),
				Entry("mesh with empty name", testCase{
					nodeID:      "demo.",
					expectedErr: "name must not be empty",
				}),
			)
		})
	})

	Describe("ProxyId(...).ToResourceKey()", func() {
		It("should convert proxy ID to resource key", func() {
			// given
			id := *core_xds.BuildProxyId("default", "demo")

			// when
			key := id.ToResourceKey()

			// then
			Expect(key.Mesh).To(Equal("default"))
			Expect(key.Name).To(Equal("demo"))
		})
	})

	Describe("TagSelectorSet", func() {
		Describe("Add()", func() {
			It("should be possible to add the first element to the set", func() {
				// given
				var set core_xds.TagSelectorSet
				// when
				actual := set.Add(mesh_proto.TagSelector{"service": "redis"})
				// then
				Expect(actual).To(HaveLen(1))
				Expect(actual).To(ConsistOf(mesh_proto.TagSelector{"service": "redis"}))
			})

			It("should be possible to add the second element to the set", func() {
				// given
				set := core_xds.TagSelectorSet{mesh_proto.TagSelector{"service": "redis"}}
				// when
				actual := set.Add(mesh_proto.TagSelector{"service": "elastic"})
				// then
				Expect(actual).To(HaveLen(2))
				Expect(actual).To(ConsistOf(mesh_proto.TagSelector{"service": "redis"}, mesh_proto.TagSelector{"service": "elastic"}))
			})

			It("should not be possible to add the second identical element", func() {
				// given
				set := core_xds.TagSelectorSet{mesh_proto.TagSelector{"service": "redis"}}
				// when
				actual := set.Add(mesh_proto.TagSelector{"service": "redis"})
				// then
				Expect(actual).To(HaveLen(1))
				Expect(actual).To(ConsistOf(mesh_proto.TagSelector{"service": "redis"}))
			})
		})
	})

	Describe("EndpointList", func() {
		Describe("Filter()", func() {
			type testCase struct {
				endpoints core_xds.EndpointList
				filter    mesh_proto.TagSelector
				expected  core_xds.EndpointList
			}
			DescribeTable("should filter out endpoints that don't match a given filter",
				func(given testCase) {
					// expect
					Expect(given.endpoints.Filter(given.filter)).To(Equal(given.expected))
				},
				Entry("`nil` filter", testCase{
					endpoints: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
						},
					}},
					filter: nil,
					expected: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
						},
					}},
				}),
				Entry("empty filter", testCase{
					endpoints: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
						},
					}},
					filter: mesh_proto.TagSelector{},
					expected: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
						},
					}},
				}),
				Entry("wildcard filter", testCase{
					endpoints: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
						},
					}},
					filter: mesh_proto.TagSelector{
						"service": "*",
					},
					expected: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
						},
					}},
				}),
				Entry("filter by tag that is missing", testCase{
					endpoints: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
						},
					}},
					filter: mesh_proto.TagSelector{
						"region": "us",
					},
					expected: nil,
				}),
				Entry("filter by 1 common tag", testCase{
					endpoints: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
							"version": "v1",
						},
					}, {
						Target: "192.168.0.2",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
							"version": "v2",
						},
					}},
					filter: mesh_proto.TagSelector{
						"service": "backend",
					},
					expected: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
							"version": "v1",
						},
					}, {
						Target: "192.168.0.2",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
							"version": "v2",
						},
					}},
				}),
				Entry("filter by 1 common tag and 1 unique tag", testCase{
					endpoints: core_xds.EndpointList{{
						Target: "192.168.0.1",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
							"version": "v1",
						},
					}, {
						Target: "192.168.0.2",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
							"version": "v2",
						},
					}},
					filter: mesh_proto.TagSelector{
						"service": "backend",
						"version": "v2",
					},
					expected: core_xds.EndpointList{{
						Target: "192.168.0.2",
						Port:   8080,
						Tags: map[string]string{
							"service": "backend",
							"version": "v2",
						},
					}},
				}),
			)
		})
	})

	Describe("ContainsTags", func() {
		// given
		endpoint := core_xds.Endpoint{
			Tags: map[string]string{
				"kuma.io/service": "backend",
				"version":         "v1",
			},
		}

		It("should match single tag", func() {
			// when
			contains := endpoint.ContainsTags(map[string]string{
				"kuma.io/service": "backend",
			})

			// then
			Expect(contains).To(BeTrue())
		})

		It("should match all the tags", func() {
			// when
			contains := endpoint.ContainsTags(map[string]string{
				"kuma.io/service": "backend",
				"version":         "v1",
			})

			// then
			Expect(contains).To(BeTrue())
		})

		It("should not match when value of a tag is different", func() {
			// when
			contains := endpoint.ContainsTags(map[string]string{
				"kuma.io/service": "backend",
				"version":         "v2",
			})

			// then
			Expect(contains).To(BeFalse())
		})

		It("should not match when endpoint has no such tag", func() {
			// when
			contains := endpoint.ContainsTags(map[string]string{
				"team": "xyz",
			})

			// then
			Expect(contains).To(BeFalse())
		})
	})
})
