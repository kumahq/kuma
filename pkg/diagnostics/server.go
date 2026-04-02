package diagnostics

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync/atomic"
	"time"

	"github.com/bakito/go-log-logr-adapter/adapter"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	diagnostics_config "github.com/kumahq/kuma/v2/pkg/config/diagnostics"
	config_types "github.com/kumahq/kuma/v2/pkg/config/types"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	kuma_log "github.com/kumahq/kuma/v2/pkg/log"
	"github.com/kumahq/kuma/v2/pkg/metrics"
	kuma_srv "github.com/kumahq/kuma/v2/pkg/util/http/server"
)

var diagnosticsServerLog = core.Log.WithName("xds-server").WithName("diagnostics")

type diagnosticsServer struct {
	isReady  func() bool
	config   *diagnostics_config.DiagnosticsConfig
	metrics  metrics.Metrics
	registry *kuma_log.ComponentLevelRegistry
	ready    atomic.Bool
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
	AddLoggingHandlers(mux, s.registry)
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

type loggingResponse struct {
	Components map[string]string `json:"components"`
}

type setLevelRequest struct {
	Component string `json:"component"`
	Level     string `json:"level"`
}

// AddLoggingHandlers registers GET/PUT/DELETE /logging routes on mux.
func AddLoggingHandlers(mux *http.ServeMux, registry *kuma_log.ComponentLevelRegistry) {
	mux.HandleFunc("/logging", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleLoggingGet(w, registry)
		case http.MethodPut:
			handleLoggingPut(w, r, registry)
		case http.MethodDelete:
			overrides := registry.ResetAll()
			diagnosticsServerLog.Info("all component log level overrides reset", "count", len(overrides))
			writeOverridesResponse(w, overrides)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	})
	mux.HandleFunc("/logging/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		component := r.URL.Path[len("/logging/"):]
		if err := kuma_log.ValidateComponentName(component); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		registry.ResetLevel(component)
		diagnosticsServerLog.Info("component log level override reset", "component", component)
	})
}

func writeOverridesResponse(w http.ResponseWriter, overrides map[string]kuma_log.LogLevel) {
	components := make(map[string]string, len(overrides))
	for name, level := range overrides {
		components[name] = level.String()
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(loggingResponse{Components: components}); err != nil {
		diagnosticsServerLog.Error(err, "could not write logging response")
	}
}

func handleLoggingGet(w http.ResponseWriter, registry *kuma_log.ComponentLevelRegistry) {
	overrides := registry.ListOverrides()
	components := make(map[string]string, len(overrides))
	for name, level := range overrides {
		components[name] = level.String()
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(loggingResponse{Components: components}); err != nil {
		diagnosticsServerLog.Error(err, "could not write logging response")
	}
}

func handleLoggingPut(w http.ResponseWriter, r *http.Request, registry *kuma_log.ComponentLevelRegistry) {
	const maxBodySize = 4096 // generous for {"component":"...","level":"..."}
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	var req setLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := kuma_log.ValidateComponentName(req.Component); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	level, err := kuma_log.ParseLogLevel(req.Level)
	if err != nil {
		http.Error(w, "invalid log level: "+err.Error(), http.StatusBadRequest)
		return
	}
	if err := registry.SetLevel(req.Component, level); err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}
	diagnosticsServerLog.Info("component log level override set", "component", req.Component, "level", req.Level)
	w.WriteHeader(http.StatusOK)
}
