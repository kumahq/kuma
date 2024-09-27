package service

import (
	"context"
	"time"

	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/go-logr/logr"

	"github.com/kumahq/kuma/pkg/config/mads"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	mads_generator "github.com/kumahq/kuma/pkg/mads/v1/generator"
	mads_reconcile "github.com/kumahq/kuma/pkg/mads/v1/reconcile"
	util_watchdog "github.com/kumahq/kuma/pkg/util/watchdog"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"
	"github.com/kumahq/kuma/pkg/xds/cache/mesh"
)

type service struct {
	server   Server
	config   *mads.MonitoringAssignmentServerConfig
	log      logr.Logger
	watchdog *util_watchdog.SimpleWatchdog
}

func NewService(config *mads.MonitoringAssignmentServerConfig, rm core_manager.ReadOnlyResourceManager, log logr.Logger, meshCache *mesh.Cache) *service {
	hasher := &util_xds_v3.FallBackNodeHash{DefaultId: mads_generator.DefaultKumaClientId}
	cache := envoy_cache.NewSnapshotCache(false, hasher, util_xds.NewLogger(log))
	reconciler := mads_reconcile.NewReconciler(
		cache,
		mads_reconcile.NewSnapshotGenerator(rm, mads_generator.MonitoringAssignmentsGenerator{}, meshCache),
	)
	// We use the clientIds from the reconciler from the node hasher
	hasher.GetIds = reconciler.KnownClientIds
	watchdog := &util_watchdog.SimpleWatchdog{
		NewTicker: func() *time.Ticker {
			return time.NewTicker(config.AssignmentRefreshInterval.Duration)
		},
		OnTick: func(ctx context.Context) error {
			log.V(1).Info("on tick")
			err := reconciler.Reconcile(ctx)
			return err
		},
		OnError: func(err error) {
			log.Error(err, "OnTick() failed")
		},
	}
	watchdog.WithTickCheck()
	srv := NewServer(cache, util_xds_v3.CallbacksChain{
		util_xds_v3.AdaptMultiCallbacks(util_xds.LoggingCallbacks{Log: log}),
	})

	return &service{
		server:   srv,
		config:   config,
		log:      log,
		watchdog: watchdog,
	}
}

func (s *service) Start(ctx context.Context) {
	go s.watchdog.Start(ctx)
	s.watchdog.HasTicked(true)
}
