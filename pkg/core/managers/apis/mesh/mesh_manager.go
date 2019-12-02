package mesh

import (
	"context"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	builtin_ca "github.com/Kong/kuma/pkg/core/ca/builtin"
	provided_ca "github.com/Kong/kuma/pkg/core/ca/provided"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_registry "github.com/Kong/kuma/pkg/core/resources/registry"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	secrets_manager "github.com/Kong/kuma/pkg/core/secrets/manager"
	validators_mesh "github.com/Kong/kuma/pkg/core/validators/apis/mesh"
)

func NewMeshManager(
	store core_store.ResourceStore,
	builtinCaManager builtin_ca.BuiltinCaManager,
	providedCaManager provided_ca.ProvidedCaManager,
	otherManagers core_manager.ResourceManager,
	secretManager secrets_manager.SecretManager,
	registry core_registry.TypeRegistry,
) core_manager.ResourceManager {
	validator := validators_mesh.MeshValidator{
		ProvidedCaManager: providedCaManager,
	}
	return &meshManager{
		store:             store,
		builtinCaManager:  builtinCaManager,
		providedCaManager: providedCaManager,
		otherManagers:     otherManagers,
		secretManager:     secretManager,
		registry:          registry,
		meshValidator:     validator,
	}
}

type meshManager struct {
	store             core_store.ResourceStore
	builtinCaManager  builtin_ca.BuiltinCaManager
	providedCaManager provided_ca.ProvidedCaManager
	otherManagers     core_manager.ResourceManager
	secretManager     secrets_manager.SecretManager
	registry          core_registry.TypeRegistry
	meshValidator     validators_mesh.MeshValidator
}

func (m *meshManager) Get(ctx context.Context, resource core_model.Resource, fs ...core_store.GetOptionsFunc) error {
	mesh, err := m.mesh(resource)
	if err != nil {
		return err
	}
	return m.store.Get(ctx, mesh, fs...)
}

func (m *meshManager) List(ctx context.Context, list core_model.ResourceList, fs ...core_store.ListOptionsFunc) error {
	meshes, err := m.meshes(list)
	if err != nil {
		return err
	}
	return m.store.List(ctx, meshes, fs...)
}

func (m *meshManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) (errs error) {
	mesh, err := m.mesh(resource)
	if err != nil {
		return err
	}
	// apply defaults, e.g. Builtin CA
	mesh.Default()
	if err := resource.Validate(); err != nil {
		return err
	}
	// keep creation of Mesh and Built-in CA in sync
	var rollback func() error
	defer func() {
		if errs != nil && rollback != nil {
			errs = multierr.Append(errs, rollback())
		}
	}()
	opts := core_store.NewCreateOptions(fs...)
	switch mesh.Spec.GetMtls().GetCa().GetType().(type) {
	// create Built-in CA
	case *mesh_proto.CertificateAuthority_Builtin_:
		if err := m.builtinCaManager.Create(ctx, opts.Name); err != nil {
			return errors.Wrapf(err, "failed to create Builtin CA for a given mesh")
		}
		rollback = func() error {
			return m.builtinCaManager.Delete(ctx, opts.Name)
		}
	}
	if err := m.meshValidator.ValidateCreate(ctx, opts.Name, mesh); err != nil {
		return err
	}
	// persist Mesh
	if err := m.store.Create(ctx, mesh, fs...); err != nil {
		return err
	}
	return nil
}

func (m *meshManager) Delete(ctx context.Context, resource core_model.Resource, fs ...core_store.DeleteOptionsFunc) error {
	mesh, err := m.mesh(resource)
	if err != nil {
		return err
	}
	// delete Mesh first to avoid a state where a Mesh could exist without a Built-in CA.
	// even if removal of Built-in CA fails later on, delete opration can be safely tried again.
	var notFoundErr error
	if err := m.store.Delete(ctx, mesh, fs...); err != nil {
		if core_store.IsResourceNotFound(err) {
			notFoundErr = err
		} else { // ignore other errors so we can retry removing other resources
			return err
		}
	}
	opts := core_store.NewDeleteOptions(fs...)
	// delete CA
	name := core_store.NewDeleteOptions(fs...).Mesh
	if err := m.builtinCaManager.Delete(ctx, name); err != nil && !core_store.IsResourceNotFound(err) {
		return errors.Wrapf(err, "failed to delete Builtin CA for a given mesh")
	}
	if err := m.providedCaManager.DeleteCa(ctx, name); err != nil && !core_store.IsResourceNotFound(err) {
		return errors.Wrap(err, "failed to delete Provided CA for a given mesh")
	}
	// delete all other secrets
	if err := m.secretManager.DeleteAll(ctx, core_store.DeleteAllByMesh(opts.Mesh)); err != nil {
		return errors.Wrap(err, "could not delete associated secrets")
	}
	// delete other resources associated by mesh
	for _, typ := range m.registry.ListTypes() {
		list, err := m.registry.NewList(typ)
		if err != nil {
			return err
		}
		if err := m.otherManagers.DeleteAll(ctx, list, core_store.DeleteAllByMesh(opts.Name)); err != nil {
			return errors.Wrap(err, "could not delete associated resources")
		}
	}
	return notFoundErr
}

func (m *meshManager) DeleteAll(ctx context.Context, list core_model.ResourceList, fs ...core_store.DeleteAllOptionsFunc) error {
	if _, err := m.meshes(list); err != nil {
		return err
	}
	return core_manager.DeleteAllResources(m, ctx, list, fs...)
}

func (m *meshManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	mesh, err := m.mesh(resource)
	if err != nil {
		return err
	}
	// apply defaults, e.g. Builtin CA
	mesh.Default()
	if err := resource.Validate(); err != nil {
		return err
	}

	currentMesh := &core_mesh.MeshResource{}
	if err := m.Get(ctx, currentMesh, core_store.GetBy(core_model.MetaToResourceKey(mesh.GetMeta()))); err != nil {
		return err
	}
	if err := m.meshValidator.ValidateUpdate(ctx, currentMesh, mesh); err != nil {
		return err
	}
	return m.store.Update(ctx, mesh, fs...)
}

func (m *meshManager) mesh(resource core_model.Resource) (*core_mesh.MeshResource, error) {
	mesh, ok := resource.(*core_mesh.MeshResource)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.MeshResource)(nil), resource)
	}
	return mesh, nil
}

func (m *meshManager) meshes(list core_model.ResourceList) (*core_mesh.MeshResourceList, error) {
	meshes, ok := list.(*core_mesh.MeshResourceList)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.MeshResourceList)(nil), list)
	}
	return meshes, nil
}
