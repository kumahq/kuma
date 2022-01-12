package match

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

// MergeSelectors merges the given tags in order.
func MergeSelectors(tags ...mesh_proto.TagSelector) mesh_proto.TagSelector {
	merged := mesh_proto.TagSelector{}

	for _, t := range tags {
		for k, v := range t {
			merged[k] = v
		}
	}

	return merged
}

// MeshedResourceManager is a ReadOnlyResourceManager bound to a specific
// mesh. All operations performed by this resource manager will implicitly
// be scoped by the given mesh.
type MeshedResourceManager struct {
	mesh string
	opts []store.ListOptionsFunc
}

func ManagerForMesh(mesh string) *MeshedResourceManager {
	return &MeshedResourceManager{
		mesh: mesh,
		opts: []store.ListOptionsFunc{store.ListByMesh(mesh)},
	}
}
