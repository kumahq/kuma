package dataplaneinsight

import (
	"context"

	"github.com/Kong/kuma/pkg/core"

	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func NewDataplaneInsightManager(
	store core_store.ResourceStore,
	otherManagers core_manager.ResourceManager,
) core_manager.ResourceManager {
	return &dataplaneInsightManager{
		ResourceManager: core_manager.NewResourceManager(store),
		store:           store,
		otherManagers:   otherManagers,
	}
}

type dataplaneInsightManager struct {
	core_manager.ResourceManager

	store         core_store.ResourceStore
	otherManagers core_manager.ResourceManager
}

func (m *dataplaneInsightManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	if err := resource.Validate(); err != nil {
		return err
	}
	opts := core_store.NewCreateOptions(fs...)

	dp := core_mesh.DataplaneResource{}
	if err := m.store.Get(ctx, &dp, core_store.GetByKey(opts.Name, opts.Mesh)); err != nil {
		return err
	}
	return m.store.Create(ctx, resource, append(fs, core_store.CreatedAt(core.Now()), core_store.CreateWithOwner(&dp))...)
}
