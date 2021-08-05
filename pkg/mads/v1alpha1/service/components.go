package service

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/core"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	mads_generator "github.com/kumahq/kuma/pkg/mads/v1alpha1/generator"
	mads_reconcile "github.com/kumahq/kuma/pkg/mads/v1alpha1/reconcile"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
)

func NewSnapshotGenerator(rm core_manager.ReadOnlyResourceManager) mads_reconcile.SnapshotGenerator {
	return mads_reconcile.NewSnapshotGenerator(rm, mads_generator.MonitoringAssignmentsGenerator{})
}

func NewVersioner() util_xds.SnapshotVersioner {
	return util_xds.SnapshotAutoVersioner{UUID: core.NewUUID}
}

func NewReconciler(hasher envoy_cache.NodeHash, cache util_xds.SnapshotCache,
	generator mads_reconcile.SnapshotGenerator, versioner util_xds.SnapshotVersioner) mads_reconcile.Reconciler {
	return mads_reconcile.NewReconciler(hasher, cache, generator, versioner)
}

func NewSyncTracker(reconciler mads_reconcile.Reconciler, refresh time.Duration, log logr.Logger) envoy_xds.Callbacks {
	return util_xds_v2.NewWatchdogCallbacks(func(ctx context.Context, node *envoy_core.Node, streamID int64) (util_watchdog.Watchdog, error) {
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
