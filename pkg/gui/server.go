package gui

import (
	"context"
	"fmt"
	gui_server "github.com/Kong/kuma/pkg/config/gui-server"
	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
	"net/http"
)

var log = core.Log.WithName("gui-server")

func SetupServer(rt core_runtime.Runtime) error {
	srv := Server{rt.Config().GuiServer}
	if err := core_runtime.Add(rt, &srv); err != nil {
		return err
	}
	return nil
}

type Server struct {
	Config *gui_server.GuiServerConfig
}

var _ core_runtime.Component = &Server{}

func (g *Server) Start(stop <-chan struct{}) error {
	fileServer := http.FileServer(GuiDir)

	guiServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", g.Config.Port),
		Handler: fileServer,
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
