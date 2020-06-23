package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/runtime/component"
)

var (
	diagnosticsServerLog = core.Log.WithName("xds-server").WithName("diagnostics")
)

type diagnosticsServer struct {
	port int
}

func (s *diagnosticsServer) NeedLeaderElection() bool {
	return false
}

// Make sure that grpcServer implements all relevant interfaces
var (
	_ component.Component = &diagnosticsServer{}
)

func (s *diagnosticsServer) Start(stop <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/healthy", func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", s.port), Handler: mux}

	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err := httpServer.ListenAndServe(); err != nil {
			if err.Error() != "http: Server closed" {
				diagnosticsServerLog.Error(err, "terminated with an error")
				errChan <- err
				return
			}
		}
		diagnosticsServerLog.Info("terminated normally")
	}()
	diagnosticsServerLog.Info("starting", "interface", "0.0.0.0", "port", s.port)

	select {
	case <-stop:
		diagnosticsServerLog.Info("stopping")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}
