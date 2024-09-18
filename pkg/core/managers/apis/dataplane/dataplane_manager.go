package dataplane

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewDataplaneManager(
	store core_store.ResourceStore,
	zone string,
	mode string,
	isK8s bool,
	systemNamespace string,
	validator Validator,
) core_manager.ResourceManager {
	return &dataplaneManager{
		ResourceManager: core_manager.NewResourceManager(store),
		store:           store,
		zone:            zone,
		mode:            mode,
		isK8s:           isK8s,
		systemNamespace: systemNamespace,
		validator:       validator,
	}
}

type dataplaneManager struct {
	core_manager.ResourceManager
	store           core_store.ResourceStore
	zone            string
	mode            string
	isK8s           bool
	systemNamespace string
	validator       Validator
}

func (m *dataplaneManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	if err := core_model.Validate(resource); err != nil {
		return err
	}
	dp, err := m.dataplane(resource)
	if err != nil {
		return err
	}

	opts := core_store.NewCreateOptions(fs...)
	m.setInboundsClusterTag(dp)
	m.setGatewayClusterTag(dp)
	m.setHealth(dp)
	labels, err := core_model.ComputeLabels(
		resource.Descriptor(),
		resource.GetSpec(),
		opts.Labels,
		model.ResourceNameExtensions{},
		opts.Mesh,
		m.mode,
		m.isK8s,
		m.systemNamespace,
		m.zone,
	)
	if err != nil {
		return err
	}
	fs = append(fs, core_store.CreateWithLabels(labels))

	owner := core_mesh.NewMeshResource()
	if err := m.store.Get(ctx, owner, core_store.GetByKey(opts.Mesh, core_model.NoMesh)); err != nil {
		return core_manager.MeshNotFound(opts.Mesh)
	}

	key := core_model.ResourceKey{
		Mesh: opts.Mesh,
		Name: opts.Name,
	}
	if err := m.validator.ValidateCreate(ctx, key, dp, owner); err != nil {
		return err
	}

	return m.store.Create(ctx, resource, append(fs, core_store.CreatedAt(core.Now()))...)
}

func (m *dataplaneManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	dp, err := m.dataplane(resource)
	if err != nil {
		return err
	}

	m.setInboundsClusterTag(dp)
	m.setGatewayClusterTag(dp)
	labels, err := core_model.ComputeLabels(
		resource.Descriptor(),
		resource.GetSpec(),
		resource.GetMeta().GetLabels(),
		resource.GetMeta().GetNameExtensions(),
		resource.GetMeta().GetMesh(),
		m.mode,
		m.isK8s,
		m.systemNamespace,
		m.zone,
	)
	if err != nil {
		return err
	}
	fs = append(fs, core_store.UpdateWithLabels(labels))

	owner := core_mesh.NewMeshResource()
	if err := m.store.Get(ctx, owner, core_store.GetByKey(resource.GetMeta().GetMesh(), core_model.NoMesh)); err != nil {
		return core_manager.MeshNotFound(resource.GetMeta().GetMesh())
	}
	if err := m.validator.ValidateUpdate(ctx, dp, owner); err != nil {
		return err
	}

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
	if m.zone == "" || dp.Spec.Networking == nil {
		return
	}

	for _, inbound := range dp.Spec.Networking.Inbound {
		if inbound.Tags == nil {
			inbound.Tags = make(map[string]string)
		}
		inbound.Tags[mesh_proto.ZoneTag] = m.zone
	}
}

func (m *dataplaneManager) setGatewayClusterTag(dp *core_mesh.DataplaneResource) {
	if m.zone == "" || dp.Spec.GetNetworking().GetGateway() == nil {
		return
	}
	if dp.Spec.Networking.Gateway.Tags == nil {
		dp.Spec.Networking.Gateway.Tags = make(map[string]string)
	}
	dp.Spec.Networking.Gateway.Tags[mesh_proto.ZoneTag] = m.zone
}

func (m *dataplaneManager) setHealth(dp *core_mesh.DataplaneResource) {
	for _, inbound := range dp.Spec.Networking.Inbound {
		if inbound.ServiceProbe != nil {
			inbound.State = mesh_proto.Dataplane_Networking_Inbound_NotReady
			// write health for backwards compatibility with Kuma 2.5 and older
			inbound.Health = &mesh_proto.Dataplane_Networking_Inbound_Health{
				Ready: false,
			}
		}
	}
}
