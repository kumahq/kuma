package server

import (
	"context"
	"fmt"
	"net/http"
	pprof "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/metrics"
)

var (
	diagnosticsServerLog = core.Log.WithName("xds-server").WithName("diagnostics")
)

type diagnosticsServer struct {
	port    int
	metrics metrics.Metrics
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
	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(s.metrics, promhttp.HandlerFor(s.metrics, promhttp.HandlerOpts{})))
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

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
