package server

import (
	"github.com/kumahq/kuma/pkg/core"
	util_xds "github.com/kumahq/kuma/pkg/util/xds"
	util_xds_v2 "github.com/kumahq/kuma/pkg/util/xds/v2"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

var (
	madsServerLog = core.Log.WithName("mads-server")
)

func SetupServer(rt core_runtime.Runtime) error {
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
			config:  *rt.Config().MonitoringAssignmentServer,
			metrics: rt.Metrics(),
		},
	)
}
