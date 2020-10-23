package bootstrap

import (
	"net/http"

	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func RegisterBootstrap(rt core_runtime.Runtime, mux *http.ServeMux) {
	bootstrapHandler := BootstrapHandler{
		Generator: NewDefaultBootstrapGenerator(rt.ResourceManager(), rt.Config().BootstrapServer.Params, rt.Config().DpServer.TlsCertFile),
	}
	log.Info("registering Bootstrap in Dataplane Server")
	mux.HandleFunc("/bootstrap", bootstrapHandler.Handle)
}
