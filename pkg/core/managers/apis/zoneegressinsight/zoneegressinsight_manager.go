package zoneegressinsight

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewZoneEgressInsightManager(store core_store.ResourceStore, config *kuma_cp.DataplaneMetrics) core_manager.ResourceManager {
	return &zoneEgressInsightManager{
		ResourceManager: core_manager.NewResourceManager(store),
		store:           store,
		config:          config,
	}
}

type zoneEgressInsightManager struct {
	core_manager.ResourceManager
	store  core_store.ResourceStore
	config *kuma_cp.DataplaneMetrics
}

func (m *zoneEgressInsightManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	if err := resource.Validate(); err != nil {
		return err
	}
	opts := core_store.NewCreateOptions(fs...)

	m.limitSubscription(resource.(*mesh.ZoneEgressInsightResource))

	zoneEgress := mesh.NewZoneEgressResource()
	if err := m.store.Get(ctx, zoneEgress, core_store.GetByKey(opts.Name, core_model.NoMesh)); err != nil {
		return err
	}
	return m.store.Create(ctx, resource, append(fs, core_store.CreatedAt(core.Now()), core_store.CreateWithOwner(zoneEgress))...)
}

func (m *zoneEgressInsightManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	m.limitSubscription(resource.(*mesh.ZoneEgressInsightResource))
	return m.ResourceManager.Update(ctx, resource, fs...)
}

func (m *zoneEgressInsightManager) limitSubscription(zoneEgressInsight *mesh.ZoneEgressInsightResource) {
	if m.config.SubscriptionLimit == 0 {
		zoneEgressInsight.Spec.Subscriptions = []*mesh_proto.DiscoverySubscription{}
		return
	}
	if len(zoneEgressInsight.Spec.Subscriptions) <= m.config.SubscriptionLimit {
		return
	}
	s := zoneEgressInsight.Spec.Subscriptions
	zoneEgressInsight.Spec.Subscriptions = s[len(s)-m.config.SubscriptionLimit:]
}
