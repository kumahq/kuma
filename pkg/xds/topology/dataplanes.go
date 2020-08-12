package topology

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/runtime"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

// GetDataplanes returns list of Dataplane in provided Mesh and Ingresses (which are cluster-scoped, not mesh-scoped)
func GetDataplanes(rt runtime.RuntimeContext, ctx context.Context, mesh string) (*core_mesh.DataplaneResourceList, error) {
	dataplanes := &core_mesh.DataplaneResourceList{}
	if err := rt.ReadOnlyResourceManager().List(ctx, dataplanes); err != nil {
		return nil, err
	}
	rv := &core_mesh.DataplaneResourceList{}
	for _, d := range dataplanes.Items {
		if err := ResolveAddress(rt, d); err != nil {
			return nil, err
		}
		if d.GetMeta().GetMesh() == mesh || d.Spec.IsIngress() {
			_ = rv.AddItem(d)
		}
	}
	return rv, nil
}

// ResolveAddress resolves 'dataplane.networking.address' if it has DNS name in it. This is a crucial feature for
// some environments specifically AWS ECS. Dataplane resource has to be created before running Kuma DP, but IP address
// will be assigned only after container's start. Being able to set dp's address as a DNS name solves this problem
func ResolveAddress(rt runtime.RuntimeContext, dataplane *core_mesh.DataplaneResource) error {
	ips, err := rt.LookupIP()(dataplane.Spec.Networking.Address)
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return errors.Errorf("can't resolve address %v", dataplane.Spec.Networking.Address)
	}
	dataplane.Spec.Networking.Address = ips[0].String()
	return nil
}
