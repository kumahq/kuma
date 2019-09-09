package server

import (
	"context"
	"fmt"
	"net/http"

	envoy_xds "github.com/envoyproxy/go-control-plane/pkg/server"

	"github.com/Kong/kuma/pkg/core"
	core_runtime "github.com/Kong/kuma/pkg/core/runtime"
)

var (
	httpServerLog = core.Log.WithName("xds-server").WithName("http")
)

type httpGateway struct {
	srv  envoy_xds.Server
	port int
}

// Make sure that httpGateway implements all relevant interfaces
var (
	_ core_runtime.Component = &httpGateway{}
)

func (g *httpGateway) Start(stop <-chan struct{}) error {
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", g.port), Handler: &envoy_xds.HTTPGateway{Server: g.srv}}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err := httpServer.ListenAndServe(); err != nil {
			if err.Error() != "http: Server closed" {
				httpServerLog.Error(err, "terminated with an error")
				errChan <- err
				return
			}
		}
		httpServerLog.Info("terminated normally")
	}()
	httpServerLog.Info("starting", "port", g.port)

	select {
	case <-stop:
		httpServerLog.Info("stopping")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}
