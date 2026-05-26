package readiness

import (
	"context"
	"errors"
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
	tproxy_validate "github.com/kumahq/kuma/v2/pkg/transparentproxy/validate"
)

const (
	pathPrefixReady      = "/ready"
	stateReady           = "READY"
	stateNotReady        = "NOT_READY"
	stateTerminating     = "TERMINATING"
	stateRedirectFailed  = "INBOUND_REDIRECT_FAILED"
	dnsConfigGateTimeout = 15 * time.Second
	tproxyCheckInterval  = 10 * time.Second
	tproxyDialTimeout    = 100 * time.Millisecond
)

// Reporter reports the health status of this Kuma Dataplane Proxy.
// When adminSocketPath is set, it also reverse-proxies Envoy admin
// endpoints so that external tools can reach admin over TCP even when
// Envoy admin listens on a Unix domain socket. Mutating endpoints
// (/quitquitquit, /drain_listeners, /runtime_modify) are blocked.
type Reporter struct {
	localListenAddr string
	localListenPort uint32
	adminSocketPath string
	adminClient     *http.Client
	isTerminating   atomic.Bool
	// dnsConfigReady, when non-nil, blocks /ready until the DNS proxy has
	// loaded its first configuration from Envoy. Closed by the DNS proxy
	// server after the first successful ReloadMap call.
	dnsConfigReady    <-chan struct{}
	dnsConfigDeadline time.Time
	dnsBypassed       atomic.Bool
	// tproxyCheckEnabled gates the periodic inbound-redirect self-test.
	// When enabled, the readiness handler dials the redirect path that
	// `pkg/transparentproxy/validate.SelfTest` exercises (127.0.0.6 on
	// IPv4, ::6 on IPv6 — both rewritten by KUMA_MESH_OUTBOUND for
	// UID 5678 traffic on lo). A connect failure indicates the netns
	// iptables redirect chain is gone (FTI-7529 / K8S-5010 z977j: long-
	// running pods whose iptables have been wiped by an external
	// host-side process keep accepting Service traffic but silently
	// drop mesh requests with WRONG_VERSION_NUMBER on the peer side).
	tproxyCheckEnabled bool
	tproxyUseIPv6      bool
	tproxyHealthy      atomic.Bool
	tproxyLastChecked  atomic.Int64
	// tproxyProbe is the actual TCP self-test. Defaults to
	// `tproxy_validate.SelfTest`; tests override it to simulate
	// intact-vs-broken iptables without needing a real netns.
	tproxyProbe func(useIPv6 bool, timeout time.Duration) error
}

var logger = core.Log.WithName("readiness")

func NewReporter(localIPAddr string, localListenPort uint32, adminSocketPath string, dnsConfigReady <-chan struct{}) *Reporter {
	var deadline time.Time
	if dnsConfigReady != nil {
		deadline = time.Now().Add(dnsConfigGateTimeout)
	}
	return newReporterWithDeadline(localIPAddr, localListenPort, adminSocketPath, dnsConfigReady, deadline)
}

func newReporterWithDeadline(localIPAddr string, localListenPort uint32, adminSocketPath string, dnsConfigReady <-chan struct{}, dnsConfigDeadline time.Time) *Reporter {
	r := &Reporter{
		localListenPort:   localListenPort,
		localListenAddr:   localIPAddr,
		adminSocketPath:   adminSocketPath,
		dnsConfigReady:    dnsConfigReady,
		dnsConfigDeadline: dnsConfigDeadline,
	}
	if adminSocketPath != "" {
		c := httpclient.NewUDS(adminSocketPath, 2*time.Second, 3*time.Second)
		r.adminClient = &c
	}
	r.tproxyHealthy.Store(true)
	return r
}

// EnableTproxyCheck turns on the periodic inbound-redirect self-test.
// useIPv6 selects the v6 redirect target (::6 -> KUMA_MESH_INBOUND_REDIRECT)
// instead of the v4 one (127.0.0.6); the choice should follow the family
// the dataplane is using.
func (r *Reporter) EnableTproxyCheck(useIPv6 bool) {
	r.tproxyCheckEnabled = true
	r.tproxyUseIPv6 = useIPv6
	if r.tproxyProbe == nil {
		r.tproxyProbe = tproxy_validate.SelfTest
	}
}

func (r *Reporter) Start(stop <-chan struct{}) error {
	// Use "tcp" (not "tcp6") so that when localListenAddr is the IPv6
	// wildcard "::", Linux gives a dual-stack listener that accepts
	// both IPv4 and IPv6 probes. "tcp6" sets IPV6_V6ONLY and would
	// refuse IPv4 probes. For concrete addresses the family is
	// determined by the address itself.
	addr := fmt.Sprintf("%s:%d", r.localListenAddr, r.localListenPort)
	if govalidator.IsIPv6(r.localListenAddr) {
		addr = fmt.Sprintf("[%s]:%d", r.localListenAddr, r.localListenPort)
	}
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
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
	// Mutating endpoints are blocked by path denylist.
	if r.adminSocketPath != "" {
		mux.Handle("/", r.adminProxy())
	}
	server := &http.Server{
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		Handler:           mux,
		ErrorLog:          adapter.ToStd(logger),
	}

	errCh := make(chan error, 1)
	go func() {
		// ErrServerClosed is returned after Shutdown is called; it is not an
		// actual error and must not be forwarded to avoid blocking the goroutine
		// on an already-abandoned errCh.
		if err := server.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	// Gate readiness on DNS proxy receiving its first config from Envoy.
	// This ensures mesh-generated DNS names resolve before app containers start.
	// After dnsConfigGateTimeout we bypass the gate and log a warning to
	// avoid blocking deploys when misconfigured.
	if r.dnsConfigReady != nil && !r.dnsBypassed.Load() {
		select {
		case <-r.dnsConfigReady:
		default:
			if time.Now().After(r.dnsConfigDeadline) {
				if r.dnsBypassed.CompareAndSwap(false, true) {
					logger.Info("[WARNING] DNS proxy config not received within timeout, bypassing readiness gate",
						"timeout", dnsConfigGateTimeout)
				}
			} else {
				r.writeState(writer, req, stateNotReady, http.StatusServiceUnavailable)
				return
			}
		}
	}

	if r.tproxyCheckEnabled && !r.checkTproxyHealthy() {
		r.writeState(writer, req, stateRedirectFailed, http.StatusServiceUnavailable)
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

// checkTproxyHealthy returns true when the transparent-proxy inbound
// redirect chain is intact. Result is cached for tproxyCheckInterval to
// keep readiness probes cheap (default kubelet period is 5s). The probe
// itself is `pkg/transparentproxy/validate.SelfTest` so the canonical
// dial target lives in one place; do not inline the dial here.
func (r *Reporter) checkTproxyHealthy() bool {
	now := time.Now().UnixNano()
	last := r.tproxyLastChecked.Load()
	if last != 0 && time.Duration(now-last) < tproxyCheckInterval {
		return r.tproxyHealthy.Load()
	}
	if !r.tproxyLastChecked.CompareAndSwap(last, now) {
		// another goroutine is refreshing; use the cached value
		return r.tproxyHealthy.Load()
	}

	err := r.tproxyProbe(r.tproxyUseIPv6, tproxyDialTimeout)
	healthy := err == nil
	prev := r.tproxyHealthy.Swap(healthy)
	if prev && !healthy {
		logger.Info("[WARNING] transparent proxy self-test failed; netns iptables redirect to inbound listener is missing or broken; reporting NOT_READY",
			"err", err)
	} else if !prev && healthy {
		logger.Info("transparent proxy self-test recovered")
	}
	return healthy
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
