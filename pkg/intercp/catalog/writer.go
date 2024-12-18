package catalog

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/multitenant"
)

var writerLog = core.Log.WithName("intercp").WithName("catalog").WithName("writer")

type catalogWriter struct {
	catalog    Catalog
	heartbeats *Heartbeats
	instance   Instance
	interval   time.Duration
	metric     prometheus.Summary
}

var _ component.Component = &catalogWriter{}

func NewWriter(
	catalog Catalog,
	heartbeats *Heartbeats,
	instance Instance,
	interval time.Duration,
	metrics core_metrics.Metrics,
) (component.Component, error) {
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "component_catalog_writer",
		Help:       "Summary of Inter CP Catalog Writer component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}

	leaderInstance := instance
	leaderInstance.Leader = true
	return &catalogWriter{
		catalog:    catalog,
		heartbeats: heartbeats,
		instance:   leaderInstance,
		interval:   interval,
		metric:     metric,
	}, nil
}

func (r *catalogWriter) Start(stop <-chan struct{}) error {
	writerLog.Info("starting catalog writer")
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ctx = multitenant.WithTenant(ctx, multitenant.GlobalTenantID)
	ctx, cancelFn := context.WithCancel(ctx)
	writerLog.Info("replacing a leader in the catalog")
	if err := r.catalog.ReplaceLeader(ctx, r.instance); err != nil {
		writerLog.Error(err, "could not replace leader") // continue, it will be replaced in ticker anyways
	}
	ticker := time.NewTicker(r.interval)
	for {
		select {
		case <-ticker.C:
			start := core.Now()
			instances := r.heartbeats.ResetAndCollect()
			instances = append(instances, r.instance)
			updated, err := r.catalog.Replace(ctx, instances)
			if err != nil {
				writerLog.Error(err, "could not update catalog")
				continue
			}
			if updated {
				writerLog.Info("instances catalog updated", "instances", instances)
			} else {
				writerLog.V(1).Info("no need to update instances, because the catalog is the same", "instances", instances)
			}
			r.metric.Observe(float64(core.Now().Sub(start).Milliseconds()))
		case <-stop:
			cancelFn()
			if err := r.catalog.DropLeader(context.WithoutCancel(ctx), r.instance); err != nil {
				writerLog.Info("could not drop leader, it will be replaced by the next leader", "err", err)
			} else {
				writerLog.Info("leader dropped")
			}
			return nil
		}
	}
}

func (r *catalogWriter) NeedLeaderElection() bool {
	return true
}
