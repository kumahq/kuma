package api_server

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"net/http"

	"github.com/Kong/kuma/pkg/api-server/definitions"
	config "github.com/Kong/kuma/pkg/config/api-server"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/emicklei/go-restful"
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

func NewApiServer(resManager manager.ResourceManager, defs []definitions.ResourceWsDefinition, config kuma_cp.Config) (*ApiServer, error) {
	container := restful.NewContainer()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.ApiServer.Port),
		Handler: container.ServeMux,
	}

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{restful.HEADER_AccessControlAllowOrigin},
		AllowedDomains: config.ApiServer.CorsAllowedDomains,
		Container:      container,
	}

	ws := new(restful.WebService)
	ws.
		Path("/meshes").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	addToWs(ws, defs, resManager, config.ApiServer)
	container.Add(ws)
	container.Add(indexWs())
	container.Add(catalogWs(*config.ApiServer.Catalog))
	configWs, err := configWs(&config)
	if err != nil {
		return nil, errors.Wrap(err, "could not create configuration webservice")
	}
	container.Add(configWs)

	container.Filter(cors.Filter)
	return &ApiServer{
		server: srv,
	}, nil
}

func addToWs(ws *restful.WebService, defs []definitions.ResourceWsDefinition, resManager manager.ResourceManager, config *config.ApiServerConfig) {
	overviewWs := overviewWs{
		resManager: resManager,
	}
	overviewWs.AddToWs(ws)

	for _, definition := range defs {
		resourceWs := resourceWs{
			resManager:           resManager,
			readOnly:             config.ReadOnly,
			ResourceWsDefinition: definition,
		}
		resourceWs.AddToWs(ws)
	}
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
	log.Info("starting", "port", a.Address())
	select {
	case <-stop:
		log.Info("Stopping down API Server")
		return a.server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func SetupServer(rt runtime.Runtime) error {
	apiServer, err := NewApiServer(rt.ResourceManager(), definitions.All, rt.Config())
	if err != nil {
		return err
	}
	return rt.Add(apiServer)
}
