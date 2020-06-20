package server

import (
	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

var (
	kdsServerLog = core.Log.WithName("kds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	hasher, cache := NewXdsContext(kdsServerLog)
	generator := NewSnapshotGenerator(rt)
	versioner := NewVersioner()
	reconciler := NewReconciler(hasher, cache, generator, versioner)
	syncTracker := NewSyncTracker(reconciler, rt.Config().KDSServer.RefreshInterval)
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: kdsServerLog},
		syncTracker,
	}
	srv := NewServer(cache, callbacks, kdsServerLog)
	return rt.Add(
		&grpcServer{srv, *rt.Config().KDSServer},
	)
}
