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

func meshResource() *core_mesh.MeshResource {
	return &core_mesh.MeshResource{
		Spec: &mesh_proto.Mesh{},
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
		"should require the feature",
		func(given testCase) {
			Expect(unified_naming.Enabled(given.meta, given.mesh)).To(Equal(given.expected))
		},
		Entry("feature present", testCase{
			meta:     metaWithFeature(),
			mesh:     meshResource(),
			expected: true,
		}),
		Entry("feature absent", testCase{
			meta:     &core_xds.DataplaneMetadata{},
			mesh:     meshResource(),
			expected: false,
		}),
		Entry("nil metadata", testCase{
			meta:     nil,
			mesh:     meshResource(),
			expected: false,
		}),
		Entry("nil mesh", testCase{
			meta:     metaWithFeature(),
			mesh:     nil,
			expected: false,
		}),
	)
})
