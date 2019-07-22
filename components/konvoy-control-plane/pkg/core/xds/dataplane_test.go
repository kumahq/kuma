package xds_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/xds"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

var _ = Describe("xDS", func() {

	Describe("ParseDataplaneId(..)", func() {

		Context("valid input", func() {
			type testCase struct {
				node     *envoy_core.Node
				expected core_model.ResourceKey
			}

			DescribeTable("should successfully parse",
				func(given testCase) {
					// when
					key, err := core_xds.ParseDataplaneId(given.node)

					// then
					Expect(err).ToNot(HaveOccurred())
					// and
					Expect(*key).To(Equal(given.expected))
				},
				Entry("no name and namespace", testCase{
					node: &envoy_core.Node{},
					expected: core_model.ResourceKey{
						Mesh: "default", Namespace: "default", Name: "",
					},
				}),
				Entry("name without namespace", testCase{
					node: &envoy_core.Node{
						Id: "example",
					},
					expected: core_model.ResourceKey{
						Mesh: "default", Namespace: "default", Name: "example",
					},
				}),
				Entry("name with namespace", testCase{
					node: &envoy_core.Node{
						Id: "example.demo",
					},
					expected: core_model.ResourceKey{
						Mesh: "default", Namespace: "demo", Name: "example",
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
					key, err := core_xds.ParseDataplaneId(given.node)

					// then
					Expect(err).To(MatchError(given.expectedErr))
					// and
					Expect(key).To(BeNil())
				},
				Entry("`nil`", testCase{
					node:        nil,
					expectedErr: "Envoy node must not be nil",
				}),
			)
		})
	})
})
