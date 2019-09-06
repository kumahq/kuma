package server

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	util_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/xds"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/xds/bootstrap"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

var (
	xdsServerLog = core.Log.WithName("xds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	reconciler := DefaultReconciler(rt)

	tracker, err := DefaultDataplaneSyncTracker(rt, reconciler)
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		tracker,
		DefaultDataplaneStatusTracker(rt),
	}

	srv := envoy_xds.NewServer(rt.XDS().Cache(), callbacks)
	return core_runtime.Add(
		rt,
		// xDS gRPC API
		&grpcServer{srv, rt.Config().XdsServer.GrpcPort},
		// xDS HTTP API
		&httpGateway{srv, rt.Config().XdsServer.HttpPort},
		// diagnostics server
		&diagnosticsServer{rt.Config().XdsServer.DiagnosticsPort},
		// bootstrap server
		&bootstrap.BootstrapServer{
			Port:      rt.Config().BootstrapServer.Port,
			Generator: bootstrap.NewDefaultBootstrapGenerator(rt.ResourceManager(), rt.Config().BootstrapServer.Params),
		},
	)
}
