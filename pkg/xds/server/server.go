package server

import (
	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
	"github.com/Kong/kuma/pkg/xds/bootstrap"
	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"
)

var (
	xdsServerLog = core.Log.WithName("xds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	reconciler := DefaultReconciler(rt)

	metadataTracker := newDataplaneMetadataTracker()

	tracker, err := DefaultDataplaneSyncTracker(rt, reconciler, metadataTracker)
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		tracker,
		metadataTracker,
		DefaultDataplaneStatusTracker(rt),
	}

	srv := envoy_xds.NewServer(rt.XDS().Cache(), callbacks)
	return core_runtime.Add(
		rt,
		// xDS gRPC API
		&grpcServer{srv, rt.Config().XdsServer.GrpcPort},
		// diagnostics server
		&diagnosticsServer{rt.Config().XdsServer.DiagnosticsPort},
		// bootstrap server
		&bootstrap.BootstrapServer{
			Port:      rt.Config().BootstrapServer.Port,
			Generator: bootstrap.NewDefaultBootstrapGenerator(rt.ResourceManager(), rt.Config().BootstrapServer.Params),
		},
	)
}
