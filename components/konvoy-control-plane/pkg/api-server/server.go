package api_server

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server/mesh"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/emicklei/go-restful"
	"github.com/prometheus/common/log"
	"net/http"
)

type ApiServerConfig struct {
	BindAddress string
}

type ApiServer struct {
	server *http.Server
}

func (a *ApiServer) Address() string {
	return a.server.Addr
}

func NewApiServer(resourceStore store.ResourceStore, config ApiServerConfig) *ApiServer {
	apiServer := &ApiServer{}

	srv := &http.Server{Addr: config.BindAddress}
	apiServer.server = srv

	container := restful.NewContainer()
	container.Add(mesh.NewProxyTemplateWs(resourceStore))

	srv.Handler = container.ServeMux

	return apiServer
}

func (a *ApiServer) Start() {
	go func(){
		err := a.server.ListenAndServe()
		if err != nil {
			log.Fatalf("Could not start an HTTP Server", err)
		}
	}()
}

func (a *ApiServer) Stop() error {
	return a.server.Shutdown(context.Background())
}
