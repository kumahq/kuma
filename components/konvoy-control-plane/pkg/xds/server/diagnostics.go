package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core"
	core_runtime "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/runtime"
)

var (
	diagnosticsServerLog = core.Log.WithName("xds-server").WithName("diagnostics")
)

type diagnosticsServer struct {
	port int
}

// Make sure that grpcServer implements all relevant interfaces
var (
	_ core_runtime.Component = &diagnosticsServer{}
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
	diagnosticsServerLog.Info("starting", "port", s.port)

	select {
	case <-stop:
		diagnosticsServerLog.Info("stopping")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}
