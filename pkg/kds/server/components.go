package server

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	"github.com/kumahq/kuma/pkg/kds/status"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func New(
	log logr.Logger,
	rt core_runtime.Runtime,
	providedTypes []model.ResourceType,
	serverID string,
	refresh time.Duration,
	filter reconcile.ResourceFilter,
	mapper reconcile.ResourceMapper,
	nackBackoff time.Duration,
) (Server, error) {
	hashFn, cache := newKDSContext(log)
	generator := reconcile.NewSnapshotGenerator(rt.ReadOnlyResourceManager(), providedTypes, filter, mapper)
	versioner := util_xds_v3.SnapshotAutoVersioner{UUID: core.NewUUID}
	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "kds")
	if err != nil {
		return nil, err
	}
	reconciler := reconcile.NewReconciler(hashFn, cache, generator, versioner, rt.Config().Mode, statsCallbacks)
	syncTracker, err := newSyncTracker(log, reconciler, refresh, rt.Metrics())
	if err != nil {
		return nil, err
	}
	callbacks := util_xds_v3.CallbacksChain{
		&typeAdjustCallbacks{},
		util_xds_v3.NewControlPlaneIdCallbacks(serverID),
		util_xds_v3.AdaptCallbacks(util_xds.LoggingCallbacks{Log: log}),
		util_xds_v3.AdaptCallbacks(statsCallbacks),
		// util_xds_v3.AdaptCallbacks(NewNackBackoff(nackBackoff)), todo(jakubdyszkiewicz) temporarily disable to see if it's a reason for E2E flakes.
		syncTracker,
		status.DefaultStatusTracker(rt, log),
	}
	return NewServer(cache, callbacks, log), nil
}

func newSyncTracker(log logr.Logger, reconciler reconcile.Reconciler, refresh time.Duration, metrics core_metrics.Metrics) (envoy_xds.Callbacks, error) {
	kdsGenerations := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:       "kds_generation",
		Help:       "Summary of KDS Snapshot generation",
		Objectives: core_metrics.DefaultObjectives,
	})
	if err := metrics.Register(kdsGenerations); err != nil {
		return nil, err
	}
	kdsGenerationsErrors := prometheus.NewCounter(prometheus.CounterOpts{
		Help: "Counter of errors during KDS generation",
		Name: "kds_generation_errors",
	})
	if err := metrics.Register(kdsGenerationsErrors); err != nil {
		return nil, err
	}
	return util_xds_v3.NewWatchdogCallbacks(func(ctx context.Context, node *envoy_core.Node, streamID int64) (util_watchdog.Watchdog, error) {
		log := log.WithValues("streamID", streamID, "node", node)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refresh)
			},
			OnTick: func(context.Context) error {
				start := core.Now()
				defer func() {
					kdsGenerations.Observe(float64(core.Now().Sub(start).Milliseconds()))
				}()
				log.V(1).Info("on tick")
				return reconciler.Reconcile(ctx, node)
			},
			OnError: func(err error) {
				kdsGenerationsErrors.Inc()
				log.Error(err, "OnTick() failed")
			},
			OnStop: func() {
				reconciler.Clear(ctx, node)
			},
		}, nil
	}), nil
}

func newKDSContext(log logr.Logger) (envoy_cache.NodeHash, util_xds_v3.SnapshotCache) { //nolint:unparam
	hashFn := util_xds_v3.IDHash{}
	logger := util_xds.NewLogger(log)
	return hashFn, util_xds_v3.NewSnapshotCache(false, hashFn, logger)
}
