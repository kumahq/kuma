package topology

import (
	"context"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/pkg/errors"
)

// GetDataplanes returns list of Dataplane in provided Mesh and Ingresses (which are cluster-scoped, not mesh-scoped)
func GetDataplanes(ctx context.Context, rm manager.ReadOnlyResourceManager, mesh string) (*core_mesh.DataplaneResourceList, error) {
	dataplanes := &core_mesh.DataplaneResourceList{}
	if err := rm.List(ctx, dataplanes); err != nil {
		return nil, err
	}
	rv := &core_mesh.DataplaneResourceList{}
	for _, d := range dataplanes.Items {
		if err := ResolveAddress(d); err != nil {
			return nil, err
		}
		if d.GetMeta().GetMesh() == mesh || d.Spec.IsIngress() {
			_ = rv.AddItem(d)
		}
	}
	return rv, nil
}

func ResolveAddress(dataplane *core_mesh.DataplaneResource) error {
	ips, err := core.LookupIP(dataplane.Spec.Networking.Address)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.Errorf("can't resolve address %v", dataplane.Spec.Networking.Address)
	}
	dataplane.Spec.Networking.Address = ips[0].String()
	return nil
}
