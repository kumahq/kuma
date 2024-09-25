package service

import (
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/config/mads"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	mads_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	mads_reconcile "github.com/kumahq/kuma/pkg/mads/v1/reconcile"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
)

type service struct {
	server Server
	config *mads.MonitoringAssignmentServerConfig
	log    logr.Logger
}

func NewService(config *mads.MonitoringAssignmentServerConfig, rm core_manager.ReadOnlyResourceManager, log logr.Logger, meshCache *mesh.Cache) *service {
	hasher := util_xds_v3.IDHash{}
	cache := util_xds_v3.NewSnapshotCache(false, hasher, util_xds.NewLogger(log))
	reconciler := mads_reconcile.NewReconciler(
		hasher,
		cache,
		mads_reconcile.NewSnapshotGenerator(rm, mads_generator.MonitoringAssignmentsGenerator{}, meshCache),
	)

	callbacks := util_xds_v3.CallbacksChain{
		util_xds_v3.AdaptMultiCallbacks(util_xds.LoggingCallbacks{Log: log}),
		// TODO Right now we are creating sync tracker for first REST request pkg/util/xds/v3/watchdog_callbacks.go:48, because of this
		// using MeshMetrics with specified clientId will refresh only one client in background
		// issue: https://github.com/kumahq/kuma/issues/8764
		NewSyncTracker(reconciler, config.AssignmentRefreshInterval.Duration, log),
		// For on-demand reconciliation
		util_xds_v3.AdaptRestCallbacks(NewReconcilerRestCallbacks(reconciler, log)),
	}
	srv := NewServer(cache, callbacks)

	return &service{
		server: srv,
		config: config,
		log:    log,
	}
}
