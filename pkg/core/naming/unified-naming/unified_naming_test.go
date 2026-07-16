package unified_naming_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	unified_naming "github.com/kumahq/kuma/v3/pkg/core/naming/unified-naming"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
)

func meshWithMode(mode mesh_proto.Mesh_MeshServices_Mode) *core_mesh.MeshResource {
	return &core_mesh.MeshResource{
		Spec: &mesh_proto.Mesh{
			MeshServices: &mesh_proto.Mesh_MeshServices{Mode: mode},
		},
	}
}

func metaWithFeature() *core_xds.DataplaneMetadata {
	return &core_xds.DataplaneMetadata{
		Features: xds_types.Features{xds_types.FeatureUnifiedResourceNaming: true},
	}
}

var _ = Describe("Enabled", func() {
	type testCase struct {
		meta     *core_xds.DataplaneMetadata
		mesh     *core_mesh.MeshResource
		expected bool
	}

	DescribeTable(
		"should require both the feature and an Exclusive mesh",
		func(given testCase) {
			Expect(unified_naming.Enabled(given.meta, given.mesh)).To(Equal(given.expected))
		},
		Entry("feature and Exclusive mode", testCase{
			meta:     metaWithFeature(),
			mesh:     meshWithMode(mesh_proto.Mesh_MeshServices_Exclusive),
			expected: true,
		}),
		Entry("feature and nil MeshServices block, which defaults to Exclusive", testCase{
			meta:     metaWithFeature(),
			mesh:     &core_mesh.MeshResource{Spec: &mesh_proto.Mesh{}},
			expected: true,
		}),
		Entry("feature but explicitly Disabled mode", testCase{
			meta:     metaWithFeature(),
			mesh:     meshWithMode(mesh_proto.Mesh_MeshServices_Disabled),
			expected: false,
		}),
		Entry("feature but explicitly Everywhere mode", testCase{
			meta:     metaWithFeature(),
			mesh:     meshWithMode(mesh_proto.Mesh_MeshServices_Everywhere),
			expected: false,
		}),
		Entry("feature but explicitly ReachableBackends mode", testCase{
			meta:     metaWithFeature(),
			mesh:     meshWithMode(mesh_proto.Mesh_MeshServices_ReachableBackends),
			expected: false,
		}),
		Entry("Exclusive mode but no feature", testCase{
			meta:     &core_xds.DataplaneMetadata{},
			mesh:     meshWithMode(mesh_proto.Mesh_MeshServices_Exclusive),
			expected: false,
		}),
		Entry("nil metadata", testCase{
			meta:     nil,
			mesh:     meshWithMode(mesh_proto.Mesh_MeshServices_Exclusive),
			expected: false,
		}),
		Entry("nil mesh", testCase{
			meta:     metaWithFeature(),
			mesh:     nil,
			expected: false,
		}),
	)
})
