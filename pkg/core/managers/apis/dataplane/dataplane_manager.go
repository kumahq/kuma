package dataplane

import (
	"context"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"

	"github.com/Kong/kuma/pkg/core"

	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
)

func NewDataplaneManager(store core_store.ResourceStore, clusterName string) core_manager.ResourceManager {
	return &dataplaneManager{
		ResourceManager: core_manager.NewResourceManager(store),
		store:           store,
		clusterName:     clusterName,
	}
}

type dataplaneManager struct {
	core_manager.ResourceManager
	store       core_store.ResourceStore
	clusterName string
}

func (m *dataplaneManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	if err := resource.Validate(); err != nil {
		return err
	}
	opts := core_store.NewCreateOptions(fs...)

	dp := core_mesh.DataplaneResource{}
	if err := m.store.Get(ctx, &dp, core_store.GetByKey(opts.Name, opts.Mesh)); err != nil {
		return err
	}
	if m.clusterName != "" {
		for _, inbound := range dp.Spec.Networking.Inbound {
			if inbound.Tags == nil {
				inbound.Tags = make(map[string]string)
			}
			inbound.Tags[mesh_proto.ClusterTag] = m.clusterName
		}
	}
	return m.store.Create(ctx, resource, append(fs, core_store.CreatedAt(core.Now()), core_store.CreateWithOwner(&dp))...)
}

func (m *dataplaneManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	return m.ResourceManager.Update(ctx, resource, fs...)
}
