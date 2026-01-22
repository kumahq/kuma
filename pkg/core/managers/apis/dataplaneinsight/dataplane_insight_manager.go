package dataplaneinsight

import (
	"context"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/v2/pkg/core"
	core_mesh "github.com/kumahq/kuma/v2/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/v2/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core/resources/validator"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
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
	if err := validator.Validate(resource); err != nil {
		return err
	}
	opts := core_store.NewCreateOptions(fs...)

	m.limitSubscription(resource.(*core_mesh.DataplaneInsightResource))

	dp := core_mesh.NewDataplaneResource()
	if err := m.store.Get(ctx, dp, core_store.GetByKey(opts.Name, opts.Mesh)); err != nil {
		return err
	}
	createOpts := []core_store.CreateOptionsFunc{core_store.CreatedAt(core.Now()), core_store.CreateWithOwner(dp)}
	if workload, ok := dp.GetMeta().GetLabels()[metadata.KumaWorkload]; ok {
		createOpts = append(createOpts, core_store.CreateWithLabels(map[string]string{metadata.KumaWorkload: workload}))
	}
	return m.store.Create(ctx, resource, append(fs, createOpts...)...)
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
