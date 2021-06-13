package api_server

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kumahq/kuma/pkg/api-server/customization"

	"github.com/emicklei/go-restful"
	http_prometheus "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/app/kuma-ui/pkg/resources"
	"github.com/kumahq/kuma/pkg/api-server/authz"
	"github.com/kumahq/kuma/pkg/api-server/definitions"
	api_server "github.com/kumahq/kuma/pkg/config/api-server"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/metrics"
	"github.com/kumahq/kuma/pkg/tokens/builtin"
	tokens_server "github.com/kumahq/kuma/pkg/tokens/builtin/server"
	util_prometheus "github.com/kumahq/kuma/pkg/util/prometheus"
)

var (
	log = core.Log.WithName("api-server")
)

type ApiServer struct {
	mux    *http.ServeMux
	config api_server.ApiServerConfig
}

func (a *ApiServer) NeedLeaderElection() bool {
	return false
}

func (a *ApiServer) Address() string {
	return net.JoinHostPort(a.config.HTTP.Interface, strconv.FormatUint(uint64(a.config.HTTP.Port), 10))
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

func NewApiServer(resManager manager.ResourceManager, wsManager customization.APIInstaller, defs []definitions.ResourceWsDefinition, cfg *kuma_cp.Config, enableGUI bool, metrics metrics.Metrics) (*ApiServer, error) {
	serverConfig := cfg.ApiServer
	container := restful.NewContainer()

	promMiddleware := middleware.New(middleware.Config{
		Recorder: http_prometheus.NewRecorder(http_prometheus.Config{
			Registry: metrics,
			Prefix:   "api_server",
		}),
	})
	container.Filter(util_prometheus.MetricsHandler("", promMiddleware))

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
	configWs, err := configWs(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "could not create configuration webservice")
	}
	container.Add(configWs)

	container.Add(versionsWs())

	zonesWs := zonesWs(resManager)
	container.Add(zonesWs)

	container.Filter(cors.Filter)

	newApiServer := &ApiServer{
		mux:    container.ServeMux,
		config: *serverConfig,
	}

	dpWs, err := dataplaneTokenWs(resManager, cfg)
	if err != nil {
		return nil, err
	}
	if dpWs != nil {
		container.Add(dpWs)
	}

	// Handle the GUI
	if enableGUI {
		container.Handle("/gui/", http.StripPrefix("/gui/", http.FileServer(http.FS(resources.GuiFS()))))
	} else {
		container.ServeMux.HandleFunc("/gui/", newApiServer.notAvailableHandler)
	}

	wsManager.Install(container)

	return newApiServer, nil
}

func addResourcesEndpoints(ws *restful.WebService, defs []definitions.ResourceWsDefinition, resManager manager.ResourceManager, cfg *kuma_cp.Config) {
	config := cfg.ApiServer
	dpOverviewEndpoints := dataplaneOverviewEndpoints{
		resManager: resManager,
	}
	dpOverviewEndpoints.addListEndpoint(ws, "/meshes/{mesh}")
	dpOverviewEndpoints.addFindEndpoint(ws, "/meshes/{mesh}")
	dpOverviewEndpoints.addListEndpoint(ws, "") // listing all resources in all meshes

	zoneOverviewEndpoints := zoneOverviewEndpoints{
		resManager: resManager,
	}
	zoneOverviewEndpoints.addFindEndpoint(ws)
	zoneOverviewEndpoints.addListEndpoint(ws)

	zoneIngressOverviewEndpoints := zoneIngressOverviewEndpoints{
		resManager: resManager,
	}
	zoneIngressOverviewEndpoints.addFindEndpoint(ws)
	zoneIngressOverviewEndpoints.addListEndpoint(ws)

	serviceInsightEndpoints := serviceInsightEndpoints{
		resourceEndpoints: resourceEndpoints{
			mode:                 cfg.Mode,
			resManager:           resManager,
			ResourceWsDefinition: definitions.ServiceInsightWsDefinition,
			adminAuth: authz.AdminAuth{
				AllowFromLocalhost: cfg.ApiServer.Auth.AllowFromLocalhost,
			},
		},
	}
	serviceInsightEndpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+definitions.ServiceInsightWsDefinition.Path)
	serviceInsightEndpoints.addDeleteEndpoint(ws, "/meshes/{mesh}/"+definitions.ServiceInsightWsDefinition.Path)
	serviceInsightEndpoints.addFindEndpoint(ws, "/meshes/{mesh}/"+definitions.ServiceInsightWsDefinition.Path)
	serviceInsightEndpoints.addListEndpoint(ws, "/meshes/{mesh}/"+definitions.ServiceInsightWsDefinition.Path)
	serviceInsightEndpoints.addListEndpoint(ws, "/"+definitions.ServiceInsightWsDefinition.Path) // listing all resources in all meshes

	for _, definition := range defs {
		if config.ReadOnly {
			definition.ReadOnly = true
		}
		endpoints := resourceEndpoints{
			mode:                 cfg.Mode,
			resManager:           resManager,
			ResourceWsDefinition: definition,
			adminAuth: authz.AdminAuth{
				AllowFromLocalhost: cfg.ApiServer.Auth.AllowFromLocalhost,
			},
		}
		switch definition.ResourceFactory().Scope() {
		case model.ScopeMesh:
			endpoints.addCreateOrUpdateEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
			endpoints.addDeleteEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
			endpoints.addFindEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
			endpoints.addListEndpoint(ws, "/meshes/{mesh}/"+definition.Path)
			endpoints.addListEndpoint(ws, "/"+definition.Path) // listing all resources in all meshes
		case model.ScopeGlobal:
			endpoints.addCreateOrUpdateEndpoint(ws, "/"+definition.Path)
			endpoints.addDeleteEndpoint(ws, "/"+definition.Path)
			endpoints.addFindEndpoint(ws, "/"+definition.Path)
			endpoints.addListEndpoint(ws, "/"+definition.Path)
		}
	}
}

func dataplaneTokenWs(resManager manager.ResourceManager, cfg *kuma_cp.Config) (*restful.WebService, error) {
	dpIssuer, err := builtin.NewDataplaneTokenIssuer(resManager)
	if err != nil {
		return nil, err
	}
	zoneIngressIssuer, err := builtin.NewZoneIngressTokenIssuer(resManager)
	if err != nil {
		return nil, err
	}
	adminAuth := authz.AdminAuth{AllowFromLocalhost: cfg.ApiServer.Auth.AllowFromLocalhost}
	return tokens_server.NewWebservice(dpIssuer, zoneIngressIssuer).Filter(adminAuth.Validate), nil
}

func (a *ApiServer) Start(stop <-chan struct{}) error {
	errChan := make(chan error)

	var httpServer, httpsServer *http.Server
	if a.config.HTTP.Enabled {
		httpServer = a.startHttpServer(errChan)
	}
	if a.config.HTTPS.Enabled {
		httpsServer = a.startHttpsServer(errChan)
	}
	select {
	case <-stop:
		log.Info("stopping down API Server")
		if httpServer != nil {
			return httpServer.Shutdown(context.Background())
		}
		if httpsServer != nil {
			return httpsServer.Shutdown(context.Background())
		}
	case err := <-errChan:
		return err
	}
	return nil
}

func (a *ApiServer) startHttpServer(errChan chan error) *http.Server {
	server := &http.Server{
		Addr:    net.JoinHostPort(a.config.HTTP.Interface, strconv.FormatUint(uint64(a.config.HTTP.Port), 10)),
		Handler: a.mux,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				log.Info("shutting down server")
			default:
				log.Error(err, "could not start an HTTP Server")
				errChan <- err
			}
		}
	}()
	log.Info("starting", "interface", a.config.HTTP.Interface, "port", a.config.HTTP.Port)
	return server
}

func (a *ApiServer) startHttpsServer(errChan chan error) *http.Server {
	tlsConfig, err := configureMTLS(a.config.Auth.ClientCertsDir)
	if err != nil {
		errChan <- err
	}

	server := &http.Server{
		Addr:      net.JoinHostPort(a.config.HTTPS.Interface, strconv.FormatUint(uint64(a.config.HTTPS.Port), 10)),
		Handler:   a.mux,
		TLSConfig: tlsConfig,
	}

	go func() {
		err := server.ListenAndServeTLS(a.config.HTTPS.TlsCertFile, a.config.HTTPS.TlsKeyFile)
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				log.Info("shutting down server")
			default:
				log.Error(err, "could not start an HTTPS Server")
				errChan <- err
			}
		}
	}()
	log.Info("starting", "interface", a.config.HTTPS.Interface, "port", a.config.HTTPS.Port, "tls", true)
	return server
}

func configureMTLS(certsDir string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	if certsDir != "" {
		log.Info("loading client certificates")
		clientCertPool := x509.NewCertPool()
		files, err := ioutil.ReadDir(certsDir)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if !strings.HasSuffix(file.Name(), ".pem") && !strings.HasSuffix(file.Name(), ".crt") {
				log.Info("skipping file, all the client certificates has to have .pem or .crt extension", "file", file.Name())
				continue
			}
			log.Info("adding client certificate", "file", file.Name())
			path := filepath.Join(certsDir, file.Name())
			caCert, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, errors.Wrapf(err, "could not read certificate %s", path)
			}
			clientCertPool.AppendCertsFromPEM(caCert)
		}
		tlsConfig.ClientCAs = clientCertPool
	}
	tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven // client certs are required only for some endpoints
	return tlsConfig, nil
}

func (a *ApiServer) notAvailableHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte("" +
		"<!DOCTYPE html><html lang=en>" +
		"<head>\n<style>\n.center {\n  display: flex;\n  justify-content: center;\n  align-items: center;\n  height: 200px;\n  border: 3px solid green; \n}\n</style>\n</head>" +
		"<body><div class=\"center\"><strong>" +
		"GUI is disabled. If this is a Zone CP, please check the GUI on the Global CP." +
		"</strong></div></body>" +
		"</html>"))
	if err != nil {
		log.Error(err, "could not write the response")
	}
}

func SetupServer(rt runtime.Runtime) error {
	cfg := rt.Config()
	enableGUI := cfg.Mode != config_core.Zone
	if cfg.Mode != config_core.Standalone {
		for i, definition := range definitions.All {
			switch cfg.Mode {
			case config_core.Global:
				if definition.ResourceFactory().GetType() == mesh.DataplaneType {
					definitions.All[i].ReadOnly = true
				}
			case config_core.Zone:
				if definition.ResourceFactory().GetType() != mesh.DataplaneType {
					definitions.All[i].ReadOnly = true
				}
			}
		}
	}
	apiServer, err := NewApiServer(rt.ResourceManager(), rt.APIInstaller(), definitions.DefaultCRUDLEndpoints, &cfg, enableGUI, rt.Metrics())
	if err != nil {
		return err
	}
	return rt.Add(apiServer)
}
