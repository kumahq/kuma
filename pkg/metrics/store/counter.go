package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
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
	for _, resType := range registry.Global().ListTypes() {
		list, err := registry.Global().NewList(resType)
		if err != nil {
			return err
		}
		if err := s.resManager.List(context.Background(), list); err != nil {
			return err
		}
		s.counts.WithLabelValues(string(resType)).Set(float64(len(list.GetItems())))
	}
	return nil
}
