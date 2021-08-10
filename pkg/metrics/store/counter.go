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
	"github.com/kumahq/kuma/pkg/metrics"
)

var log = core.Log.WithName("metrics").WithName("store-counter")

type storeCounter struct {
	resManager manager.ReadOnlyResourceManager
	counts     *prometheus.GaugeVec
}

var _ component.Component = &storeCounter{}

func NewStoreCounter(resManager manager.ReadOnlyResourceManager, metrics metrics.Metrics) (*storeCounter, error) {
	counts := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resources_count",
	}, []string{"resource_type"})

	if err := metrics.Register(counts); err != nil {
		return nil, err
	}

	return &storeCounter{
		resManager: resManager,
		counts:     counts,
	}, nil
}

func (s *storeCounter) Start(stop <-chan struct{}) error {
	ticker := time.NewTicker(1 * time.Minute)
	return s.StartWithTicker(stop, ticker)
}

func (s *storeCounter) StartWithTicker(stop <-chan struct{}, ticker *time.Ticker) error {
	defer ticker.Stop()

	log.Info("starting the resource counter")
	if err := s.count(); err != nil {
		log.Error(err, "unable to count resources")
	}
	for {
		select {
		case <-ticker.C:
			if err := s.count(); err != nil {
				log.Error(err, "unable to count resources")
			}
		case <-stop:
			log.Info("stopping")
			return nil
		}
	}
}

func (s *storeCounter) NeedLeaderElection() bool {
	return true
}

func (s *storeCounter) count() error {
	resourceCount := map[string]uint32{}
	if err := s.countGlobalScopedResources(resourceCount); err != nil {
		return err
	}
	if err := s.countMeshScopedResources(resourceCount); err != nil {
		return err
	}
	for resType, counter := range resourceCount {
		s.counts.WithLabelValues(resType).Set(float64(counter))
	}
	return nil
}

func (s *storeCounter) countGlobalScopedResources(resourceCount map[string]uint32) error {
	for _, resDesc := range registry.Global().ObjectDescriptors() {
		if resDesc.Scope == model.ScopeMesh {
			continue
		}
		list := resDesc.NewList()
		if err := s.resManager.List(context.Background(), list); err != nil {
			return err
		}
		resourceCount[string(resDesc.Name)] += uint32(len(list.GetItems()))
	}
	return nil
}

func (s *storeCounter) countMeshScopedResources(resourceCount map[string]uint32) error {
	insights := &mesh.MeshInsightResourceList{}
	if err := s.resManager.List(context.Background(), insights); err != nil {
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
