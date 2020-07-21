package api_server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kumahq/kuma/app/kuma-ui/pkg/server/types"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	gui_server "github.com/kumahq/kuma/pkg/config/gui-server"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/kumahq/kuma/pkg/config/mode"

	"github.com/kumahq/kuma/pkg/zones/poller"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/app/kuma-ui/pkg/resources"
	"github.com/kumahq/kuma/pkg/api-server/definitions"
	api_server_config "github.com/kumahq/kuma/pkg/config/api-server"
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
	GuiServerConfig *gui_server.GuiServerConfig
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

func NewApiServer(resManager manager.ResourceManager, clusters poller.ZoneStatusPoller, defs []definitions.ResourceWsDefinition, serverConfig *api_server_config.ApiServerConfig, cfg *kuma_cp.Config) (*ApiServer, error) {
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

	addResourcesEndpoints(ws, defs, resManager, serverConfig)
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

	clustersWs := clustersWs(clusters)
	container.Add(clustersWs)

	container.Filter(cors.Filter)

	newApiServer := &ApiServer{
		server: srv,
		GuiServerConfig: cfg.GuiServer,
	}

	// Handle the GUI
	container.Handle("/gui/", http.StripPrefix("/gui/", http.FileServer(resources.GuiDir)))
	container.Handle("/api/", http.RedirectHandler("/", http.StatusPermanentRedirect))
	container.ServeMux.HandleFunc("/config", newApiServer.configHandler)

	return newApiServer, nil
}

func addResourcesEndpoints(ws *restful.WebService, defs []definitions.ResourceWsDefinition, resManager manager.ResourceManager, config *api_server_config.ApiServerConfig) {
	endpoints := dataplaneOverviewEndpoints{
		publicURL:  config.Catalog.ApiServer.Url,
		resManager: resManager,
	}
	endpoints.addListEndpoint(ws, "/meshes/{mesh}")
	endpoints.addFindEndpoint(ws, "/meshes/{mesh}")
	endpoints.addListEndpoint(ws, "") // listing all resources in all meshes

	for _, definition := range defs {
		switch definition.ResourceFactory().GetType() {
		case mesh.MeshType:
			endpoints := resourceEndpoints{
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

func (a *ApiServer) configHandler(writer http.ResponseWriter, request *http.Request) {
	bytes, err := json.Marshal(fromServerConfig(*a.GuiServerConfig.GuiConfig))
	if err != nil {
		log.Error(err, "could not marshall config")
		writer.WriteHeader(500)
		return
	}
	writer.Header().Add("content-type", "application/json")
	if _, err := writer.Write(bytes); err != nil {
		log.Error(err, "could not write the response")
	}
}

func fromServerConfig(config gui_server.GuiConfig) types.GuiConfig {
	return types.GuiConfig{
		ApiUrl:      config.ApiUrl,
		Environment: config.Environment,
	}
}

func SetupServer(rt runtime.Runtime) error {
	cfg := rt.Config()
	if cfg.Mode.Mode != mode.Standalone {
		for i, definition := range definitions.All {
			switch cfg.Mode.Mode {
			case mode.Global:
				if definition.ResourceFactory().GetType() == mesh.DataplaneType {
					definitions.All[i].ReadOnly = true
				}
			case mode.Remote:
				if definition.ResourceFactory().GetType() != mesh.DataplaneType {
					definitions.All[i].ReadOnly = true
				}
			}
		}
	}
	apiServer, err := NewApiServer(rt.ResourceManager(), rt.Zones(), definitions.All, rt.Config().ApiServer, &cfg)
	if err != nil {
		return err
	}
	return rt.Add(apiServer)
}
