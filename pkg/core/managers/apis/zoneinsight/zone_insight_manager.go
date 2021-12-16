package zoneinsight

import (
	"context"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewZoneInsightManager(store core_store.ResourceStore, config *kuma_cp.ZoneMetrics) core_manager.ResourceManager {
	return &zoneInsightManager{
		ResourceManager: core_manager.NewResourceManager(store),
		store:           store,
		config:          config,
	}
}

type zoneInsightManager struct {
	core_manager.ResourceManager
	store  core_store.ResourceStore
	config *kuma_cp.ZoneMetrics
}

func (m *zoneInsightManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	if err := resource.Validate(); err != nil {
		return err
	}
	opts := core_store.NewCreateOptions(fs...)

	m.limitSubscription(resource.(*system.ZoneInsightResource))

	zone := system.NewZoneResource()
	if err := m.store.Get(ctx, zone, core_store.GetByKey(opts.Name, core_model.NoMesh)); err != nil {
		return err
	}
	return m.store.Create(ctx, resource, append(fs, core_store.CreatedAt(core.Now()), core_store.CreateWithOwner(zone))...)
}

func (m *zoneInsightManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	m.limitSubscription(resource.(*system.ZoneInsightResource))
	return m.ResourceManager.Update(ctx, resource, fs...)
}

func (m *zoneInsightManager) limitSubscription(zoneInsight *system.ZoneInsightResource) {
	if m.config.SubscriptionLimit == 0 {
		zoneInsight.Spec.Subscriptions = []*system_proto.KDSSubscription{}
		return
	}
	if len(zoneInsight.Spec.Subscriptions) <= m.config.SubscriptionLimit {
		return
	}
	s := zoneInsight.Spec.Subscriptions
	zoneInsight.Spec.Subscriptions = s[len(s)-m.config.SubscriptionLimit:]
}
