package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

type MeteredStore struct {
	delegate  store.ResourceStore
	metric    *prometheus.HistogramVec
	conflicts *prometheus.CounterVec
}

func NewMeteredStore(delegate store.ResourceStore, metrics core_metrics.Metrics) (*MeteredStore, error) {
	meteredStore := MeteredStore{
		delegate: delegate,
		metric: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "store",
			Help: "Summary of Store operations",
		}, []string{"operation", "resource_type"}),
		conflicts: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "store_conflicts",
			Help: "Counter of store conflicts while update",
		}, []string{"resource_type"}),
	}
	if err := metrics.BulkRegister(meteredStore.metric, meteredStore.conflicts); err != nil {
		return nil, err
	}
	return &meteredStore, nil
}

func (m *MeteredStore) Create(ctx context.Context, resource model.Resource, optionsFunc ...store.CreateOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("create", string(resource.Descriptor().Name)).Observe(core.Now().Sub(start).Seconds())
	}()
	err := m.delegate.Create(ctx, resource, optionsFunc...)
	if store.IsAlreadyExists(err) {
		m.conflicts.WithLabelValues(string(resource.Descriptor().Name)).Inc()
	}
	return err
}

func (m *MeteredStore) Update(ctx context.Context, resource model.Resource, optionsFunc ...store.UpdateOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("update", string(resource.Descriptor().Name)).Observe(core.Now().Sub(start).Seconds())
	}()
	err := m.delegate.Update(ctx, resource, optionsFunc...)
	if store.IsConflict(err) {
		m.conflicts.WithLabelValues(string(resource.Descriptor().Name)).Inc()
	}
	return err
}

func (m *MeteredStore) Delete(ctx context.Context, resource model.Resource, optionsFunc ...store.DeleteOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("delete", string(resource.Descriptor().Name)).Observe(core.Now().Sub(start).Seconds())
	}()
	return m.delegate.Delete(ctx, resource, optionsFunc...)
}

func (m *MeteredStore) Get(ctx context.Context, resource model.Resource, optionsFunc ...store.GetOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("get", string(resource.Descriptor().Name)).Observe(core.Now().Sub(start).Seconds())
	}()
	return m.delegate.Get(ctx, resource, optionsFunc...)
}

func (m *MeteredStore) List(ctx context.Context, list model.ResourceList, optionsFunc ...store.ListOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("list", string(list.GetItemType())).Observe(core.Now().Sub(start).Seconds())
	}()
	return m.delegate.List(ctx, list, optionsFunc...)
}

var _ store.ResourceStore = &MeteredStore{}
