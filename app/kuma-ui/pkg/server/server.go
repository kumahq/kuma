package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kong/kuma/app/kuma-ui/pkg/resources"
	"github.com/Kong/kuma/app/kuma-ui/pkg/server/types"
	gui_server "github.com/Kong/kuma/pkg/config/gui-server"
	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var log = core.Log.WithName("gui-server")

func SetupServer(rt core_runtime.Runtime) error {
	srv := Server{
		Config:        rt.Config().GuiServer,
		ApiServerPort: rt.Config().ApiServer.Port,
	}
	if err := core_runtime.Add(rt, &srv); err != nil {
		return err
	}
	return nil
}

type Server struct {
	Config        *gui_server.GuiServerConfig
	ApiServerPort int
}

var _ core_runtime.Component = &Server{}

func (g *Server) Start(stop <-chan struct{}) error {
	fileServer := http.FileServer(resources.GuiDir)

	mux := http.NewServeMux()
	mux.Handle("/", fileServer)
	apiHandler, err := g.apiHandler()
	if err != nil {
		return err
	}
	mux.Handle("/api/", apiHandler)
	mux.HandleFunc("/config", g.configHandler)

	guiServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", g.Config.Port),
		Handler: mux,
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		if err := guiServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Error(err, "terminated with an error")
				errChan <- err
				return
			}
		}
		log.Info("terminated normally")
	}()
	log.Info("starting", "port", g.Config.Port)

	select {
	case <-stop:
		log.Info("stopping")
		return guiServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (g *Server) apiHandler() (http.Handler, error) {
	apiServerUrl, err := url.Parse(fmt.Sprintf("http://localhost:%d", g.ApiServerPort))
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(apiServerUrl)
	proxy.Director = func(request *http.Request) {
		request.URL.Scheme = apiServerUrl.Scheme
		request.URL.Host = apiServerUrl.Host
		// strip the /api prefix
		request.URL.Path = strings.TrimPrefix(request.URL.Path, "/api")
	}
	return proxy, nil
}

func (g *Server) configHandler(writer http.ResponseWriter, request *http.Request) {
	bytes, err := json.Marshal(fromServerConfig(*g.Config.GuiConfig))
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
