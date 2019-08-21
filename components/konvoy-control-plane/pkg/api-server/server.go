package api_server

import (
	"context"
	"fmt"
	core_resources "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources"
	"net/http"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/api-server/definitions"
	config "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/api-server"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
)

var (
	log = core.Log.WithName("api-server")
)

type ApiServer struct {
	server *http.Server
}

func (a *ApiServer) Address() string {
	return a.server.Addr
}

func NewApiServer(resources core_resources.Resources, defs []definitions.ResourceWsDefinition, config config.ApiServerConfig) *ApiServer {
	container := restful.NewContainer()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: container.ServeMux,
	}

	ws := new(restful.WebService)
	ws.
		Path("/meshes").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	addToWs(ws, defs, resources, config)
	container.Add(ws)
	configureOpenApi(config, container, ws)

	return &ApiServer{
		server: srv,
	}
}

func addToWs(ws *restful.WebService, defs []definitions.ResourceWsDefinition, resources core_resources.Resources, config config.ApiServerConfig) {
	overviewWs := overviewWs{
		resources: resources,
	}
	overviewWs.AddToWs(ws)

	for _, definition := range defs {
		resourceWs := resourceWs{
			resources:            resources,
			readOnly:             config.ReadOnly,
			ResourceWsDefinition: definition,
		}
		resourceWs.AddToWs(ws)
	}
}

func configureOpenApi(config config.ApiServerConfig, container *restful.Container, webService *restful.WebService) {
	openApiConfig := restfulspec.Config{
		WebServices: []*restful.WebService{webService},
		APIPath:     config.ApiDocsPath,
	}
	container.Add(restfulspec.NewOpenAPIService(openApiConfig))

	// todo(jakubdyszkiewicz) figure out how to pack swagger ui dist package and expose swagger ui
	//container.Handle("/apidocs/", http.StripPrefix("/apidocs/", http.FileServer(http.Dir("path/to/swagger-ui-dist"))))
}

func (a *ApiServer) Start(stop <-chan struct{}) error {
	errChan := make(chan error)
	go func() {
		err := a.server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				log.Info("Shutting down server")
			default:
				log.Error(err, "Could not start an HTTP Server")
				errChan <- err
			}
		}
	}()
	select {
	case <-stop:
		log.Info("Stopping down API Server")
		return a.server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func SetupServer(rt runtime.Runtime) error {
	apiServer := NewApiServer(rt.Resources(), definitions.All, *rt.Config().ApiServer)
	return rt.Add(apiServer)
}
