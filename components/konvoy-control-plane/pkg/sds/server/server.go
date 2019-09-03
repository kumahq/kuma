package server

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	util_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/xds"
)

var (
	sdsServerLog = core.Log.WithName("sds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	handler, err := DefaultSecretDiscoveryHandler(rt)
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: sdsServerLog},
	}
	srv := NewServer(handler, callbacks, sdsServerLog)
	return core_runtime.Add(rt, &grpcServer{srv, *rt.Config().SdsServer})
}
