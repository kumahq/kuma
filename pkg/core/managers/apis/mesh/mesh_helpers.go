package mesh

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"

	"github.com/pkg/errors"
)

func CreateDefaultMesh(resManager core_manager.ResourceManager, name string, template mesh_proto.Mesh) error {
	defaultMesh := mesh_core.MeshResource{}

	key := core_model.ResourceKey{Mesh: name, Name: name}

	if err := resManager.Get(context.Background(), &defaultMesh, core_store.GetBy(key)); err != nil {
		if core_store.IsResourceNotFound(err) {
			defaultMesh.Spec = template
			core.Log.WithName("bootstrap").Info("Creating default mesh from the settings", "mesh", defaultMesh.Spec, "name", name)

			if err := resManager.Create(context.Background(), &defaultMesh, core_store.CreateBy(key)); err != nil {
				return errors.Wrapf(err, "Failed to create `default` Mesh resource in a given resource store")
			}
		} else {
			return err
		}
	}

	return nil
}
