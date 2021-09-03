package mesh

import (
	"context"

	"github.com/pkg/errors"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func EnsureCAs(ctx context.Context, caManagers core_ca.Managers, mesh *core_mesh.MeshResource, meshName string) error {
	for _, backend := range mesh.Spec.GetMtls().GetBackends() {
		caManager, exist := caManagers[backend.Type]
		if !exist { // this should be caught by validator earlier
			return errors.Errorf("CA manager for type %s does not exist", backend.Type)
		}
		if err := caManager.Ensure(ctx, meshName, backend); err != nil {
			return errors.Wrapf(err, "could not create CA of backend name %s", backend.Name)
		}
	}
	return nil
}
