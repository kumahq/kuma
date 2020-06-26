package remote

import (
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	kds_server "github.com/Kong/kuma/pkg/kds/server"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

var (
	kdsServerLog  = core.Log.WithName("kds-server")
	resourceTypes = []model.ResourceType{mesh.DataplaneType}
)

func SetupServer(rt core_runtime.Runtime) error {
	hasher, cache := kds_server.NewXdsContext(kdsServerLog)
	generator := kds_server.NewSnapshotGenerator(rt, resourceTypes)
	versioner := kds_server.NewVersioner()
	reconciler := kds_server.NewReconciler(hasher, cache, generator, versioner)
	syncTracker := kds_server.NewSyncTracker(kdsServerLog, reconciler, rt.Config().KDSServer.RefreshInterval)
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: kdsServerLog},
		syncTracker,
	}
	srv := kds_server.NewServer(cache, callbacks, kdsServerLog)
	return rt.Add(kds_server.NewKDSServer(srv, *rt.Config().KDSServer))
}
