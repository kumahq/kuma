package admin_server

import (
	"context"
	"fmt"
	admin_server "github.com/Kong/kuma/pkg/config/admin-server"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/ca/provided/rest"
	"github.com/Kong/kuma/pkg/core/runtime"
	"github.com/emicklei/go-restful"
	"go.uber.org/multierr"
	"net/http"
)

var (
	log = core.Log.WithName("admin-server")
)

type AdminServer struct {
	cfg       admin_server.AdminServerConfig
	container *restful.Container
}

func NewAdminServer(cfg admin_server.AdminServerConfig, services ...*restful.WebService) *AdminServer {
	container := restful.NewContainer()
	for _, service := range services {
		container.Add(service)
	}
	return &AdminServer{
		cfg:       cfg,
		container: container,
	}
}

func (a *AdminServer) Start(stop <-chan struct{}) error {
	httpServer, httpErrChan := a.startHttpServer()

	select {
	case <-stop:
		log.Info("stopping")
		var multiErr error
		if err := httpServer.Shutdown(context.Background()); err != nil {
			multiErr = multierr.Combine(err)
		}
		return multiErr
	case err := <-httpErrChan:
		return err
	}
}

func (a *AdminServer) startHttpServer() (*http.Server, chan error) {
	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", a.cfg.Local.Port),
		Handler: a.container,
	}

	errChan := make(chan error)

	go func() {
		defer close(errChan)
		if err := server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				log.Error(err, "http server terminated with an error")
				errChan <- err
				return
			}
		}
		log.Info("http server terminated normally")
	}()
	log.Info("starting server", "port", a.cfg.Local.Port)
	return server, errChan
}

func SetupServer(rt runtime.Runtime) error {
	ws := rest.NewWebservice(rt.ProvidedCaManager())
	srv := NewAdminServer(*rt.Config().AdminServer, ws)
	return rt.Add(srv)
}
