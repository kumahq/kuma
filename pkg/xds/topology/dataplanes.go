package topology

import (
	"context"

	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
)

// GetDataplanes returs list of Dataplane in provided Mesh and Ingresses (which are cluster-scoped, not mesh-scoped)
func GetDataplanes(ctx context.Context, rm manager.ReadOnlyResourceManager, mesh string) (*core_mesh.DataplaneResourceList, error) {
	dataplanes := &core_mesh.DataplaneResourceList{}
	if err := rm.List(ctx, dataplanes); err != nil {
		return nil, err
	}
	rv := &core_mesh.DataplaneResourceList{}
	for _, d := range dataplanes.Items {
		if d.GetMeta().GetMesh() == mesh || d.Spec.IsIngress() {
			_ = rv.AddItem(d)
		}
	}
	return rv, nil
}
