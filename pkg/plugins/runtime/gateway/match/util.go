package match

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
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
	mgr  manager.ReadOnlyResourceManager
	opts []store.ListOptionsFunc
}

var _ manager.ReadOnlyResourceManager = &MeshedResourceManager{}

func (m *MeshedResourceManager) Get(ctx context.Context, r model.Resource, opts ...store.GetOptionsFunc) error {
	return m.mgr.Get(ctx, r, opts...)
}

func (m *MeshedResourceManager) List(ctx context.Context, r model.ResourceList, opts ...store.ListOptionsFunc) error {
	return m.mgr.List(ctx, r, append(m.opts, opts...)...)
}

func ManagerForMesh(m manager.ReadOnlyResourceManager, mesh string) *MeshedResourceManager {
	return &MeshedResourceManager{
		mgr:  m,
		opts: []store.ListOptionsFunc{store.ListByMesh(mesh)},
	}
}
