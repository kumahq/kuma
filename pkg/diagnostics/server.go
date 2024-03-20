package diagnostics

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	diagnostics_config "github.com/kumahq/kuma/pkg/config/diagnostics"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/metrics"
	kuma_srv "github.com/kumahq/kuma/pkg/util/http/server"
)

var diagnosticsServerLog = core.Log.WithName("xds-server").WithName("diagnostics")

type diagnosticsServer struct {
	isReady func() bool
	config  *diagnostics_config.DiagnosticsConfig
	metrics metrics.Metrics
	ready   atomic.Bool
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
		if s.isReady() {
			resp.WriteHeader(http.StatusOK)
		} else {
			resp.WriteHeader(http.StatusServiceUnavailable)
		}
	})
	mux.HandleFunc("/healthy", func(resp http.ResponseWriter, _ *http.Request) {
		resp.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.InstrumentMetricHandler(s.metrics, promhttp.HandlerFor(s.metrics, promhttp.HandlerOpts{})))
	if s.config.DebugEndpoints {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}
	var tlsConfig *tls.Config
	if s.config.TlsEnabled {
		cert, err := tls.LoadX509KeyPair(s.config.TlsCertFile, s.config.TlsKeyFile)
		if err != nil {
			return errors.Wrap(err, "failed to load TLS certificate")
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12, // Make gosec pass, (In practice it's always set after).
		}
		if tlsConfig.MinVersion, err = config_types.TLSVersion(s.config.TlsMinVersion); err != nil {
			return err
		}
		if tlsConfig.MaxVersion, err = config_types.TLSVersion(s.config.TlsMaxVersion); err != nil {
			return err
		}
		if tlsConfig.CipherSuites, err = config_types.TLSCiphers(s.config.TlsCipherSuites); err != nil {
			return err
		}
	}

	httpServer := &http.Server{
		TLSConfig:         tlsConfig,
		Addr:              fmt.Sprintf(":%d", s.config.ServerPort),
		Handler:           mux,
		ReadHeaderTimeout: time.Second,
		ErrorLog:          adapter.ToStd(diagnosticsServerLog),
	}
	errChan := make(chan error)
	if err := kuma_srv.StartServer(diagnosticsServerLog, httpServer, &s.ready, errChan); err != nil {
		return err
	}

	select {
	case <-stop:
		s.ready.Store(false)
		diagnosticsServerLog.Info("stopping")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		s.ready.Store(false)
		return err
	}
}
