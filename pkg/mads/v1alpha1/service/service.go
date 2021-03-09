package service

import (
	"github.com/go-logr/logr"
	"github.com/kumahq/kuma/pkg/config/mads"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
)

type service struct {
	server Server
	config *mads.MonitoringAssignmentServerConfig
	log logr.Logger
}

func NewService(config *mads.MonitoringAssignmentServerConfig, rm core_manager.ReadOnlyResourceManager, log logr.Logger) *service {
	hasher, cache := NewXdsContext(log)
	generator := NewSnapshotGenerator(rm)
	versioner := NewVersioner()
	reconciler := NewReconciler(hasher, cache, generator, versioner)
	syncTracker := NewSyncTracker(reconciler, config.AssignmentRefreshInterval, log)
	callbacks := util_xds_v2.CallbacksChain{
		util_xds_v2.AdaptCallbacks(util_xds.LoggingCallbacks{Log: log}),
		syncTracker,
	}
	srv := NewServer(cache, callbacks)

	return &service{
		server: srv,
		config: config,
		log: log,
	}
}
