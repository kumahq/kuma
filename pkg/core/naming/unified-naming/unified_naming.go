package unified_naming

import (
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
)

func Enabled(meta *core_xds.DataplaneMetadata, mesh *core_mesh.MeshResource) bool {
	if meta == nil || mesh == nil {
		return false
	}

	return meta.HasFeature(xds_types.FeatureUnifiedResourceNaming)
}
