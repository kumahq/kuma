package bootstrap

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	dp_server "github.com/kumahq/kuma/pkg/config/dp-server"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func RegisterBootstrap(rt core_runtime.Runtime) error {
	generator, err := NewDefaultBootstrapGenerator(
		rt.ResourceManager(),
		rt.Config().BootstrapServer,
		rt.Config().Proxy,
		rt.Config().DpServer.TlsCertFile,
		map[string]bool{
			string(mesh_proto.DataplaneProxyType): rt.Config().DpServer.Authn.DpProxy.Type != dp_server.DpServerAuthNone,
			string(mesh_proto.IngressProxyType):   rt.Config().DpServer.Authn.ZoneProxy.Type != dp_server.DpServerAuthNone,
			string(mesh_proto.EgressProxyType):    rt.Config().DpServer.Authn.ZoneProxy.Type != dp_server.DpServerAuthNone,
		},
		rt.Config().DpServer.Authn.EnableReloadableTokens,
		rt.Config().DpServer.Hds.Enabled,
		rt.Config().GetEnvoyAdminPort(),
		rt.Config().Experimental.UseDeltaXDS,
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
