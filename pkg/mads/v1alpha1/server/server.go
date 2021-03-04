package server

import (
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/mads"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"
	"github.com/pkg/errors"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

var (
	madsServerLog = core.Log.WithName("mads-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	config := *rt.Config().MonitoringAssignmentServer
	if !config.VersionIsEnabled(mads.MADS_V1_ALPHA1) {
		madsServerLog.Info("MADS v1alpha1 not enabled")
		return nil
	}

	if !config.GrpcEnabled {
		return errors.Errorf("MADS v1alpha1 only supports gRPC, which is not enabled")
	}

	hasher, cache := NewXdsContext(madsServerLog)
	generator := NewSnapshotGenerator(rt)
	versioner := NewVersioner()
	reconciler := NewReconciler(hasher, cache, generator, versioner)
	syncTracker := NewSyncTracker(reconciler, rt.Config().MonitoringAssignmentServer.AssignmentRefreshInterval)
	callbacks := util_xds_v2.CallbacksChain{
		util_xds_v2.AdaptCallbacks(util_xds.LoggingCallbacks{Log: madsServerLog}),
		syncTracker,
	}
	srv := NewServer(cache, callbacks)
	return rt.Add(
		&grpcServer{
			server:  srv,
			config:  config,
			metrics: rt.Metrics(),
		},
	)
}
