package server

import (
	"context"
	"fmt"
	"net/http"

	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	httpServerLog = ctrl.Log.WithName("xds-server").WithName("http")
)

type httpGateway struct {
	srv  xds.Server
	port int
}

// Make sure that httpGateway implements all relevant interfaces
var (
	_ manager.Runnable               = &httpGateway{}
	_ manager.LeaderElectionRunnable = &httpGateway{}
)

func (g *httpGateway) Start(stop <-chan struct{}) error {
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", g.port), Handler: &xds.HTTPGateway{Server: g.srv}}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err := httpServer.ListenAndServe(); err != nil {
			httpServerLog.Error(err, "terminated with an error")
			errChan <- err
		}
	}()
	httpServerLog.Info("starting", "port", g.port)

	select {
	case <-stop:
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

func (g *httpGateway) NeedLeaderElection() bool {
	return false
}
