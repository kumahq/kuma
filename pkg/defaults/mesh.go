package defaults

import (
	"context"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

var defaultMeshKey = core_model.ResourceKey{
	Mesh: core_model.DefaultMesh,
	Name: core_model.DefaultMesh,
}

func (d *defaultsComponent) meshExists() (bool, error) {
	mesh := &mesh_core.MeshResource{}
	if err := d.resManager.Get(context.Background(), mesh, core_store.GetBy(defaultMeshKey)); err != nil {
		if core_store.IsResourceNotFound(err) {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (d *defaultsComponent) createDefaultMesh() error {
	return d.resManager.Create(context.Background(), &mesh_core.MeshResource{}, core_store.CreateBy(defaultMeshKey))
}

func (d *defaultsComponent) createMeshIfNotExist() error {
	exists, err := d.meshExists()
	if err != nil {
		return err
	}
	if exists {
		log.V(1).Info("default Mesh already exists. Skip creating default Mesh.")
	} else {
		log.Info("trying to create default Mesh")
		if err := d.createDefaultMesh(); err != nil {
			log.V(1).Info("could not create default mesh", "err", err)
			return err
		}
		log.Info("default Mesh created")
	}
	return nil
}
