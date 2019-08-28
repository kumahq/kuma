package server

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
)

var (
	sdsServerLog = core.Log.WithName("sds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	srv := NewServer(DefaultSecretDiscoveryHandler(), nil, sdsServerLog)
	return core_runtime.Add(rt, &grpcServer{srv, rt.Config().SdsServer.GrpcPort})
}
