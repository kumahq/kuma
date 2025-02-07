package gc

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/api/generic"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

var gcLog logr.Logger

type InsightToResource struct {
	Insight  core_model.ResourceType
	Resource core_model.ResourceType
}

type collector struct {
	rm                 manager.ResourceManager
	cleanupAge         time.Duration
	newTicker          func() *time.Ticker
	metric             prometheus.Summary
	resourcesToCleanup []InsightToResource
}

func NewCollector(
	rm manager.ResourceManager,
	newTicker func() *time.Ticker,
	cleanupAge time.Duration,
	metrics core_metrics.Metrics,
	metricsName string,
	resourcesToCleanup []InsightToResource,
) (component.Component, error) {
	gcLog = core.Log.WithName(fmt.Sprintf("%s-gc", metricsName))
	metric := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       fmt.Sprintf("component_%s_gc", metricsName),
		Help:       "Summary of Dataplane GC component interval",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(metric); err != nil {
		return nil, err
	}
	return &collector{
		cleanupAge:         cleanupAge,
		rm:                 rm,
		newTicker:          newTicker,
		metric:             metric,
		resourcesToCleanup: resourcesToCleanup,
	}, nil
}

func (d *collector) Start(stop <-chan struct{}) error {
	ticker := d.newTicker()
	defer ticker.Stop()
	gcLog.Info("started")
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	for {
		select {
		case now := <-ticker.C:
			start := core.Now()
			for _, res := range d.resourcesToCleanup {
				if err := d.cleanup(ctx, now, res); err != nil {
					gcLog.Error(err, "unable to cleanup")
					continue
				}
			}
			d.metric.Observe(float64(core.Now().Sub(start).Milliseconds()))
		case <-stop:
			gcLog.Info("stopped")
			return nil
		}
	}
}

func (d *collector) cleanup(ctx context.Context, now time.Time, res InsightToResource) error {
	insights := registry.Global().MustNewList(res.Insight)
	if err := d.rm.List(ctx, insights); err != nil {
		return err
	}
	onDelete := []model.ResourceKey{}
	for _, item := range insights.GetItems() {
		insight := item.GetSpec().(generic.Insight)
		if insight.IsOnline() {
			continue
		}
		if s := insight.GetLastSubscription().(*mesh_proto.DiscoverySubscription); s != nil {
			if err := s.GetDisconnectTime().CheckValid(); err != nil {
				gcLog.Error(err, "unable to parse DisconnectTime", "disconnect time", s.GetDisconnectTime(), "mesh", item.GetMeta().GetMesh(), res.Insight, item.GetMeta().GetName())
				continue
			}
			if now.Sub(s.GetDisconnectTime().AsTime()) > d.cleanupAge {
				onDelete = append(onDelete, model.ResourceKey{Name: item.GetMeta().GetName(), Mesh: item.GetMeta().GetMesh()})
			}
		}
	}
	for _, rk := range onDelete {
		gcLog.Info(fmt.Sprintf("deleting %s which is offline for %v", res.Resource, d.cleanupAge), "name", rk.Name, "mesh", rk.Mesh)
		resource := registry.Global().MustNewObject(res.Resource)
		if err := d.rm.Delete(ctx, resource, store.DeleteBy(rk)); err != nil {
			gcLog.Error(err, "unable to delete", "resourceType", res.Resource, "name", rk.Name, "mesh", rk.Mesh)
			continue
		}
	}
	return nil
}

func (d *collector) NeedLeaderElection() bool {
	return true
}
