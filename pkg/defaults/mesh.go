package defaults

import (
	"context"

	"github.com/go-logr/logr"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

var defaultMeshKey = core_model.ResourceKey{
	Name: core_model.DefaultMesh,
}

func EnsureDefaultMeshExists(
	ctx context.Context,
	resManager core_manager.ResourceManager,
	logger logr.Logger,
	cfg kuma_cp.Config,
) error {
	if cfg.Defaults.SkipMeshCreation {
		log.V(1).Info("skipping default Mesh creation because KUMA_DEFAULTS_SKIP_MESH_CREATION is set to true")
		return nil
	}
	if cfg.IsFederatedZoneCP() {
		return nil // Mesh should be synced from Global CP
	}
	mesh := core_mesh.NewMeshResource()
	err := resManager.Get(ctx, mesh, core_store.GetBy(defaultMeshKey))
	if err == nil {
		logger.V(1).Info("default Mesh already exists. Skip creating default Mesh.")
		return nil
	}
	if !core_store.IsNotFound(err) {
		return err
	}
	if err := resManager.Create(ctx, mesh, core_store.CreateBy(defaultMeshKey)); err != nil {
		logger.V(1).Info("could not create default mesh", "err", err)
		return err
	}
	logger.Info("default Mesh created")
	return nil
}
