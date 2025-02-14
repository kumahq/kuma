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
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
)

var gcLog logr.Logger

type (
	InsightType  core_model.ResourceType
	ResourceType core_model.ResourceType
	Age          time.Duration
)

type collector struct {
	rm                 manager.ResourceManager
	cleanupAge         time.Duration
	newTicker          func() *time.Ticker
	metric             prometheus.Summary
	resourcesToCleanup map[InsightType]ResourceType
}

func NewCollector(
	rm manager.ResourceManager,
	newTicker func() *time.Ticker,
	cleanupAge time.Duration,
	metrics core_metrics.Metrics,
	metricsName string,
	resourcesToCleanup map[InsightType]ResourceType,
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
			for insightType, resourceType := range d.resourcesToCleanup {
				if err := d.cleanup(ctx, now, insightType, resourceType); err != nil {
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

func (d *collector) cleanup(ctx context.Context, now time.Time, insightType InsightType, resourceType ResourceType) error {
	insights := registry.Global().MustNewList(core_model.ResourceType(insightType))
	if err := d.rm.List(ctx, insights); err != nil {
		return err
	}
	onDelete := map[core_model.ResourceKey]Age{}
	for _, item := range insights.GetItems() {
		insight := item.GetSpec().(generic.Insight)
		if insight.IsOnline() {
			continue
		}
		if s := insight.GetLastSubscription().(*mesh_proto.DiscoverySubscription); s != nil {
			if err := s.GetDisconnectTime().CheckValid(); err != nil {
				gcLog.Error(err, "unable to parse DisconnectTime", "disconnect time", s.GetDisconnectTime(), "mesh", item.GetMeta().GetMesh(), insightType, item.GetMeta().GetName())
				continue
			}
			age := now.Sub(s.GetDisconnectTime().AsTime())
			if age > d.cleanupAge {
				onDelete[core_model.MetaToResourceKey(item.GetMeta())] = Age(age)
			}
		}
	}
	for rk, age := range onDelete {
		gcLog.Info(fmt.Sprintf("deleting %s which is offline for %v", resourceType, age), "name", rk.Name, "mesh", rk.Mesh)
		resource := registry.Global().MustNewObject(core_model.ResourceType(resourceType))
		if err := d.rm.Delete(ctx, resource, store.DeleteBy(rk)); err != nil {
			gcLog.Error(err, "unable to delete", "resourceType", resourceType, "name", rk.Name, "mesh", rk.Mesh)
			continue
		}
	}
	return nil
}

func (d *collector) NeedLeaderElection() bool {
	return true
}
