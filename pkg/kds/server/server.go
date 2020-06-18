package server

import (
	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	mads_server "github.com/Kong/kuma/pkg/mads/server"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

var (
	kdsServerLog = core.Log.WithName("kds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	hasher, cache := mads_server.NewXdsContext(kdsServerLog)
	generator := NewSnapshotGenerator(rt)
	versioner := mads_server.NewVersioner()
	reconciler := mads_server.NewReconciler(hasher, cache, generator, versioner)
	syncTracker := mads_server.NewSyncTracker(reconciler, rt.Config().KdsServer.RefreshInterval)
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: kdsServerLog},
		syncTracker,
	}
	srv := NewServer(cache, callbacks, kdsServerLog)
	return rt.Add(
		&grpcServer{srv, *rt.Config().KdsServer},
	)
}
