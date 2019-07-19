package api_server

import (
	"context"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"net/http"
)

type ApiServerConfig struct {
	BindAddress string `envconfig:"api_server_address" default:"0.0.0.0:8091"`
	ReadOnly    bool   `envconfig:"api_server_read_only" default:"false"`
	ApiPath     string `envconfig:"api_server_api_path" default:"/apidocs.json"`
}

type ApiServer struct {
	server *http.Server
}

func (a *ApiServer) Address() string {
	return a.server.Addr
}

func NewApiServer(resourceStore store.ResourceStore, definitions []ResourceWsDefinition, config ApiServerConfig) *ApiServer {
	container := restful.NewContainer()
	srv := &http.Server{
		Addr:    config.BindAddress,
		Handler: container.ServeMux,
	}

	ws := new(restful.WebService)
	ws.
		Path("/meshes").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	addToWs(ws, definitions, resourceStore, config)
	container.Add(ws)
	configureOpenApi(config, container, ws)

	return &ApiServer{
		server: srv,
	}
}

func addToWs(ws *restful.WebService, definitions []ResourceWsDefinition, resourceStore store.ResourceStore, config ApiServerConfig) {
	for _, definition := range definitions {
		resourceWs := resourceWs{
			resourceStore:        resourceStore,
			readOnly:             config.ReadOnly,
			ResourceWsDefinition: definition,
		}
		resourceWs.AddToWs(ws)
	}
}

func configureOpenApi(config ApiServerConfig, container *restful.Container, webService *restful.WebService) {
	openApiConfig := restfulspec.Config{
		WebServices: []*restful.WebService{webService},
		APIPath:     config.ApiPath,
	}
	container.Add(restfulspec.NewOpenAPIService(openApiConfig))

	// todo(jakubdyszkiewicz) figure out how to pack swagger ui dist package and expose swagger ui
	//container.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("path/to/swagger-ui-dist"))))
}

func (a *ApiServer) Start() {
	go func() {
		err := a.server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				core.Log.Info("Shutting down server")
			default:
				core.Log.Error(err, "Could not start an HTTP Server")
			}
		}
	}()
}

func (a *ApiServer) Stop() error {
	return a.server.Shutdown(context.Background())
}
