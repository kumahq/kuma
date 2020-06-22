package dataplane

import (
	"context"

	"github.com/go-errors/errors"

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
	dp, err := m.dataplane(resource)
	if err != nil {
		return err
	}

	m.setInboundsClusterTag(dp)
	m.setGatewayClusterTag(dp)

	return m.store.Create(ctx, resource, append(fs, core_store.CreatedAt(core.Now()))...)
}

func (m *dataplaneManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	dp, err := m.dataplane(resource)
	if err != nil {
		return err
	}

	m.setInboundsClusterTag(dp)
	m.setGatewayClusterTag(dp)

	return m.ResourceManager.Update(ctx, resource, fs...)
}

func (m *dataplaneManager) dataplane(resource core_model.Resource) (*core_mesh.DataplaneResource, error) {
	dp, ok := resource.(*core_mesh.DataplaneResource)
	if !ok {
		return nil, errors.Errorf("invalid resource type: expected=%T, got=%T", (*core_mesh.DataplaneResource)(nil), resource)
	}
	return dp, nil
}

func (m *dataplaneManager) setInboundsClusterTag(dp *core_mesh.DataplaneResource) {
	if m.clusterName == "" || dp.Spec.Networking == nil {
		return
	}

	for _, inbound := range dp.Spec.Networking.Inbound {
		if inbound.Tags == nil {
			inbound.Tags = make(map[string]string)
		}
		inbound.Tags[mesh_proto.ClusterTag] = m.clusterName
	}
}

func (m *dataplaneManager) setGatewayClusterTag(dp *core_mesh.DataplaneResource) {
	if m.clusterName == "" || dp.Spec.Networking == nil || dp.Spec.Networking.Gateway == nil {
		return
	}
	if dp.Spec.Networking.Gateway.Tags == nil {
		dp.Spec.Networking.Gateway.Tags = make(map[string]string)
	}
	dp.Spec.Networking.Gateway.Tags[mesh_proto.ClusterTag] = m.clusterName
}
