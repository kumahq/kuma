package mesh

import (
	"context"
	"errors"
	"fmt"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type MeshValidator interface {
	ValidateCreate(ctx context.Context, name string, resource *core_mesh.MeshResource) error
	ValidateUpdate(ctx context.Context, previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error
	ValidateDelete(ctx context.Context, name string) error
}

type meshValidator struct {
	CaManagers core_ca.Managers
	Store      core_store.ResourceStore
}

func NewMeshValidator(caManagers core_ca.Managers, store core_store.ResourceStore) MeshValidator {
	return &meshValidator{
		CaManagers: caManagers,
		Store:      store,
	}
}

func (m *meshValidator) ValidateCreate(ctx context.Context, name string, resource *core_mesh.MeshResource) error {
	if err := ValidateMTLSBackends(ctx, m.CaManagers, name, resource); err != nil {
		return err
	}
	return nil
}

func ValidateMTLSBackends(ctx context.Context, caManagers core_ca.Managers, name string, resource *core_mesh.MeshResource) error {
	verr := validators.ValidationError{}
	path := validators.RootedAt("mtls").Field("backends")

	for idx, backend := range resource.Spec.GetMtls().GetBackends() {
		caManager, exist := caManagers[backend.Type]
		if !exist {
			verr.AddViolationAt(path.Index(idx).Field("type"), "could not find installed plugin for this type")
			return verr.OrNil()
		} else if err := caManager.ValidateBackend(ctx, name, backend); err != nil {
			var configErr *validators.ValidationError
			if errors.As(err, &configErr) {
				verr.AddErrorAt(path.Index(idx).Field("conf"), *configErr)
			} else {
				verr.AddViolationAt(path.Index(idx), err.Error())
				return err
			}
		}
	}
	return verr.OrNil()
}

func (m *meshValidator) ValidateUpdate(ctx context.Context, previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error {
	if err := m.validateMTLSBackendChange(previousMesh, newMesh); err != nil {
		return err
	}
	if err := ValidateMTLSBackends(ctx, m.CaManagers, newMesh.Meta.GetName(), newMesh); err != nil {
		return err
	}
	return nil
}

func (m *meshValidator) ValidateDelete(ctx context.Context, name string) error {
	if err := ValidateNoActiveDP(ctx, name, m.Store); err != nil {
		return err
	}
	return nil
}

func ValidateNoActiveDP(ctx context.Context, name string, store core_store.ResourceStore) error {
	dps := core_mesh.DataplaneResourceList{}
	validationErr := &validators.ValidationError{}
	if err := store.List(ctx, &dps, core_store.ListByMesh(name)); err != nil {
		return fmt.Errorf("unable to list Dataplanes: %w", err)
	}
	if len(dps.Items) != 0 {
		validationErr.AddViolation("mesh", "unable to delete mesh, there are still some dataplanes attached")
		return validationErr
	}
	return nil
}

func (m *meshValidator) validateMTLSBackendChange(previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error {
	verr := validators.ValidationError{}
	if previousMesh.MTLSEnabled() && newMesh.MTLSEnabled() && previousMesh.Spec.GetMtls().GetEnabledBackend() != newMesh.Spec.GetMtls().GetEnabledBackend() {
		verr.AddViolation("mtls.enabledBackend", "Changing CA when mTLS is enabled is forbidden. Disable mTLS first and then change the CA")
	}
	return verr.OrNil()
}
