package bootstrap

import (
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	dp_server "github.com/kumahq/kuma/v3/pkg/config/dp-server"
	core_runtime "github.com/kumahq/kuma/v3/pkg/core/runtime"
)

func RegisterBootstrap(rt core_runtime.Runtime) error {
	// The same key pair the DP server serves, so that the CA handed to a proxy
	// and the SANs validated against it describe the certificate that proxy is
	// about to be presented.
	dpServerKeyPair, err := rt.CertWatchers().Watch(rt.Config().DpServer.TlsCertFile, rt.Config().DpServer.TlsKeyFile)
	if err != nil {
		return err
	}
	generator, err := NewDefaultBootstrapGenerator(
		rt.ResourceManager(),
		rt.Config().BootstrapServer,
		dpServerKeyPair,
		map[string]bool{
			string(mesh_proto.DataplaneProxyType): rt.Config().DpServer.Authn.DpProxy.Type != dp_server.DpServerAuthNone,
		},
		rt.Config().DpServer.Authn.EnableReloadableTokens,
		rt.Config().DpServer.Hds.Enabled,
		rt.Config().GetEnvoyAdminPort(),
		rt.Config().BootstrapServer.Params.EnvoyAdminUnixSocket,
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
