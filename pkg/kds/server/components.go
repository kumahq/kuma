package server

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core/resources/model"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/kds/reconcile"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
)

func New(log logr.Logger, rt core_runtime.Runtime, providedTypes []model.ResourceType, serverID string, refresh time.Duration, filter reconcile.ResourceFilter, insight bool) (Server, error) {
	hasher, cache := newKDSContext(log)
	generator := reconcile.NewSnapshotGenerator(rt.ReadOnlyResourceManager(), providedTypes, filter)
	versioner := util_xds.SnapshotAutoVersioner{UUID: core.NewUUID}
	reconciler := reconcile.NewReconciler(hasher, cache, generator, versioner)
	syncTracker := newSyncTracker(log, reconciler, refresh)
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: log},
		syncTracker,
	}
	if insight {
		callbacks = append(callbacks, DefaultStatusTracker(rt, log))
	}
	return NewServer(cache, callbacks, log, serverID), nil
}

func DefaultStatusTracker(rt core_runtime.Runtime, log logr.Logger) StatusTracker {
	return NewStatusTracker(rt, func(accessor StatusAccessor, l logr.Logger) ZoneInsightSink {
		return NewZoneInsightSink(
			accessor,
			func() *time.Ticker {
				return time.NewTicker(1 * time.Second)
			},
			NewDataplaneInsightStore(rt.ResourceManager()),
			l)
	}, log)
}

func newSyncTracker(log logr.Logger, reconciler reconcile.Reconciler, refresh time.Duration) envoy_xds.Callbacks {
	return util_xds.NewWatchdogCallbacks(func(ctx context.Context, node *envoy_core.Node, streamID int64) (util_watchdog.Watchdog, error) {
		log := log.WithValues("streamID", streamID, "node", node)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(refresh)
			},
			OnTick: func() error {
				log.V(1).Info("on tick")
				return reconciler.Reconcile(ctx, node)
			},
			OnError: func(err error) {
				log.Error(err, "OnTick() failed")
			},
		}, nil
	})
}

func newKDSContext(log logr.Logger) (envoy_cache.NodeHash, util_xds.SnapshotCache) {
	hasher := hasher{}
	logger := util_xds.NewLogger(log)
	return hasher, util_xds.NewSnapshotCache(false, hasher, logger)
}

type hasher struct {
}

func (_ hasher) ID(node *envoy_core.Node) string {
	return node.Id
}
