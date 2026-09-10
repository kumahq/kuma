package readiness

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bakito/go-log-logr-adapter/adapter"

	"github.com/kumahq/kuma/v3/app/kuma-dp/pkg/dataplane/httpclient"
	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
)

const (
	pathPrefixReady      = "/ready"
	stateReady           = "READY"
	stateNotReady        = "NOT_READY"
	stateTerminating     = "TERMINATING"
	dnsConfigGateTimeout = 15 * time.Second
)

// IdentityGate checks that Envoy loaded the identity certificate selected by the control plane.
type IdentityGate struct {
	client http.Client
	config atomic.Pointer[core_xds.IdentityReadinessConfig]
}

// NewIdentityGate creates an identity readiness gate backed by an Envoy Unix listener.
func NewIdentityGate(socketPath string) *IdentityGate {
	dialer := &net.Dialer{Timeout: 2 * time.Second}
	return &IdentityGate{client: http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, "unix", socketPath)
			},
			DisableKeepAlives: true,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // #nosec G402 -- local sentinel; peer certificate is checked below
				MinVersion:         tls.VersionTLS12,
			},
		},
	}}
}

// OnChange updates the identity certificate required for readiness.
func (g *IdentityGate) OnChange(_ context.Context, reader io.Reader) error {
	config := &core_xds.IdentityReadinessConfig{}
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(config); err != nil {
		return err
	}
	g.config.Store(config)
	return nil
}

// Ready reports whether Envoy serves the expected, unexpired identity certificate.
func (g *IdentityGate) Ready(ctx context.Context) (bool, error) {
	config := g.config.Load()
	if config == nil {
		return false, nil
	}
	if !config.Required {
		return true, nil
	}
	if config.ExpirationTime != nil && !time.Now().Before(*config.ExpirationTime) {
		return false, errors.New("identity certificate expired")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://localhost/ready", http.NoBody)
	if err != nil {
		return false, err
	}
	resp, err := g.client.Do(req) // #nosec G704 -- fixed URL over the generated Envoy Unix listener
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("identity listener returned status %d", resp.StatusCode)
	}
	if resp.TLS == nil || len(resp.TLS.PeerCertificates) == 0 {
		return false, errors.New("identity listener did not present a certificate")
	}
	certificate := resp.TLS.PeerCertificates[0]
	now := time.Now()
	if now.Before(certificate.NotBefore) || !now.Before(certificate.NotAfter) {
		return false, errors.New("identity listener presented an invalid certificate")
	}
	if config.CertificateHash != "" {
		hash := sha256.Sum256(certificate.Raw)
		if fmt.Sprintf("%x", hash) != config.CertificateHash {
			return false, errors.New("identity listener presented a stale certificate")
		}
	}
	return true, nil
}

// EnvoyAdmin locates the Envoy admin interface. SocketPath is set when admin
// runs over a Unix domain socket, otherwise Address and Port are.
type EnvoyAdmin struct {
	SocketPath string
	Address    string
	Port       uint32
}

func (e EnvoyAdmin) configured() bool {
	return e.SocketPath != "" || e.Port != 0
}

func (e EnvoyAdmin) readyURL() string {
	if e.SocketPath != "" {
		return "http://localhost/ready"
	}
	return "http://" + net.JoinHostPort(e.Address, strconv.Itoa(int(e.Port))) + "/ready"
}

// Reporter reports the health status of this Kuma Dataplane Proxy.
// The listener serves only /ready, which is proxied to Envoy admin so the pod
// is marked ready only after Envoy has its config. No other Envoy admin
// endpoint is exposed on this listener: the readiness port is reachable over
// the pod network (K8s probes), so proxying admin here would expose
// config_dump, certs, logging, etc. to any co-located or pod-network peer.
type Reporter struct {
	localListenAddr string
	localListenPort uint32
	envoyAdmin      EnvoyAdmin
	adminClient     *http.Client
	isTerminating   atomic.Bool
	// dnsConfigReady, when non-nil, blocks /ready until the DNS proxy has
	// loaded its first configuration from Envoy. Closed by the DNS proxy
	// server after the first successful ReloadMap call.
	dnsConfigReady    <-chan struct{}
	dnsConfigDeadline time.Time
	dnsBypassed       atomic.Bool
	identityGate      *IdentityGate
}

var logger = core.Log.WithName("readiness")

func NewReporter(localIPAddr string, localListenPort uint32, envoyAdmin EnvoyAdmin, dnsConfigReady <-chan struct{}) *Reporter {
	var deadline time.Time
	if dnsConfigReady != nil {
		deadline = time.Now().Add(dnsConfigGateTimeout)
	}
	return newReporterWithDeadline(localIPAddr, localListenPort, envoyAdmin, dnsConfigReady, deadline)
}

func newReporterWithDeadline(localIPAddr string, localListenPort uint32, envoyAdmin EnvoyAdmin, dnsConfigReady <-chan struct{}, dnsConfigDeadline time.Time) *Reporter {
	r := &Reporter{
		localListenPort:   localListenPort,
		localListenAddr:   localIPAddr,
		envoyAdmin:        envoyAdmin,
		dnsConfigReady:    dnsConfigReady,
		dnsConfigDeadline: dnsConfigDeadline,
	}
	if envoyAdmin.configured() {
		c := httpclient.NewTCPOrUDS(envoyAdmin.SocketPath, 2*time.Second, 3*time.Second)
		r.adminClient = &c
	}
	return r
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
	// Only /ready is served. Every other path returns 404. Do not add a
	// catch-all proxy to Envoy admin here: this listener is reachable over
	// the pod network, so it must not expose admin endpoints.
	mux.HandleFunc(pathPrefixReady, r.handleReadiness)
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

// SetIdentityGate enables identity certificate readiness checks.
func (r *Reporter) SetIdentityGate(gate *IdentityGate) {
	r.identityGate = gate
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

	if r.identityGate != nil {
		ready, err := r.identityGate.Ready(req.Context())
		if err != nil {
			logger.V(1).Info("identity certificate not ready", "err", err)
		}
		if !ready {
			r.writeState(writer, req, stateNotReady, http.StatusServiceUnavailable)
			return
		}
	}

	// Proxy /ready to Envoy admin so that the pod is only marked ready
	// after Envoy receives its config.
	if r.envoyAdmin.configured() {
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
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, r.envoyAdmin.readyURL(), http.NoBody)
	if err != nil {
		logger.Info("envoy admin not ready", "err", err)
		http.Error(writer, "envoy not ready", http.StatusServiceUnavailable)
		return
	}
	resp, err := r.adminClient.Do(req) // #nosec G704 -- fixed /ready URL on the Envoy admin endpoint from bootstrap
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

func (r *Reporter) NeedLeaderElection() bool {
	return false
}

var _ component.Component = &Reporter{}
