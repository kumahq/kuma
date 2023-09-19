package mesh

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

func EnsureCAs(ctx context.Context, caManagers core_ca.Managers, mesh *core_mesh.MeshResource, meshName string) error {
	backendsForType := make(map[string][]*v1alpha1.CertificateAuthorityBackend)
	for _, backend := range mesh.Spec.GetMtls().GetBackends() {
		backendsForType[backend.Type] = append(backendsForType[backend.Type], backend)
	}
	for typ, backends := range backendsForType {
		caManager, exist := caManagers[typ]
		if !exist { // this should be caught by validator earlier
			return errors.Errorf("CA manager for type %s does not exist", typ)
		}
		if err := caManager.EnsureBackends(ctx, mesh, backends); err != nil {
			return errors.Wrapf(err, "could not ensure CA backends of type %s", typ)
		}
	}
	return nil
}
