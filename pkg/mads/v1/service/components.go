package service

import (
	"context"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	mads_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	mads_reconcile "github.com/kumahq/kuma/pkg/mads/v1/reconcile"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
)

func NewSnapshotGenerator(rm core_manager.ReadOnlyResourceManager) util_xds_v3.SnapshotGenerator {
	return mads_reconcile.NewSnapshotGenerator(rm, mads_generator.MonitoringAssignmentsGenerator{})
}

func NewVersioner() util_xds_v3.SnapshotVersioner {
	return util_xds_v3.SnapshotAutoVersioner{UUID: core.NewUUID}
}

func NewReconciler(hasher envoy_cache.NodeHash, cache envoy_cache.SnapshotCache,
	generator util_xds_v3.SnapshotGenerator, versioner util_xds_v3.SnapshotVersioner,
) mads_reconcile.Reconciler {
	return mads_reconcile.NewReconciler(hasher, cache, generator, versioner)
}

type restReconcilerCallbacks struct {
	reconciler mads_reconcile.Reconciler
}

func (r *restReconcilerCallbacks) OnFetchRequest(ctx context.Context, request util_xds.DiscoveryRequest) error {
	nodei := request.Node()

	node, ok := nodei.(*envoy_core.Node)
	if !ok {
		return errors.Errorf("expecting a v3 Node, got: %v", nodei)
	}

	// only reconcile if there is not a valid response present
	if !r.reconciler.NeedsReconciliation(node) {
		return nil
	}

	return r.reconciler.Reconcile(ctx, node)
}

func (r *restReconcilerCallbacks) OnFetchResponse(request util_xds.DiscoveryRequest, response util_xds.DiscoveryResponse) {
}

func NewReconcilerRestCallbacks(reconciler mads_reconcile.Reconciler) util_xds.RestCallbacks {
	return &restReconcilerCallbacks{reconciler: reconciler}
}

func NewSyncTracker(reconciler mads_reconcile.Reconciler, refresh time.Duration, log logr.Logger) envoy_xds.Callbacks {
	return util_xds_v3.NewWatchdogCallbacks(func(ctx context.Context, node *envoy_core.Node, streamID int64) (util_watchdog.Watchdog, error) {
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

func NewXdsContext(log logr.Logger) (envoy_cache.NodeHash, envoy_cache.SnapshotCache) {
	hasher := hasher{}
	logger := util_xds.NewLogger(log)
	return hasher, envoy_cache.NewSnapshotCache(false, hasher, logger)
}

type hasher struct{}

func (_ hasher) ID(node *envoy_core.Node) string {
	// in the very first implementation, we don't differentiate clients
	return ""
}
