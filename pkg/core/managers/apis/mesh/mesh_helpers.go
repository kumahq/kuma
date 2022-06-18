package mesh

import (
	"context"
	"fmt"

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
			return fmt.Errorf("CA manager for type %s does not exist", typ)
		}
		if err := caManager.EnsureBackends(ctx, meshName, backends); err != nil {
			return fmt.Errorf("could not ensure CA backends of type %s: %w", typ, err)
		}
	}
	return nil
}
