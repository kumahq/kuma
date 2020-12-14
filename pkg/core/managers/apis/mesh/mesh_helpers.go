package mesh

import (
	"context"

	"github.com/pkg/errors"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func EnsureEnabledCA(ctx context.Context, caManagers core_ca.Managers, mesh *mesh_core.MeshResource, meshName string) error {
	if mesh.GetEnabledCertificateAuthorityBackend() != nil {
		backend := mesh.GetEnabledCertificateAuthorityBackend()
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
