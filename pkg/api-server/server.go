package api_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"io"
	"net/http"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/app/kuma-ui/pkg/resources"
	"github.com/kumahq/kuma/pkg/api-server/definitions"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime"
)

var (
	log = core.Log.WithName("api-server")
)

type ApiServer struct {
	server *http.Server
}

func (a *ApiServer) NeedLeaderElection() bool {
	return false
}

func (a *ApiServer) Address() string {
	return a.server.Addr
}

func init() {
	// turn off escape & character so the link in "next" fields for resources is user friendly
	restful.NewEncoder = func(w io.Writer) *json.Encoder {
		encoder := json.NewEncoder(w)
		encoder.SetEscapeHTML(false)
		return encoder
	}
	restful.MarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
		var buf bytes.Buffer
		encoder := restful.NewEncoder(&buf)
		encoder.SetIndent(prefix, indent)
		if err := encoder.Encode(v); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
}

func NewApiServer(resManager manager.ResourceManager, defs []definitions.ResourceWsDefinition, cfg *kuma_cp.Config, enableGUI bool) (*ApiServer, error) {
	serverConfig := cfg.ApiServer
	container := restful.NewContainer()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverConfig.Port),
		Handler: container.ServeMux,
	}

	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{restful.HEADER_AccessControlAllowOrigin},
		AllowedDomains: serverConfig.CorsAllowedDomains,
		Container:      container,
	}

	// We create a WebService and set up resources endpoints and index endpoint instead of creating WebService
	// for every resource like /meshes/{mesh}/traffic-permissions, /meshes/{mesh}/traffic-log etc.
	// because go-restful detects it as a clash (you cannot register 2 WebServices with path /meshes/)
	ws := new(restful.WebService)
	ws.
		Path("/").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	addResourcesEndpoints(ws, defs, resManager, cfg)
	container.Add(ws)

	if err := addIndexWsEndpoints(ws); err != nil {
		return nil, errors.Wrap(err, "could not create index webservice")
	}
	container.Add(catalogWs(*serverConfig.Catalog))
	configWs, err := configWs(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not create configuration webservice")
	}
	container.Add(configWs)

	zonesWs := zonesWs(resManager)
	container.Add(zonesWs)

	container.Filter(cors.Filter)

	newApiServer := &ApiServer{
		server: srv,
	}

	// Handle the GUI
	if enableGUI {
		container.Handle("/gui/", http.StripPrefix("/gui/", http.FileServer(resources.GuiDir)))
	} else {
		container.ServeMux.HandleFunc("/gui/", newApiServer.notAvailableHandler)
	}
	return newApiServer, nil
}

func addResourcesEndpoints(ws *restful.WebService, defs []definitions.ResourceWsDefinition, resManager manager.ResourceManager, cfg *kuma_cp.Config) {
	config := cfg.ApiServer
	endpoints := dataplaneOverviewEndpoints{
		publicURL:  config.Catalog.ApiServer.Url,
		resManager: resManager,
	}
	endpoints.addListEndpoint(ws, "/meshes/{mesh}")
	endpoints.addFindEndpoint(ws, "/meshes/{mesh}")
	endpoints.addListEndpoint(ws, "") // listing all resources in all meshes

	zoneOverviewEndpoints := zoneOverviewEndpoints{
		publicURL:  config.Catalog.ApiServer.Url,
		resManager: resManager,
	}
	zoneOverviewEndpoints.addFindEndpoint(ws)
	zoneOverviewEndpoints.addListEndpoint(ws)

	for _, definition := range defs {
		switch definition.ResourceFactory().GetType() {
		case mesh.MeshType:
			endpoints := resourceEndpoints{
				mode:                 cfg.Mode,
				publicURL:            config.Catalog.ApiServer.Url,
				resManager:           resManager,
				ResourceWsDefinition: definition,
				meshFromRequest:      meshFromPathParam("name"),
			}
			if config.ReadOnly || definition.ReadOnly {
				endpoints.addCreateOrUpdateEndpointReadOnly(ws, "/meshes")
				endpoints.addDeleteEndpointReadOnly(ws, "/meshes")
			} else {
				endpoints.addCreateOrUpdateEndpoint(ws, "/meshes")
				endpoints.addDeleteEndpoint(ws, "/meshes")
			}
			endpoints.addFindEndpoint(ws, "/meshes")
			endpoints.addListEndpoint(ws, "/meshes")
		case system.ZoneType:
			endpoints := resourceEndpoints{
				publicURL:            config.Catalog.ApiServer.Url,
				resManager:           resManager,
				ResourceWsDefinition: definition,
				meshFromRequest: func(request *restful.Request) string {
					return "default"
				},
			}
			if config.ReadOnly || definition.ReadOnly {
				endpoints.addCreateOrUpdateEndpointReadOnly(ws, "/zones")
				endpoints.addDeleteEndpointReadOnly(ws, "/zones")
			} else {
				endpoints.addCreateOrUpdateEndpoint(ws, "/zones")
				endpoints.addDeleteEndpoint(ws, "/zones")
			}
			endpoints.addFindEndpoint(ws, "/zones")
			endpoints.addListEndpoint(ws, "/zones")
		default:
			endpoints := resourceEndpoints{
				publicURL:            config.Catalog.ApiServer.Url,
				resManager:           resManager,
				ResourceWsDefinition: definition,
				meshFromRequest:      meshFromPathParam("mesh"),
			}
			if config.ReadOnly || definition.ReadOnly {
				endpoints.addCreateOrUpdateEndpointReadOnly(ws, "/meshes/{mesh}/"+definition.Path)
				endpoints.addDeleteEndpointReadOnly(ws, "/meshes/{mesh}/"+definition.Path)
			} else {
				endpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
				endpoints.addDeleteEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
			}
			endpoints.addFindEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
			endpoints.addListEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
			endpoints.addListEndpoint(ws, "/"+definition.Path) // listing all resources in all meshes
		}
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
	log.Info("starting", "interface", "0.0.0.0", "port", a.Address())
	select {
	case <-stop:
		log.Info("Stopping down API Server")
		return a.server.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (a *ApiServer) notAvailableHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("" +
		"<!DOCTYPE html><html lang=en>" +
		"<head>\n<style>\n.center {\n  display: flex;\n  justify-content: center;\n  align-items: center;\n  height: 200px;\n  border: 3px solid green; \n}\n</style>\n</head>" +
		"<body><div class=\"center\"><strong>" +
		"GUI is disabled. If this is a Remote CP, please check the GUI on the Global CP." +
		"</strong></div></body>" +
		"</html>"))
	if err != nil {
		log.Error(err, "could not write the response")
	}
}

func SetupServer(rt runtime.Runtime) error {
	cfg := rt.Config()
	enableGUI := cfg.Mode != config_core.Remote
	if cfg.Mode != config_core.Standalone {
		for i, definition := range definitions.All {
			switch cfg.Mode {
			case config_core.Global:
				if definition.ResourceFactory().GetType() == mesh.DataplaneType {
					definitions.All[i].ReadOnly = true
				}
			case config_core.Remote:
				if definition.ResourceFactory().GetType() != mesh.DataplaneType {
					definitions.All[i].ReadOnly = true
				}
			}
		}
	}
	apiServer, err := NewApiServer(rt.ResourceManager(), definitions.All, &cfg, enableGUI)
	if err != nil {
		return err
	}
	return rt.Add(apiServer)
}
