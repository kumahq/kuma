package dataplaneinsight

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

func NewDataplaneInsightManager(store core_store.ResourceStore, config *kuma_cp.DataplaneMetrics) core_manager.ResourceManager {
	return &dataplaneInsightManager{
		ResourceManager: core_manager.NewResourceManager(store),
		store:           store,
		config:          config,
	}
}

type dataplaneInsightManager struct {
	core_manager.ResourceManager
	store  core_store.ResourceStore
	config *kuma_cp.DataplaneMetrics
}

func (m *dataplaneInsightManager) Create(ctx context.Context, resource core_model.Resource, fs ...core_store.CreateOptionsFunc) error {
	if err := resource.Validate(); err != nil {
		return err
	}
	opts := core_store.NewCreateOptions(fs...)

	m.limitSubscription(resource.(*core_mesh.DataplaneInsightResource))

	dp := core_mesh.NewDataplaneResource()
	if err := m.store.Get(ctx, dp, core_store.GetByKey(opts.Name, opts.Mesh)); err != nil {
		return err
	}
	return m.store.Create(ctx, resource, append(fs, core_store.CreatedAt(core.Now()), core_store.CreateWithOwner(dp))...)
}

func (m *dataplaneInsightManager) Update(ctx context.Context, resource core_model.Resource, fs ...core_store.UpdateOptionsFunc) error {
	m.limitSubscription(resource.(*core_mesh.DataplaneInsightResource))
	return m.ResourceManager.Update(ctx, resource, fs...)
}

func (m *dataplaneInsightManager) limitSubscription(dpInsight *core_mesh.DataplaneInsightResource) {
	if m.config.SubscriptionLimit == 0 {
		dpInsight.Spec.Subscriptions = []*mesh_proto.DiscoverySubscription{}
		return
	}
	if len(dpInsight.Spec.Subscriptions) <= m.config.SubscriptionLimit {
		return
	}
	s := dpInsight.Spec.Subscriptions
	dpInsight.Spec.Subscriptions = s[len(s)-m.config.SubscriptionLimit:]
}
