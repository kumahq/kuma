package mesh

import (
	"context"
	"time"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_registry "github.com/kumahq/kuma/pkg/core/resources/registry"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	defaults_mesh "github.com/kumahq/kuma/pkg/defaults/mesh"
)

func NewMeshManager(
	store core_store.ResourceStore,
	otherManagers core_manager.ResourceManager,
	caManagers core_ca.Managers,
	registry core_registry.TypeRegistry,
	validator MeshValidator,
	extensions context.Context,
	config kuma_cp.Config,
) core_manager.ResourceManager {
	meshManager := &meshManager{
		store:                      store,
		otherManagers:              otherManagers,
		caManagers:                 caManagers,
		registry:                   registry,
		meshValidator:              validator,
		unsafeDelete:               config.Store.UnsafeDelete,
		extensions:                 extensions,
		createMeshRoutingResources: config.Defaults.CreateMeshRoutingResources,
	}
	if config.Store.Type == config_store.KubernetesStore {
		meshManager.k8sStore = true
		meshManager.systemNamespace = config.Store.Kubernetes.SystemNamespace
	}
	return meshManager
}

type meshManager struct {
	store                      core_store.ResourceStore
	otherManagers              core_manager.ResourceManager
	caManagers                 core_ca.Managers
	registry                   core_registry.TypeRegistry
	meshValidator              MeshValidator
	unsafeDelete               bool
	extensions                 context.Context
	createMeshRoutingResources bool
	k8sStore                   bool
	systemNamespace            string
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

func (m *meshManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	opts := core_store.NewCreateOptions(fs...)
	mesh, err := m.mesh(resource)
	if err != nil {
		return err
	}
	if err := mesh.Default(); err != nil {
		return err
	}
	if err := core_model.Validate(resource); err != nil {
		return err
	}
	if err := m.meshValidator.ValidateCreate(ctx, opts.Name, mesh); err != nil {
		return err
	}
	// persist Mesh
	if err := m.store.Create(ctx, mesh, append(fs, core_store.CreatedAt(time.Now()))...); err != nil {
		return err
	}
	// We need to first persist the mesh so that we can hook up secrets (cert/key) as their owner in EnsureCAs.
	if err := EnsureCAs(ctx, m.caManagers, mesh, opts.Name); err != nil {
		return err
	}
	if err := defaults_mesh.EnsureDefaultMeshResources(
		ctx,
		m.otherManagers,
		mesh,
		mesh.Spec.GetSkipCreatingInitialPolicies(),
		m.extensions,
		m.createMeshRoutingResources,
		m.k8sStore,
		m.systemNamespace,
	); err != nil {
		return err
	}
	return nil
}

func (m *meshManager) Delete(ctx context.Context, resource core_model.Resource, fs ...core_store.DeleteOptionsFunc) error {
	mesh, err := m.mesh(resource)
	if err != nil {
		return err
	}
	opts := core_store.NewDeleteOptions(fs...)

	if !m.unsafeDelete {
		if err := m.meshValidator.ValidateDelete(ctx, opts.Name); err != nil {
			return err
		}
	}
	// delete Mesh first to avoid a state where a Mesh could exist without secrets.
	// even if removal of secrets fails later on, delete operation can be safely tried again.
	var notFoundErr error
	if err := m.store.Delete(ctx, mesh, fs...); err != nil {
		if core_store.IsNotFound(err) {
			notFoundErr = err
		} else { // ignore other errors so we can retry removing other resources
			return err
		}
	}
	// secrets are deleted via owner reference
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
	if err := mesh.Default(); err != nil {
		return err
	}
	if err := core_model.Validate(resource); err != nil {
		return err
	}

	currentMesh := core_mesh.NewMeshResource()
	if err := m.Get(ctx, currentMesh, core_store.GetBy(core_model.MetaToResourceKey(mesh.GetMeta())), core_store.GetByVersion(mesh.GetMeta().GetVersion())); err != nil {
		return err
	}
	if err := m.meshValidator.ValidateUpdate(ctx, currentMesh, mesh); err != nil {
		return err
	}
	if err := EnsureCAs(ctx, m.caManagers, mesh, mesh.Meta.GetName()); err != nil {
		return err
	}
	return m.store.Update(ctx, mesh, append(fs, core_store.ModifiedAt(time.Now()))...)
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
