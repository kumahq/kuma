package server

import (
	"context"
	"math/rand"
	"strings"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
	"google.golang.org/protobuf/types/known/structpb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/events"
	"github.com/kumahq/kuma/pkg/kds/status"
	reconcile_v2 "github.com/kumahq/kuma/pkg/kds/v2/reconcile"
	"github.com/kumahq/kuma/pkg/kds/v2/util"
	kuma_log "github.com/kumahq/kuma/pkg/log"
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
	filter reconcile_v2.ResourceFilter,
	mapper reconcile_v2.ResourceMapper,
	nackBackoff time.Duration,
) (Server, error) {
	hasher, cache := newKDSContext(log)
	generator := reconcile_v2.NewSnapshotGenerator(rt.ReadOnlyResourceManager(), filter, mapper)
	statsCallbacks, err := util_xds.NewStatsCallbacks(rt.Metrics(), "kds_delta", kdsVersionExtractor)
	if err != nil {
		return nil, err
	}
	reconciler := reconcile_v2.NewReconciler(hasher, cache, generator, rt.GetMode(), statsCallbacks, rt.Tenants())
	syncTracker, err := newSyncTracker(
		log,
		reconciler,
		refresh,
		rt.Metrics(),
		providedTypes,
		rt.EventBus(),
		rt.Config().Experimental.KDSEventBasedWatchdog,
		rt.Extensions(),
	)
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
		syncTracker,
		status.DefaultStatusTracker(rt, log),
	}
	return NewServer(cache, callbacks, log), nil
}

func newSyncTracker(
	log logr.Logger,
	reconciler reconcile_v2.Reconciler,
	refresh time.Duration,
	metrics core_metrics.Metrics,
	providedTypes []model.ResourceType,
	eventBus events.EventBus,
	experimentalWatchdogCfg kuma_cp.ExperimentalKDSEventBasedWatchdog,
	extensions context.Context,
) (envoy_xds.Callbacks, error) {
	kdsMetrics, err := NewMetrics(metrics)
	if err != nil {
		return nil, err
	}
	changedTypes := map[model.ResourceType]struct{}{}
	for _, typ := range providedTypes {
		changedTypes[typ] = struct{}{}
	}
	return util_xds_v3.NewWatchdogCallbacks(func(ctx context.Context, node *envoy_core.Node, streamID int64) (util_xds_v3.Watchdog, error) {
		log := log.WithValues("streamID", streamID, "nodeID", node.Id)
		log = kuma_log.AddFieldsFromCtx(log, ctx, extensions)
		if experimentalWatchdogCfg.Enabled {
			return &EventBasedWatchdog{
				Node:          node,
				EventBus:      eventBus,
				Reconciler:    reconciler,
				ProvidedTypes: changedTypes,
				Metrics:       kdsMetrics,
				Log:           log,
				NewFlushTicker: func() *time.Ticker {
					return time.NewTicker(experimentalWatchdogCfg.FlushInterval.Duration)
				},
				NewFullResyncTicker: func() *time.Ticker {
					if experimentalWatchdogCfg.DelayFullResync {
						// To ensure an even distribution of connections over time, we introduce a random delay within
						// the full resync interval. This prevents clustering connections within a short timeframe
						// and spreads them evenly across the entire interval. After the initial trigger, we reset
						// the ticker, returning it to its full resync interval.
						// #nosec G404 - math rand is enough
						delay := time.Duration(experimentalWatchdogCfg.FullResyncInterval.Duration.Seconds()*rand.Float64()) * time.Second
						ticker := time.NewTicker(experimentalWatchdogCfg.FullResyncInterval.Duration + delay)
						go func() {
							<-time.After(delay)
							ticker.Reset(experimentalWatchdogCfg.FullResyncInterval.Duration)
						}()
						return ticker
					} else {
						return time.NewTicker(experimentalWatchdogCfg.FullResyncInterval.Duration)
					}
				},
			}, nil
		}
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refresh)
			},
			OnTick: func(ctx context.Context) error {
				start := core.Now()
				log.V(1).Info("on tick")
				err, changed := reconciler.Reconcile(ctx, node, changedTypes, log)
				if err == nil {
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
				kdsMetrics.KdsGenerationErrors.Inc()
				log.Error(err, "OnTick() failed")
			},
			OnStop: func() {
				if err := reconciler.Clear(node); err != nil {
					log.Error(err, "OnStop() failed")
				}
			},
		}, nil
	}), nil
}

func kdsVersionExtractor(metadata *structpb.Struct) string {
	version := system_proto.NewVersion()
	if err := status.ReadVersion(metadata, version); err != nil {
		return "unknown"
	}
	ver := version.GetKumaCp().GetVersion()
	if strings.Contains(ver, "preview") {
		return "preview" // avoid high cardinality metrics
	}
	return ver
}

func newKDSContext(log logr.Logger) (envoy_cache.NodeHash, envoy_cache.SnapshotCache) { //nolint:unparam
	hasher := Hasher{}
	logger := util_xds.NewLogger(log)
	return hasher, envoy_cache.NewSnapshotCache(false, hasher, logger)
}

type Hasher struct{}

func (_ Hasher) ID(node *envoy_core.Node) string {
	tenantID, found := util.TenantFromMetadata(node)
	if !found {
		return node.Id
	}
	return node.Id + ":" + tenantID
}
