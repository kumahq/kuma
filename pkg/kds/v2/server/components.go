package server

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	kds_server "github.com/kumahq/kuma/pkg/kds/server"
	reconcile_v2 "github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	"github.com/kumahq/kuma/pkg/kds/v2/util"
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
	insight bool,
	nackBackoff time.Duration,
) (Server, error) {
	hasher, cache := newKDSContext(log)
	generator := reconcile_v2.NewSnapshotGenerator(rt.ReadOnlyResourceManager(), filter, mapper)
	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "kds_delta")
	if err != nil {
		return nil, err
	}
	reconciler := reconcile_v2.NewReconciler(hasher, cache, generator, rt.Config().Mode, statsCallbacks, rt.Tenants())
	syncTracker, err := newSyncTracker(log, reconciler, refresh, rt.Metrics(), providedTypes, rt.EventBus(), rt.Config().Experimental.KDSEventBasedWatchdog)
	if err != nil {
		return nil, err
	}
	callbacks := util_xds_v3.CallbacksChain{
		NewTenancyCallbacks(rt.Tenants()),
		&typeAdjustCallbacks{},
		util_xds_v3.NewControlPlaneIdCallbacks(serverID),
		util_xds_v3.AdaptDeltaCallbacks(util_xds.LoggingCallbacks{Log: log}),
		util_xds_v3.AdaptDeltaCallbacks(statsCallbacks),
		// util_xds_v3.AdaptDeltaCallbacks(NewNackBackoff(nackBackoff)),
		newKdsRetryForcer(log, cache, hasher),
		syncTracker,
	}
	if insight {
		callbacks = append(callbacks, DefaultStatusTracker(rt, log))
	}
	return NewServer(cache, callbacks, log), nil
}

func DefaultStatusTracker(rt core_runtime.Runtime, log logr.Logger) StatusTracker {
	return NewStatusTracker(rt, func(accessor StatusAccessor, l logr.Logger) kds_server.ZoneInsightSink {
		return kds_server.NewZoneInsightSink(accessor, func() *time.Ticker {
			return time.NewTicker(rt.Config().Multizone.Global.KDS.ZoneInsightFlushInterval.Duration)
		}, func() *time.Ticker {
			return time.NewTicker(rt.Config().Metrics.Zone.IdleTimeout.Duration / 2)
		}, rt.Config().Multizone.Global.KDS.ZoneInsightFlushInterval.Duration/10, kds_server.NewZonesInsightStore(rt.ResourceManager()), l)
	}, log)
}

func newSyncTracker(
	log logr.Logger,
	reconciler reconcile_v2.Reconciler,
	refresh time.Duration,
	metrics core_metrics.Metrics,
	providedTypes []model.ResourceType,
	eventBus events.EventBus,
	experimentalWatchdogCfg kuma_cp.ExperimentalKDSEventBasedWatchdog,
) (envoy_xds.Callbacks, error) {
	kdsMetrics, err := NewMetrics(metrics)
	if err != nil {
		return nil, err
	}
	changedTypes := map[model.ResourceType]struct{}{}
	for _, typ := range providedTypes {
		changedTypes[typ] = struct{}{}
	}
	return util_xds_v3.NewWatchdogCallbacks(func(ctx context.Context, node *envoy_core.Node, streamID int64) (util_watchdog.Watchdog, error) {
		log := log.WithValues("streamID", streamID, "node", node)
		if experimentalWatchdogCfg.Enabled {
			return &EventBasedWatchdog{
				Ctx:           ctx,
				Node:          node,
				Listener:      eventBus.Subscribe(),
				Reconciler:    reconciler,
				ProvidedTypes: changedTypes,
				Metrics:       kdsMetrics,
				Log:           log,
				NewFlushTicker: func() *time.Ticker {
					return time.NewTicker(experimentalWatchdogCfg.FlushInterval.Duration)
				},
				NewFullResyncTicker: func() *time.Ticker {
					return time.NewTicker(experimentalWatchdogCfg.FullResyncInterval.Duration)
				},
			}, nil
		}
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refresh)
			},
			OnTick: func(context.Context) error {
				start := core.Now()
				log.V(1).Info("on tick")
				err, changed := reconciler.Reconcile(ctx, node, changedTypes)
				if err != nil {
					result := ResultNoChanges
					if changed {
						result = ResultChanged
					}
					kdsMetrics.KdsGenerations.WithLabelValues(ReasonResync, result).
						Observe(float64(core.Now().Sub(start).Milliseconds()))
				}
				return err
			},
			OnError: func(err error) {
				kdsMetrics.KdsGenerationsErrors.Inc()
				log.Error(err, "OnTick() failed")
			},
			OnStop: func() {
				if err := reconciler.Clear(ctx, node); err != nil {
					log.Error(err, "OnStop() failed")
				}
			},
		}, nil
	}), nil
}

func newKDSContext(log logr.Logger) (envoy_cache.NodeHash, envoy_cache.SnapshotCache) { //nolint:unparam
	hasher := hasher{}
	logger := util_xds.NewLogger(log)
	return hasher, envoy_cache.NewSnapshotCache(false, hasher, logger)
}

type hasher struct{}

func (_ hasher) ID(node *envoy_core.Node) string {
	tenantID, found := util.TenantFromMetadata(node)
	if !found {
		return node.Id
	}
	return node.Id + ":" + tenantID
}
