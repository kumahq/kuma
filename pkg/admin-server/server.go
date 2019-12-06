package admin_server

import (
	"context"
	"fmt"
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

const port = 5692

type AdminServer struct {
	container *restful.Container
}

func NewAdminServer(services ...*restful.WebService) *AdminServer {
	container := restful.NewContainer()
	for _, service := range services {
		container.Add(service)
	}
	return &AdminServer{container}
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
		Addr:    fmt.Sprintf("127.0.0.1:%d", port),
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
	log.Info("starting server", "port", port)
	return server, errChan
}

func SetupServer(rt runtime.Runtime) error {
	ws := rest.NewWebservice(rt.ProvidedCaManager())
	srv := NewAdminServer(ws)
	return rt.Add(srv)
}
