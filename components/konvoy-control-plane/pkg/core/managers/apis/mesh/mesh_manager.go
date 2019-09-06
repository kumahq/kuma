package mesh

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	builtin_ca "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/ca/builtin"
	core_mesh "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/manager"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
)

func NewMeshManager(store core_store.ResourceStore, builtinCaManager builtin_ca.BuiltinCaManager) core_manager.ResourceManager {
	return &meshManager{
		store:            store,
		builtinCaManager: builtinCaManager,
	}
}

type meshManager struct {
	store            core_store.ResourceStore
	builtinCaManager builtin_ca.BuiltinCaManager
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
	// default CA
	if mesh.Spec.Mtls == nil {
		mesh.Spec.Mtls = &mesh_proto.Mesh_Mtls{}
	}
	if mesh.Spec.Mtls.Ca == nil {
		mesh.Spec.Mtls.Ca = &mesh_proto.CertificateAuthority{}
	}
	if mesh.Spec.Mtls.Ca.Type == nil {
		mesh.Spec.Mtls.Ca.Type = &mesh_proto.CertificateAuthority_Builtin_{
			Builtin: &mesh_proto.CertificateAuthority_Builtin{},
		}
	}
	// keep creation of Mesh and Built-in CA in sync
	var rollback func() error
	defer func() {
		if errs != nil && rollback != nil {
			errs = multierr.Append(errs, rollback())
		}
	}()
	// create Built-in CA
	switch mesh.Spec.GetMtls().GetCa().GetType().(type) {
	case *mesh_proto.CertificateAuthority_Builtin_:
		opts := core_store.NewCreateOptions(fs...)
		if err := m.builtinCaManager.Create(ctx, opts.Name); err != nil {
			return errors.Wrapf(err, "failed to create Builtin CA for a given mesh")
		}
		rollback = func() error {
			return m.builtinCaManager.Delete(ctx, opts.Name)
		}
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
	if err := m.store.Delete(ctx, mesh, fs...); err != nil {
		return err
	}
	// delete CA
	name := core_store.NewDeleteOptions(fs...).Mesh
	if err := m.builtinCaManager.Delete(ctx, name); err != nil {
		return errors.Wrapf(err, "failed to delete Builtin CA for a given mesh")
	}
	return nil
}

func (m *meshManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	mesh, err := m.mesh(resource)
	if err != nil {
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
