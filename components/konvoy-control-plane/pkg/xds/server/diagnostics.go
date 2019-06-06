package server

import (
	"context"
	"fmt"
	"net/http"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	diagnosticsServerLog = ctrl.Log.WithName("xds-server").WithName("diagnostics")
)

type diagnosticsServer struct {
	port int
}

// Make sure that grpcServer implements all relevant interfaces
var (
	_ manager.Runnable               = &diagnosticsServer{}
	_ manager.LeaderElectionRunnable = &diagnosticsServer{}
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

func (s *diagnosticsServer) NeedLeaderElection() bool {
	return false
}
