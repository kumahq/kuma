package readiness

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bakito/go-log-logr-adapter/adapter"

	"github.com/kumahq/kuma/v2/app/kuma-dp/pkg/dataplane/httpclient"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
)

const (
	pathPrefixReady  = "/ready"
	stateReady       = "READY"
	stateTerminating = "TERMINATING"
)

// Reporter reports the health status of this Kuma Dataplane Proxy.
// When adminSocketPath is set, it also reverse-proxies Envoy admin
// endpoints so that external tools can reach admin over TCP even when
// Envoy admin listens on a Unix domain socket.
type Reporter struct {
	unixSocketDisabled bool
	socketDir          string
	localListenAddr    string
	localListenPort    uint32
	adminSocketPath    string
	adminClient        *http.Client
	isTerminating      atomic.Bool
}

var logger = core.Log.WithName("readiness")

func NewReporter(unixSocketDisabled bool, socketDir string, localIPAddr string, localListenPort uint32, adminSocketPath string) *Reporter {
	r := &Reporter{
		unixSocketDisabled: unixSocketDisabled,
		socketDir:          socketDir,
		localListenPort:    localListenPort,
		localListenAddr:    localIPAddr,
		adminSocketPath:    adminSocketPath,
	}
	if adminSocketPath != "" {
		c := httpclient.NewUDS(adminSocketPath, 2*time.Second, 3*time.Second)
		r.adminClient = &c
	}
	return r
}

func (r *Reporter) Start(stop <-chan struct{}) error {
	var lis net.Listener
	var protocol, addr string
	if r.unixSocketDisabled {
		if govalidator.IsIPv6(r.localListenAddr) {
			protocol = "tcp6"
			addr = fmt.Sprintf("[%s]:%d", r.localListenAddr, r.localListenPort)
		} else {
			protocol = "tcp"
			addr = fmt.Sprintf("%s:%d", r.localListenAddr, r.localListenPort)
		}
	} else {
		protocol = "unix"
		addr = core_xds.ReadinessReporterSocketName(r.socketDir)
	}
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), protocol, addr)
	if err != nil {
		return err
	}

	defer func() {
		_ = lis.Close()
	}()

	logger.Info("starting readiness reporter", "addr", lis.Addr().String())

	mux := http.NewServeMux()
	mux.HandleFunc(pathPrefixReady, r.handleReadiness)
	// When admin is on UDS, reverse-proxy read-only admin endpoints so
	// that operators can still reach /stats, /config_dump, etc.
	// Only GET and HEAD are allowed to prevent mutating endpoints
	// (/quitquitquit, /runtime_modify, /drain_listeners, etc.) from
	// being reachable over TCP.
	if r.adminSocketPath != "" {
		mux.Handle("/", r.adminProxy())
	}
	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		Handler:           mux,
		ErrorLog:          adapter.ToStd(logger),
	}

	errCh := make(chan error)
	go func() {
		if err := server.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-stop:
		logger.Info("stopping readiness reporter")
		return server.Shutdown(context.Background())
	}
}

func (r *Reporter) Terminating() {
	r.isTerminating.Store(true)
}

func (r *Reporter) handleReadiness(writer http.ResponseWriter, req *http.Request) {
	if r.isTerminating.Load() {
		r.writeState(writer, req, stateTerminating, http.StatusServiceUnavailable)
		return
	}

	// When admin is on UDS, proxy /ready to Envoy admin so that
	// the pod is only marked ready after Envoy receives its config.
	if r.adminSocketPath != "" {
		r.proxyAdminReady(writer)
		return
	}

	r.writeState(writer, req, stateReady, http.StatusOK)
}

func (r *Reporter) writeState(writer http.ResponseWriter, req *http.Request, state string, status int) {
	stateBytes := []byte(state)
	writer.Header().Set("content-type", "text/plain")
	writer.Header().Set("content-length", fmt.Sprintf("%d", len(stateBytes)))
	writer.Header().Set("cache-control", "no-cache, max-age=0")
	writer.Header().Set("x-powered-by", "kuma-dp")
	writer.WriteHeader(status)
	_, err := writer.Write(stateBytes)
	logger.V(1).Info("responding readiness state", "state", state, "client", req.RemoteAddr)
	if err != nil {
		logger.Info("[WARNING] could not write response", "err", err)
	}
}

func (r *Reporter) proxyAdminReady(writer http.ResponseWriter) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://localhost/ready", http.NoBody)
	if err != nil {
		logger.Info("envoy admin not ready", "err", err)
		http.Error(writer, "envoy not ready", http.StatusServiceUnavailable)
		return
	}
	resp, err := r.adminClient.Do(req)
	if err != nil {
		logger.Info("envoy admin not ready", "err", err)
		http.Error(writer, "envoy not ready", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Info("[WARNING] could not read admin ready response body", "err", err)
	}
	for k, vals := range resp.Header {
		for _, v := range vals {
			writer.Header().Add(k, v)
		}
	}
	writer.WriteHeader(resp.StatusCode)
	if _, err := writer.Write(body); err != nil {
		logger.Info("[WARNING] could not write response", "err", err)
	}
}

func (r *Reporter) adminProxy() http.Handler {
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "localhost"
		},
		Transport: r.adminClient.Transport,
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, err error) {
			logger.Error(err, "admin proxy error")
			http.Error(w, "admin backend unavailable", http.StatusBadGateway)
		},
		ErrorLog: adapter.ToStd(logger),
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Block endpoints that terminate or destabilize Envoy.
		// Read endpoints (/stats, /config_dump, /clusters, etc.) accept
		// both GET and POST in Envoy's admin API, so we restrict by path
		// rather than method to preserve operator and test-framework access.
		switch req.URL.Path {
		case "/quitquitquit", "/drain_listeners", "/runtime_modify":
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		proxy.ServeHTTP(w, req)
	})
}

func (r *Reporter) NeedLeaderElection() bool {
	return false
}

var _ component.Component = &Reporter{}
