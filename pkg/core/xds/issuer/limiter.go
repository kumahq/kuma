// Package issuer provides throttling for certificate issuance shared between
// the legacy mTLS path (pkg/xds/secrets) and MeshIdentity
// (pkg/core/resources/apis/meshidentity/providers).
package issuer

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/v3/pkg/core"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/metrics"
)

// DefaultCircuitBreakerMinProxies is the default number of distinct proxies that
// must fail before a backend's circuit opens. Kept well above 1 so a single
// misbehaving proxy can never trip a whole backend.
const DefaultCircuitBreakerMinProxies = 10

// CertBackoff returns a factory for the per-proxy backoff applied after a
// failure: exponential from base, capped at max, with jitter so proxies don't
// retry in lockstep. A non-positive base is clamped to avoid a panic.
func CertBackoff(base, maxBackoff time.Duration) func() retry.Backoff {
	if base <= 0 {
		core.Log.WithName("xds").WithName("issuer").Info(
			"configured certificate generation base backoff is not positive, using 1s instead", "configured", base)
		base = time.Second
	}
	if maxBackoff < base {
		maxBackoff = base
	}
	return func() retry.Backoff {
		return retry.WithJitter(base, retry.WithCappedDuration(maxBackoff, retry.NewExponential(base)))
	}
}

// Config configures a Limiter.
type Config struct {
	// NewBackoff builds the per-proxy backoff applied after a failure.
	NewBackoff func() retry.Backoff
	// MinProxies is the number of distinct proxies that must fail within Window
	// before the whole backend circuit opens. 0 disables the circuit breaker,
	// leaving only per-proxy backoff.
	MinProxies int
	// Window is how long a proxy failure counts toward the MinProxies trip. It
	// should be at least the max backoff so a proxy that is still backing off
	// keeps being counted.
	Window time.Duration
	// Cooldown is how long the circuit stays open before a half-open probe.
	Cooldown time.Duration
}

// Limiter throttles certificate issuance. It applies per-proxy exponential
// backoff for isolated failures and, on top of that, a per-backend circuit
// breaker that trips only when failures span many distinct proxies - so a
// single misbehaving proxy can never trip a whole backend.
//
// Keys are kri.Identifier: the backend is the issuer (a CA backend or a
// MeshIdentity), the proxy is what the certificate is issued for.
type Limiter struct {
	cfg Config

	sync.Mutex
	backends map[kri.Identifier]*backendState

	backingOffMetric  prometheus.Gauge
	circuitOpenMetric prometheus.Gauge
}

type backendState struct {
	proxies       map[kri.Identifier]*proxyState
	openUntil     time.Time
	probing       bool
	probeDeadline time.Time
}

type proxyState struct {
	backoff     retry.Backoff
	nextTry     time.Time
	lastFailure time.Time
}

func NewLimiter(cfg Config, m metrics.Metrics) (*Limiter, error) {
	backingOff := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cert_generation_backoff",
		Help: "Number of proxies currently backing off certificate generation after a failure",
	})
	if err := m.Register(backingOff); err != nil {
		return nil, err
	}
	circuitOpen := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cert_generation_circuit_open",
		Help: "Number of backends whose certificate issuance circuit breaker is currently open",
	})
	if err := m.Register(circuitOpen); err != nil {
		return nil, err
	}
	return &Limiter{
		cfg:               cfg,
		backends:          map[kri.Identifier]*backendState{},
		backingOffMetric:  backingOff,
		circuitOpenMetric: circuitOpen,
	}, nil
}

// Allow reports whether issuance for proxy against backend may proceed now. If
// not, it returns how long to wait. Every permitted attempt must be followed by
// a call to Record. A nil Limiter allows everything (no throttling).
func (l *Limiter) Allow(backend, proxy kri.Identifier) (bool, time.Duration) {
	if l == nil {
		return true, 0
	}
	l.Lock()
	defer l.Unlock()

	b := l.backends[backend]
	if b == nil {
		return true, 0
	}

	now := core.Now()
	if b.openUntil.After(now) { // circuit open
		return false, b.openUntil.Sub(now)
	}
	if !b.openUntil.IsZero() { // cooldown elapsed - half-open, allow a single probe
		if b.probing && b.probeDeadline.After(now) {
			return false, b.probeDeadline.Sub(now)
		}
		b.probing = true
		b.probeDeadline = now.Add(l.cfg.Cooldown)
		return true, 0
	}
	if p := b.proxies[proxy]; p != nil && p.nextTry.After(now) { // per-proxy backoff
		return false, p.nextTry.Sub(now)
	}
	return true, 0
}

// Record feeds back the outcome of an attempt that Allow permitted.
func (l *Limiter) Record(backend, proxy kri.Identifier, success bool) {
	if l == nil {
		return
	}
	l.Lock()
	defer l.Unlock()

	now := core.Now()
	b := l.backends[backend]
	if b == nil {
		b = &backendState{proxies: map[kri.Identifier]*proxyState{}}
		l.backends[backend] = b
	}

	if success {
		delete(b.proxies, proxy)
		b.openUntil = time.Time{} // any success closes the circuit
		b.probing = false
		if len(b.proxies) == 0 {
			delete(l.backends, backend)
		}
		l.updateMetricsLocked()
		return
	}

	p := b.proxies[proxy]
	if p == nil {
		p = &proxyState{backoff: l.cfg.NewBackoff()}
		b.proxies[proxy] = p
	}
	delay, _ := p.backoff.Next()
	p.nextTry = now.Add(delay)
	p.lastFailure = now

	if b.probing { // half-open probe failed - reopen
		b.probing = false
		b.openUntil = now.Add(l.cfg.Cooldown)
		l.updateMetricsLocked()
		return
	}

	// Count distinct proxies that failed within the window, pruning stale ones.
	// The circuit opens only once failures span MinProxies distinct proxies, so
	// a single proxy retrying can never trip it.
	distinct := 0
	for k, ps := range b.proxies {
		if now.Sub(ps.lastFailure) > l.cfg.Window {
			delete(b.proxies, k)
			continue
		}
		distinct++
	}
	if l.cfg.MinProxies > 0 && distinct >= l.cfg.MinProxies {
		b.openUntil = now.Add(l.cfg.Cooldown)
	}
	l.updateMetricsLocked()
}

func (l *Limiter) updateMetricsLocked() {
	now := core.Now()
	backingOff := 0
	open := 0
	for _, b := range l.backends {
		if b.openUntil.After(now) {
			open++
		}
		for _, p := range b.proxies {
			if p.nextTry.After(now) {
				backingOff++
			}
		}
	}
	l.backingOffMetric.Set(float64(backingOff))
	l.circuitOpenMetric.Set(float64(open))
}
