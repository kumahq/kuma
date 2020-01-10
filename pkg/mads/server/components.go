package server

import (
	"time"

	"github.com/go-logr/logr"

	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	mads_reconcile "github.com/Kong/kuma/pkg/mads/reconcile"
	util_watchdog "github.com/Kong/kuma/pkg/util/watchdog"
	util_xds "github.com/Kong/kuma/pkg/util/xds"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

func NewSnapshotGenerator() mads_reconcile.SnapshotGenerator {
	return mads_reconcile.NewSnapshotGenerator()
}

func NewVersioner() util_xds.SnapshotVersioner {
	return util_xds.SnapshotAutoVersioner{UUID: core.NewUUID}
}

func NewReconciler(hasher envoy_cache.NodeHash, cache util_xds.SnapshotCache,
	generator mads_reconcile.SnapshotGenerator, versioner util_xds.SnapshotVersioner) mads_reconcile.Reconciler {
	return mads_reconcile.NewReconciler(hasher, cache, generator, versioner)
}

func NewSyncTracker(rt core_runtime.Runtime, reconciler mads_reconcile.Reconciler) envoy_xds.Callbacks {
	return util_xds.NewWatchdogCallbacks(func(node *envoy_core.Node, _ int64) (util_watchdog.Watchdog, error) {
		log := madsServerLog.WithValues("node", node)
		return &util_watchdog.SimpleWatchdog{
			NewTicker: func() *time.Ticker {
				return time.NewTicker(rt.Config().MonitoringAssignmentServer.AssignmentRefreshInterval)
			},
			OnTick: func() error {
				log.V(1).Info("on tick")
				return reconciler.Reconcile(node)
			},
			OnError: func(err error) {
				log.Error(err, "OnTick() failed")
			},
		}, nil
	})
}

func NewXdsContext(log logr.Logger) (envoy_cache.NodeHash, util_xds.SnapshotCache) {
	hasher := hasher{}
	logger := util_xds.NewLogger(log)
	return hasher, util_xds.NewSnapshotCache(false, hasher, logger)
}

type hasher struct {
}

func (_ hasher) ID(node *envoy_core.Node) string {
	// in the very first implementation, we don't differentiate clients
	return ""
}
