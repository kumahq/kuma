package mesh

import (
	"context"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_store "github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

type MeshValidator interface {
	ValidateCreate(ctx context.Context, name string, resource *core_mesh.MeshResource) error
	ValidateUpdate(ctx context.Context, previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error
	ValidateDelete(ctx context.Context, name string) error
}

type meshValidator struct {
	Store core_store.ResourceStore
}

func NewMeshValidator(store core_store.ResourceStore) MeshValidator {
	return &meshValidator{
		Store: store,
	}
}

func (m *meshValidator) ValidateCreate(ctx context.Context, name string, resource *core_mesh.MeshResource) error {
	var verr validators.ValidationError
	if len(name) > 63 {
		verr.AddViolation("name", "cannot be longer than 63 characters")
	}
	return verr.OrNil()
}

func (m *meshValidator) ValidateUpdate(ctx context.Context, previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error {
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
		return errors.Wrap(err, "unable to list Dataplanes")
	}
	if len(dps.Items) != 0 {
		validationErr.AddViolation("mesh", "unable to delete mesh, there are still some dataplanes attached")
		return validationErr
	}
	return nil
}
