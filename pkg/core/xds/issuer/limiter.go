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
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/metrics"
)

// DefaultCircuitBreakerMinProxies is the default number of distinct proxies that
// must fail before a backend's circuit opens. Kept well above 1 so a single
// misbehaving proxy can never trip a whole backend.
const DefaultCircuitBreakerMinProxies = 10

// defaultCooldown is the fallback used when a caller passes a non-positive
// Cooldown (e.g. derived from a misconfigured backoff), so the circuit breaker
// can't end up with a zero/negative window.
const defaultCooldown = time.Minute

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

// Limiter throttles certificate issuance. The backend (a CA backend or a
// MeshIdentity) is keyed by kri.Identifier; the proxy is keyed by its
// model.ResourceKey - the identity the callers and the cleanup path already
// carry. Use NewLimiter for the real implementation or Unlimited for a no-op.
type Limiter interface {
	// Allow reports whether issuance for proxy against backend may proceed now.
	// If not, it returns how long to wait. Every permitted attempt must be
	// followed by a call to Record.
	Allow(backend kri.Identifier, proxy model.ResourceKey) (bool, time.Duration)
	// Record feeds back the outcome of an attempt that Allow permitted.
	Record(backend kri.Identifier, proxy model.ResourceKey, success bool)
	// Forget drops all per-proxy backoff state for a proxy across every backend,
	// e.g. when a proxy disconnects.
	Forget(proxy model.ResourceKey)
}

// Unlimited returns a Limiter that never throttles, for callers (e.g. tests)
// that don't need issuance limiting.
func Unlimited() Limiter { return unlimited{} }

type unlimited struct{}

func (unlimited) Allow(kri.Identifier, model.ResourceKey) (bool, time.Duration) { return true, 0 }
func (unlimited) Record(kri.Identifier, model.ResourceKey, bool)                {}
func (unlimited) Forget(model.ResourceKey)                                      {}

// limiter applies per-proxy exponential backoff for isolated failures and, on
// top of that, a per-backend circuit breaker that trips only when failures span
// many distinct proxies - so a single misbehaving proxy can never trip a whole
// backend.
//
// Off-the-shelf circuit breakers (sony/gobreaker, failsafe-go, ...) don't fit:
// they count requests/consecutive failures per breaker, so one proxy retrying N
// times would trip the whole backend - exactly what the distinct-proxy counting
// here avoids. They also have no per-key (per-proxy) sub-state, so we'd keep the
// backoff map, the distinct-proxy windowing and Forget ourselves and use the
// library only for the state machine. Not worth a dependency for that.
type limiter struct {
	cfg Config

	sync.Mutex
	backends map[kri.Identifier]*backendState
}

var (
	_ Limiter = unlimited{}
	_ Limiter = (*limiter)(nil)
)

type backendState struct {
	proxies       map[model.ResourceKey]*proxyState
	openUntil     time.Time
	probing       bool
	probeDeadline time.Time
}

type proxyState struct {
	backoff     retry.Backoff
	nextTry     time.Time
	lastFailure time.Time
}

func NewLimiter(cfg Config, m metrics.Metrics) (Limiter, error) {
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = defaultCooldown
	}
	if cfg.Window < cfg.Cooldown {
		cfg.Window = cfg.Cooldown
	}
	if cfg.NewBackoff == nil {
		// Defensive: a nil factory would panic on the first failure. Limiter and
		// Config are public types, so don't assume the caller set it.
		cfg.NewBackoff = CertBackoff(time.Second, cfg.Cooldown)
	}

	l := &limiter{
		cfg:      cfg,
		backends: map[kri.Identifier]*backendState{},
	}

	// GaugeFuncs so the values are computed at scrape time and can't get stuck
	// reporting a stale count once a backoff/cooldown has elapsed with no
	// further events.
	backingOff := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "cert_generation_backoff",
		Help: "Number of proxies currently backing off certificate generation after a failure",
	}, l.backingOffProxies)
	if err := m.Register(backingOff); err != nil {
		return nil, err
	}
	circuitOpen := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "cert_generation_circuit_open",
		Help: "Number of backends whose certificate issuance circuit breaker is currently open",
	}, l.openCircuits)
	if err := m.Register(circuitOpen); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *limiter) Allow(backend kri.Identifier, proxy model.ResourceKey) (bool, time.Duration) {
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

func (l *limiter) Record(backend kri.Identifier, proxy model.ResourceKey, success bool) {
	l.Lock()
	defer l.Unlock()

	now := core.Now()
	b := l.backends[backend]
	if b == nil {
		b = &backendState{proxies: map[model.ResourceKey]*proxyState{}}
		l.backends[backend] = b
	}

	if success {
		delete(b.proxies, proxy)
		b.openUntil = time.Time{} // any success closes the circuit
		b.probing = false
		if len(b.proxies) == 0 {
			delete(l.backends, backend)
		}
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
}

// Forget does not affect a backend's open circuit (that reflects a backend-wide
// problem, not this proxy).
func (l *limiter) Forget(proxy model.ResourceKey) {
	l.Lock()
	defer l.Unlock()
	for backend, b := range l.backends {
		delete(b.proxies, proxy)
		if len(b.proxies) == 0 && b.openUntil.IsZero() {
			delete(l.backends, backend)
		}
	}
}

// backingOffProxies is the metric value for cert_generation_backoff, computed
// at scrape time so it reflects the current instant.
func (l *limiter) backingOffProxies() float64 {
	l.Lock()
	defer l.Unlock()
	now := core.Now()
	n := 0
	for _, b := range l.backends {
		for _, p := range b.proxies {
			if p.nextTry.After(now) {
				n++
			}
		}
	}
	return float64(n)
}

// openCircuits is the metric value for cert_generation_circuit_open, computed
// at scrape time.
func (l *limiter) openCircuits() float64 {
	l.Lock()
	defer l.Unlock()
	now := core.Now()
	n := 0
	for _, b := range l.backends {
		if b.openUntil.After(now) {
			n++
		}
	}
	return float64(n)
}
