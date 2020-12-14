package defaults

import (
	"context"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

var defaultMeshKey = core_model.ResourceKey{
	Name: core_model.DefaultMesh,
}

func (d *defaultsComponent) createMeshIfNotExist() error {
	mesh := mesh_core.NewMeshResource()
	err := d.resManager.Get(context.Background(), mesh, core_store.GetBy(defaultMeshKey))
	if err == nil {
		log.V(1).Info("default Mesh already exists. Skip creating default Mesh.")
		return nil
	}
	if !core_store.IsResourceNotFound(err) {
		return err
	}
	log.Info("trying to create default Mesh")
	if err := d.resManager.Create(context.Background(), mesh, core_store.CreateBy(defaultMeshKey)); err != nil {
		log.V(1).Info("could not create default mesh", "err", err)
		return err
	}
	log.Info("default Mesh created")
	return nil
}
