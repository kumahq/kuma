package mesh

import (
	"context"

	"github.com/pkg/errors"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
)

type MeshValidator struct {
	CaManagers core_ca.Managers
	Store      store.ResourceStore
}

func (m *MeshValidator) ValidateCreate(ctx context.Context, name string, resource *core_mesh.MeshResource) error {
	if err := m.validateMTLSBackends(ctx, name, resource); err != nil {
		return err
	}
	return nil
}

func (m *MeshValidator) validateMTLSBackends(ctx context.Context, name string, resource *core_mesh.MeshResource) error {
	verr := validators.ValidationError{}
	path := validators.RootedAt("mtls").Field("backends")
	for idx, backend := range resource.Spec.GetMtls().GetBackends() {
		caManager, exist := m.CaManagers[backend.Type]
		if !exist {
			verr.AddViolationAt(path.Index(idx).Field("type"), "could not find installed plugin for this type")
			return verr.OrNil()
		} else if err := caManager.ValidateBackend(ctx, name, backend); err != nil {
			if configErr, ok := err.(*validators.ValidationError); ok {
				verr.AddErrorAt(path.Index(idx).Field("config"), *configErr)
			} else {
				verr.AddViolationAt(path.Index(idx), err.Error())
				return err
			}
		}
	}
	return verr.OrNil()
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

func (m *MeshValidator) ValidateDelete(ctx context.Context, name string) error {
	dps := core_mesh.DataplaneResourceList{}
	validationErr := &validators.ValidationError{}
	if err := m.Store.List(ctx, &dps, store.ListByMesh(name)); err != nil {
		return errors.Wrap(err, "unable to list Dataplanes")
	}
	if len(dps.Items) != 0 {
		validationErr.AddViolation("mesh", "unable to delete mesh, there are still some dataplanes attached")
		return validationErr
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
