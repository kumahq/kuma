package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
)

var log = core.Log.WithName("metrics").WithName("store-counter")

type storeCounter struct {
	resManager manager.ReadOnlyResourceManager
	counts     *prometheus.GaugeVec
	tenants    multitenant.Tenants
	metric     prometheus.Summary
}

var _ component.Component = &storeCounter{}

func NewStoreCounter(resManager manager.ReadOnlyResourceManager, metrics core_metrics.Metrics, tenants multitenant.Tenants) (*storeCounter, error) {
	counts := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resources_count",
	}, []string{"resource_type", "tenant"})

	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_store_counter",
		Help:       "Summary of Store Counter component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.BulkRegister(counts, metric); err != nil {
		return nil, err
	}

	return &storeCounter{
		resManager: resManager,
		counts:     counts,
		tenants:    tenants,
		metric:     metric,
	}, nil
}

func (s *storeCounter) Start(stop <-chan struct{}) error {
	ticker := time.NewTicker(1 * time.Minute)
	return s.StartWithTicker(stop, ticker)
}

func (s *storeCounter) StartWithTicker(stop <-chan struct{}, ticker *time.Ticker) error {
	defer ticker.Stop()
	ctx := user.Ctx(context.Background(), user.ControlPlane)

	log.Info("starting the resource counter")
	if err := s.countForAllTenants(ctx); err != nil {
		log.Error(err, "unable to count resources")
	}
	for {
		select {
		case <-ticker.C:
			start := core.Now()
			if err := s.countForAllTenants(ctx); err != nil {
				log.Error(err, "unable to count resources")
				continue
			}
			s.metric.Observe(float64(core.Now().Sub(start).Milliseconds()))
		case <-stop:
			log.Info("stopping")
			return nil
		}
	}
}

func (s *storeCounter) NeedLeaderElection() bool {
	return true
}

func (s *storeCounter) countForAllTenants(ctx context.Context) error {
	tenantIDs, err := s.tenants.GetIDs(ctx)
	if err != nil {
		return err
	}
	for _, tenantID := range tenantIDs {
		if err := s.count(multitenant.WithTenant(ctx, tenantID), tenantID); err != nil {
			return err
		}
	}
	return nil
}

func (s *storeCounter) count(ctx context.Context, tenantID string) error {
	resourceCount := map[string]uint32{}
	if err := s.countGlobalScopedResources(ctx, resourceCount); err != nil {
		return err
	}
	if err := s.countMeshScopedResources(ctx, resourceCount); err != nil {
		return err
	}
	for resType, counter := range resourceCount {
		s.counts.WithLabelValues(resType, tenantID).Set(float64(counter))
	}
	return nil
}

func (s *storeCounter) countGlobalScopedResources(ctx context.Context, resourceCount map[string]uint32) error {
	for _, resDesc := range registry.Global().ObjectDescriptors() {
		if resDesc.Scope == model.ScopeMesh {
			continue
		}
		list := resDesc.NewList()
		if err := s.resManager.List(ctx, list); err != nil {
			return err
		}
		resourceCount[string(resDesc.Name)] += uint32(len(list.GetItems()))
	}
	return nil
}

func (s *storeCounter) countMeshScopedResources(ctx context.Context, resourceCount map[string]uint32) error {
	insights := &mesh.MeshInsightResourceList{}
	if err := s.resManager.List(ctx, insights); err != nil {
		return err
	}
	for _, meshInsight := range insights.Items {
		resourceCount[string(mesh.DataplaneType)] += meshInsight.Spec.GetDataplanes().GetTotal()
		for policy, stats := range meshInsight.Spec.GetPolicies() {
			resourceCount[policy] += stats.GetTotal()
		}
	}
	return nil
}
