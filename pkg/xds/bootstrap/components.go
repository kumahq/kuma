package bootstrap

import (
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func RegisterBootstrap(rt core_runtime.Runtime) error {
	generator, err := NewDefaultBootstrapGenerator(
		rt.ResourceManager(),
		rt.Config().BootstrapServer,
		rt.Config().DpServer.TlsCertFile,
		rt.Config().DpServer.Auth.Type != dp_server.DpServerAuthNone,
		rt.Config().DpServer.Hds.Enabled,
		rt.Config().GetEnvoyAdminPort(),
	)
	if err != nil {
		return err
	}
	bootstrapHandler := BootstrapHandler{
		Generator: generator,
	}
	log.Info("registering Bootstrap in Dataplane Server")
	rt.DpServer().HTTPMux().HandleFunc("/bootstrap", bootstrapHandler.Handle)
	return nil
}
