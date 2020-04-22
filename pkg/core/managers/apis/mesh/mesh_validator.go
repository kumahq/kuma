package mesh

import (
	"context"

	core_ca "github.com/Kong/kuma/pkg/core/ca"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
)

type MeshValidator struct {
	CaManagers core_ca.CaManagers
}

func (m *MeshValidator) ValidateCreate(ctx context.Context, name string, resource *core_mesh.MeshResource) error {
	if err := m.validateMTLSBackends(ctx, name, resource); err != nil {
		return err
	}
	return nil
}

func (m *MeshValidator) validateMTLSBackends(ctx context.Context, name string, resource *core_mesh.MeshResource) error {
	for idx, backend := range resource.Spec.GetMtls().GetBackends() {
		caManager, exist := m.CaManagers[backend.Type]
		if !exist {
			verr := validators.ValidationError{}
			verr.AddViolationAt(validators.RootedAt("backends").Index(idx).Field("type"), "could not find installed plugin for this type")
			return verr.OrNil()
		}
		if err := caManager.ValidateBackend(ctx, name, *backend); err != nil {
			return err
		}
	}
	return nil
}

func (m *MeshValidator) ValidateUpdate(ctx context.Context, previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error {
	if err := m.validateMTLSBackendChange(previousMesh, newMesh); err != nil {
		return err
	}
	if err := m.validateMTLSBackends(ctx, newMesh.Meta.GetName(), newMesh); err != nil {
		return err
	}
	return nil
}

func (m *MeshValidator) validateMTLSBackendChange(previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error {
	verr := validators.ValidationError{}
	if previousMesh.MTLSEnabled() && newMesh.MTLSEnabled() && previousMesh.Spec.GetMtls().GetEnabledBackend() != newMesh.Spec.GetMtls().GetEnabledBackend() {
		verr.AddViolation("mtls.enabledBackend", "Changing CA when mTLS is enabled is forbidden. Disable mTLS first and then change the CA")
	}
	return verr.OrNil()
}
