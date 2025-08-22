package unified_naming

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
)

func Enabled(meta *core_xds.DataplaneMetadata, mesh *core_mesh.MeshResource) bool {
	if meta == nil || mesh == nil {
		return false
	}

	return meta.HasFeature(xds_types.FeatureUnifiedResourceNaming) &&
		mesh.Spec.GetMeshServices().GetMode() == mesh_proto.Mesh_MeshServices_Exclusive
}
