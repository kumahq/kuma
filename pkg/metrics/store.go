package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type MeteredStore struct {
	delegate store.ResourceStore
	metric   *prometheus.SummaryVec
}

func NewMeteredStore(delegate store.ResourceStore, metrics Metrics) (*MeteredStore, error) {
	meteredStore := MeteredStore{
		delegate: delegate,
		metric: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       "store",
			Help:       "Summary of Store operations",
			Objectives: DefaultObjectives,
		}, []string{"operation", "resource_type"}),
	}
	if err := metrics.Register(meteredStore.metric); err != nil {
		return nil, err
	}
	return &meteredStore, nil
}

func (m *MeteredStore) Create(ctx context.Context, resource model.Resource, optionsFunc ...store.CreateOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("create", string(resource.GetType())).Observe(float64(core.Now().Sub(start).Milliseconds()))
	}()
	return m.delegate.Create(ctx, resource, optionsFunc...)
}

func (m *MeteredStore) Update(ctx context.Context, resource model.Resource, optionsFunc ...store.UpdateOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("update", string(resource.GetType())).Observe(float64(core.Now().Sub(start).Milliseconds()))
	}()
	return m.delegate.Update(ctx, resource, optionsFunc...)
}

func (m *MeteredStore) Delete(ctx context.Context, resource model.Resource, optionsFunc ...store.DeleteOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("delete", string(resource.GetType())).Observe(float64(core.Now().Sub(start).Milliseconds()))
	}()
	return m.delegate.Delete(ctx, resource, optionsFunc...)
}

func (m *MeteredStore) Get(ctx context.Context, resource model.Resource, optionsFunc ...store.GetOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("get", string(resource.GetType())).Observe(float64(core.Now().Sub(start).Milliseconds()))
	}()
	return m.delegate.Get(ctx, resource, optionsFunc...)
}

func (m *MeteredStore) List(ctx context.Context, list model.ResourceList, optionsFunc ...store.ListOptionsFunc) error {
	start := core.Now()
	defer func() {
		m.metric.WithLabelValues("list", string(list.GetItemType())).Observe(float64(core.Now().Sub(start).Milliseconds()))
	}()
	return m.delegate.List(ctx, list, optionsFunc...)
}

var _ store.ResourceStore = &MeteredStore{}
