package server

import (
	"github.com/Kong/kuma/pkg/core"

	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
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
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: madsServerLog},
		syncTracker,
	}
	srv := NewServer(cache, callbacks, madsServerLog)
	return rt.Add(
		&grpcServer{srv, *rt.Config().MonitoringAssignmentServer},
	)
}
