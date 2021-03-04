package server

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/mads"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v3 "github.com/kumahq/kuma/pkg/util/xds/v3"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

var (
	madsServerLog = core.Log.WithName("mads-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	config := *rt.Config().MonitoringAssignmentServer
	if !config.VersionIsEnabled(mads.MADS_V1) {
		madsServerLog.Info("MADS v1 not enabled")
		return nil
	}

	hasher, cache := NewXdsContext(madsServerLog)
	generator := NewSnapshotGenerator(rt)
	versioner := NewVersioner()
	reconciler := NewReconciler(hasher, cache, generator, versioner)
	syncTracker := NewSyncTracker(reconciler, rt.Config().MonitoringAssignmentServer.AssignmentRefreshInterval)
	grpcCallbacks := util_xds_v3.CallbacksChain{
		util_xds_v3.AdaptCallbacks(util_xds.LoggingCallbacks{Log: madsServerLog}),
		syncTracker,
	}
	srv := NewServer(cache, grpcCallbacks)

	if config.HttpEnabled {
		if err := rt.Add(&httpServer{
			server: srv,
			config: config,
			metrics: rt.Metrics(),
		}); err != nil {
			return err
		}
	}

	if config.GrpcEnabled {
		if err := rt.Add(&grpcServer{
			server:  srv,
			config:  config,
			metrics: rt.Metrics(),
		}); err != nil {
			return err
		}
	}

	return nil
}
