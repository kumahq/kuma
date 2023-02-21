package diagnostics

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	pprof "net/http/pprof"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	diagnostics_config "github.com/kumahq/kuma/pkg/config/diagnostics"
	config_types "github.com/kumahq/kuma/pkg/config/types"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/metrics"
)

var (
	diagnosticsServerLog = core.Log.WithName("xds-server").WithName("diagnostics")
)

type diagnosticsServer struct {
	config  *diagnostics_config.DiagnosticsConfig
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
	if s.config.DebugEndpoints {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", s.config.ServerPort), Handler: mux}

	if s.config.TlsEnabled {
		cert, err := tls.LoadX509KeyPair(s.config.TlsCertFile, s.config.TlsKeyFile)
		if err != nil {
			return errors.Wrap(err, "failed to load TLS certificate")
		}
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		if tlsConfig.MinVersion, err = config_types.TLSVersion(s.config.TlsMinVersion); err != nil {
			return err
		}
		if tlsConfig.MaxVersion, err = config_types.TLSVersion(s.config.TlsMaxVersion); err != nil {
			return err
		}
		if tlsConfig.CipherSuites, err = config_types.TLSCiphers(s.config.TlsCipherSuites); err != nil {
			return err
		}
		httpServer.TLSConfig = tlsConfig
	}

	diagnosticsServerLog.Info("starting diagnostic server", "interface", "0.0.0.0", "port", s.config.ServerPort, "tls", s.config.TlsEnabled)
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		var err error
		if s.config.TlsEnabled {
			err = httpServer.ListenAndServeTLS(s.config.TlsCertFile, s.config.TlsKeyFile)
		} else {
			err = httpServer.ListenAndServe()
		}
		if err != nil {
			switch err {
			case http.ErrServerClosed:
				diagnosticsServerLog.Info("shutting down server")
			default:
				if s.config.TlsEnabled {
					diagnosticsServerLog.Error(err, "could not start HTTPS Server")
				} else {
					diagnosticsServerLog.Error(err, "could not start HTTP Server")
				}
				errChan <- err
			}
			return
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
