package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
)

var _ = Describe("Mesh", func() {
	Describe("MeshServicesMode", func() {
		type testCase struct {
			meshServices *Mesh_MeshServices
			expected     Mesh_MeshServices_Mode
		}
		DescribeTable(
			"should resolve mode",
			func(given testCase) {
				mesh := &Mesh{MeshServices: given.meshServices}
				Expect(mesh.MeshServicesMode()).To(Equal(given.expected))
			},
			Entry("nil block defaults to Exclusive", testCase{
				meshServices: nil,
				expected:     Mesh_MeshServices_Exclusive,
			}),
			Entry("explicit empty block stays Disabled", testCase{
				meshServices: &Mesh_MeshServices{},
				expected:     Mesh_MeshServices_Disabled,
			}),
			Entry("explicit Disabled stays Disabled", testCase{
				meshServices: &Mesh_MeshServices{Mode: Mesh_MeshServices_Disabled},
				expected:     Mesh_MeshServices_Disabled,
			}),
			Entry("explicit Everywhere unchanged", testCase{
				meshServices: &Mesh_MeshServices{Mode: Mesh_MeshServices_Everywhere},
				expected:     Mesh_MeshServices_Everywhere,
			}),
			Entry("explicit ReachableBackends unchanged", testCase{
				meshServices: &Mesh_MeshServices{Mode: Mesh_MeshServices_ReachableBackends},
				expected:     Mesh_MeshServices_ReachableBackends,
			}),
			Entry("explicit Exclusive unchanged", testCase{
				meshServices: &Mesh_MeshServices{Mode: Mesh_MeshServices_Exclusive},
				expected:     Mesh_MeshServices_Exclusive,
			}),
		)
	})
})
