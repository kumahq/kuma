package diagnostics

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
	port           uint32
	metrics        metrics.Metrics
	debugEndpoints bool
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
	if s.debugEndpoints {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", s.port), Handler: mux}

	diagnosticsServerLog.Info("starting diagnostic server", "interface", "0.0.0.0", "port", s.port)
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

	select {
	case <-stop:
		diagnosticsServerLog.Info("stopping")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}
