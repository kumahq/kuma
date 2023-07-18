package defaults

import (
	"context"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

var defaultMeshKey = core_model.ResourceKey{
	Name: core_model.DefaultMesh,
}

func CreateMeshIfNotExist(ctx context.Context, resManager core_manager.ResourceManager) error {
	logger := kuma_log.AddFieldsFromCtx(log, ctx)
	mesh := core_mesh.NewMeshResource()
	err := resManager.Get(ctx, mesh, core_store.GetBy(defaultMeshKey))
	if err == nil {
		logger.V(1).Info("default Mesh already exists. Skip creating default Mesh.")
		return nil
	}
	if !core_store.IsResourceNotFound(err) {
		return err
	}
	logger.Info("trying to create default Mesh")
	if err := resManager.Create(ctx, mesh, core_store.CreateBy(defaultMeshKey)); err != nil {
		logger.V(1).Info("could not create default mesh", "err", err)
		return err
	}
	logger.Info("default Mesh created")
	return nil
}
