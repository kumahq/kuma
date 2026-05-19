package defaults

import (
	"context"

	"github.com/go-logr/logr"

	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	config_store "github.com/kumahq/kuma/v2/pkg/config/core/resources/store"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
	defaults_mesh "github.com/kumahq/kuma/v2/pkg/defaults/mesh"
)

var defaultMeshKey = core_model.ResourceKey{
	Name: core_model.DefaultMesh,
}

func EnsureDefaultMeshExists(
	ctx context.Context,
	resManager core_manager.ResourceManager,
	logger logr.Logger,
	cfg kuma_cp.Config,
	extensions context.Context,
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

// EnsureDefaultMeshResourcesUpToDate reconciles the default policies of every
// Mesh on CP boot. Default policies created by older CP versions predate
// computed labels — without 'kuma.io/zone' their KRI carries an empty zone
// slot and /_kri lookups fail. Reconciliation heals such resources in place.
// Default policies an operator deleted are left absent — only labels of
// existing resources are reconciled, never recreated.
func EnsureDefaultMeshResourcesUpToDate(
	ctx context.Context,
	resManager core_manager.ResourceManager,
	logger logr.Logger,
	cfg kuma_cp.Config,
	extensions context.Context,
) error {
	if cfg.Defaults.SkipMeshCreation {
		return nil
	}
	if cfg.IsFederatedZoneCP() {
		return nil // default resources are synced from Global CP
	}
	meshes := &core_mesh.MeshResourceList{}
	if err := resManager.List(ctx, meshes); err != nil {
		return err
	}
	zone := cfg.Multizone.Zone.Name
	for _, mesh := range meshes.Items {
		if err := defaults_mesh.EnsureDefaultMeshResources(
			ctx,
			resManager,
			mesh,
			mesh.Spec.GetSkipCreatingInitialPolicies(),
			extensions,
			cfg.Defaults.CreateMeshRoutingResources,
			cfg.Store.Type == config_store.KubernetesStore,
			cfg.Store.Kubernetes.SystemNamespace,
			cfg.Mode,
			zone,
			true, // reconcile labels of existing resources, never recreate
		); err != nil {
			return err
		}
	}
	return nil
}
