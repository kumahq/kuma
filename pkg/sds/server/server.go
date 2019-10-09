package server

import (
	config_core "github.com/Kong/kuma/pkg/config/core"
	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	util_xds "github.com/Kong/kuma/pkg/util/xds"
)

var (
	sdsServerLog = core.Log.WithName("sds-server")
)

func SetupServer(rt core_runtime.Runtime) error {
	if err := setupInitialTokenServer(rt); err != nil {
		return err
	}
	if err := setupGrpcServer(rt); err != nil {
		return err
	}
	return nil
}

func setupInitialTokenServer(rt core_runtime.Runtime) error {
	if rt.Config().Environment != config_core.KubernetesEnvironment {
		generator, err := NewCredentialGenerator(rt)
		if err != nil {
			return err
		}
		srv := &InitialTokenServer{
			LocalHttpPort:       rt.Config().InitialTokenServer.LocalHttpPort,
			CredentialGenerator: generator,
		}
		if err := core_runtime.Add(rt, srv); err != nil {
			return err
		}
	}
	return nil
}

func setupGrpcServer(rt core_runtime.Runtime) error {
	handler, err := DefaultSecretDiscoveryHandler(rt)
	if err != nil {
		return err
	}
	callbacks := util_xds.CallbacksChain{
		util_xds.LoggingCallbacks{Log: sdsServerLog},
	}
	srv := NewServer(handler, callbacks, sdsServerLog)
	return core_runtime.Add(
		rt,
		&grpcServer{srv, *rt.Config().SdsServer},
	)
}
