package xds_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	core_xds "github.com/Kong/kuma/pkg/core/xds"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

var _ = Describe("xDS", func() {

	Describe("ParseProxyId(..)", func() {

		Context("valid input", func() {
			type testCase struct {
				node     *envoy_core.Node
				expected core_xds.ProxyId
			}

			DescribeTable("should successfully parse",
				func(given testCase) {
					// when
					proxyId, err := core_xds.ParseProxyId(given.node)

					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(*proxyId).To(Equal(given.expected))
				},
				Entry("mesh and name without namespace", testCase{
					node: &envoy_core.Node{
						Id: "pilot.example",
					},
					expected: core_xds.ProxyId{
						Mesh: "pilot", Namespace: "default", Name: "example",
					},
				}),
				Entry("name with namespace and mesh", testCase{
					node: &envoy_core.Node{
						Id: "pilot.example.demo",
					},
					expected: core_xds.ProxyId{
						Mesh: "pilot", Namespace: "demo", Name: "example",
					},
				}),
			)
		})

		Context("invalid input", func() {
			type testCase struct {
				node        *envoy_core.Node
				expectedErr interface{}
			}

			DescribeTable("should fail to parse",
				func(given testCase) {
					// when
					key, err := core_xds.ParseProxyId(given.node)

					// then
					Expect(err).To(MatchError(given.expectedErr))
					// and
					Expect(key).To(BeNil())
				},
				Entry("`nil`", testCase{
					node:        nil,
					expectedErr: "Envoy node must not be nil",
				}),
				Entry("empty", testCase{
					node:        &envoy_core.Node{},
					expectedErr: "mesh must not be empty",
				}),
				Entry("mesh without name and namespace", testCase{
					node: &envoy_core.Node{
						Id: "pilot",
					},
					expectedErr: "the name should be provided after the dot",
				}),
				Entry("mesh with empty name", testCase{
					node: &envoy_core.Node{
						Id: "pilot.",
					},
					expectedErr: "name must not be empty",
				}),
				Entry("mesh with empty namespace", testCase{
					node: &envoy_core.Node{
						Id: "pilot.default.",
					},
					expectedErr: "namespace must not be empty",
				}),
			)
		})
	})

	Describe("ProxyId(...).ToResourceKey()", func() {
		It("should convert proxy ID to resource key", func() {
			// given
			id := core_xds.ProxyId{
				Mesh:      "default",
				Namespace: "pilot",
				Name:      "demo",
			}

			// when
			key := id.ToResourceKey()

			// then
			Expect(key.Namespace).To(Equal("pilot"))
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
})
