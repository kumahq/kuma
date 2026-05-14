package component

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/util/channels"
)

var log = core.Log.WithName("bootstrap")

// Component defines a process that will be run in the application
// Component should be designed in such a way that it can be stopped by stop channel and started again (for example when instance is reelected for a leader).
type Component interface {
	// Start blocks until the channel is closed or an error occurs.
	// The component will stop running when the channel is closed.
	Start(<-chan struct{}) error

	// NeedLeaderElection indicates if component should be run only by one instance of Control Plane even with many Control Plane replicas.
	NeedLeaderElection() bool
}

// NamedComponent allows a Component to provide a stable name used in supervision logs.
// Name must return a non-empty string; an empty return falls back to the component's Go type.
type NamedComponent interface {
	Name() string
}

func componentName(c Component) string {
	if n, ok := c.(NamedComponent); ok {
		if name := n.Name(); name != "" {
			return name
		}
	}
	return fmt.Sprintf("%T", c)
}

type ReadyComponent interface {
	Component

	// Ready returns true if the component is ready to serve traffic.
	// This is useful to read this for a readiness probe
	Ready() bool
}

// GracefulComponent is a component that supports waiting until it's finished.
// It's useful if there is cleanup logic that has to be executed before the process exits
// (i.e. sending SIGTERM signals to subprocesses started by this component).
type GracefulComponent interface {
	Component

	// WaitForDone blocks until all components are done.
	// If a component was not started (i.e. leader components on non-leader CP) it returns immediately.
	WaitForDone()
}

// Component of Kuma, i.e. gRPC Server, HTTP server, reconciliation loop.
var _ Component = ComponentFunc(nil)

type ComponentFunc func(<-chan struct{}) error

func (f ComponentFunc) NeedLeaderElection() bool {
	return false
}

func (f ComponentFunc) Start(stop <-chan struct{}) error {
	return f(stop)
}

var _ Component = LeaderComponentFunc(nil)

type LeaderComponentFunc func(<-chan struct{}) error

func (f LeaderComponentFunc) NeedLeaderElection() bool {
	return true
}

func (f LeaderComponentFunc) Start(stop <-chan struct{}) error {
	return f(stop)
}

type Manager interface {
	// Add registers a component, i.e. gRPC Server, HTTP server, reconciliation loop.
	Add(...Component) error

	// Start starts registered components and blocks until the Stop channel is closed.
	// Returns an error if there is an error starting any component.
	// If there are any GracefulComponent, it waits until all components are done.
	Start(<-chan struct{}) error

	Ready() bool
}

var _ Manager = &manager{}

func NewManager(leaderElector LeaderElector, instanceId ...string) Manager {
	id := ""
	if len(instanceId) > 0 {
		id = instanceId[0]
	}
	return &manager{
		leaderElector: leaderElector,
		instanceId:    id,
		panicLimiter:  newPanicLogLimiter(10 * time.Second),
	}
}

var LeaderComponentAddAfterStartErr = errors.New("cannot add leader component after component manager is started")

type limiterEntry struct {
	last       time.Time
	suppressed int
}

type panicLogLimiter struct {
	window  time.Duration
	mu      sync.Mutex
	entries map[string]limiterEntry // at most len(registered components) entries; no cleanup needed
}

func newPanicLogLimiter(window time.Duration) *panicLogLimiter {
	return &panicLogLimiter{window: window, entries: map[string]limiterEntry{}}
}

// shouldLog returns (true, n) at most once per window per component name, where n is the number
// of panics suppressed since the last allowed log. Returns (false, 0) within the window.
func (p *panicLogLimiter) shouldLog(name string) (bool, int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	e := p.entries[name]
	if !e.last.IsZero() && now.Sub(e.last) < p.window {
		e.suppressed++
		p.entries[name] = e
		return false, 0
	}
	n := e.suppressed
	p.entries[name] = limiterEntry{last: now}
	return true, n
}

type manager struct {
	leaderElector LeaderElector
	instanceId    string
	panicLimiter  *panicLogLimiter

	sync.Mutex // protects access to fields below
	components []Component
	started    bool
	stopCh     <-chan struct{}
	errCh      chan error
}

func (cm *manager) Ready() bool {
	cm.Lock()
	defer cm.Unlock()

	for _, component := range cm.components {
		if ready, ok := component.(ReadyComponent); ok {
			if !ready.Ready() {
				return false
			}
		}
	}
	return true
}

// runComponent supervises a single component goroutine: emits start/stop bookend
// logs, recovers panics with per-name rate limiting, and forwards errors to errCh.
//
// Note: components wrapped in resilientComponent have their own recover() and restart
// loop; they never propagate panics to this layer. This path covers only unwrapped components.
func (cm *manager) runComponent(c Component, stop <-chan struct{}, errCh chan<- error) {
	name := componentName(c)
	cLog := log.WithValues("component", name)
	cLog.Info("component starting")
	panicked := false
	// LIFO: the recovery defer (registered second) fires before the bookend defer,
	// sets panicked=true so "component stopped" is skipped on a panic exit.
	defer func() {
		if !panicked {
			cLog.Info("component stopped")
		}
	}()
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			panicErr := fmt.Errorf("component %q panicked: %v", name, r)
			// Non-blocking: manager.Start consumes only the first error from errCh;
			// goroutines that send after Start returns would block forever otherwise.
			// shouldLog fires inside each branch so exactly one token is consumed per panic.
			select {
			case errCh <- panicErr:
				if ok, suppressed := cm.panicLimiter.shouldLog(name); ok {
					cLog.Error(panicErr, "component panicked", "stack", string(debug.Stack()), "suppressed", suppressed)
				}
			default:
				// errCh consumer already exited; log here so the panic is not lost.
				if ok, suppressed := cm.panicLimiter.shouldLog(name); ok {
					cLog.Error(panicErr, "component panicked, manager already exited", "stack", string(debug.Stack()), "suppressed", suppressed)
				}
			}
		}
	}()
	if err := c.Start(stop); err != nil {
		// Info not Error: error is forwarded to errCh and returned from manager.Start; logging Error here would double-log.
		cLog.Info("component exited with error", "error", err)
		select {
		case errCh <- err:
		default:
			cLog.Error(err, "component exited with error, manager already exited")
		}
	}
}

func (cm *manager) Add(c ...Component) error {
	cm.Lock()
	defer cm.Unlock()
	cm.components = append(cm.components, c...)
	if cm.started {
		// start component if it's added in runtime after Start is called.
		for _, component := range c {
			if component.NeedLeaderElection() {
				return LeaderComponentAddAfterStartErr
			}
			go cm.runComponent(component, cm.stopCh, cm.errCh)
		}
	}
	return nil
}

func (cm *manager) Start(stop <-chan struct{}) error {
	// Buffer size 1: ensures the first error is never dropped even if a component
	// panics before Start reaches its select (unbuffered would silently discard it).
	// Subsequent errors still hit the non-blocking default path.
	errCh := make(chan error, 1)

	cm.Lock()
	internalDone := make(chan struct{})
	cm.startNonLeaderComponents(internalDone, errCh)
	cm.started = true
	cm.stopCh = internalDone
	cm.errCh = errCh
	cm.Unlock()
	// this has to be called outside of lock because it can be leader at any time, and it locks in leader callbacks.
	cm.startLeaderComponents(internalDone, errCh)

	defer func() {
		close(internalDone)
		// limitation: waitForDone does not wait for components added after Start() is called.
		// This is ok for now, because it's used only in context of Kuma DP where we are not adding components in runtime.
		for _, c := range cm.components {
			if gc, ok := c.(GracefulComponent); ok {
				gc.WaitForDone()
			}
		}
	}()
	select {
	case <-stop:
		return nil
	case err := <-errCh:
		return err
	}
}

func (cm *manager) startNonLeaderComponents(stop <-chan struct{}, errCh chan error) {
	for _, component := range cm.components {
		if !component.NeedLeaderElection() {
			go cm.runComponent(component, stop, errCh)
		}
	}
}

func (cm *manager) startLeaderComponents(stop <-chan struct{}, errCh chan error) {
	// leader stop channel needs to be stored in atomic because it will be written by leader elector goroutine
	// and read by the last goroutine in this function.
	// we need separate channel for leader components because they can be restarted
	mutex := sync.Mutex{}
	leaderStopCh := make(chan struct{})
	closeLeaderCh := func() {
		mutex.Lock()
		defer mutex.Unlock()
		if !channels.IsClosed(leaderStopCh) {
			close(leaderStopCh)
		}
	}

	cm.leaderElector.AddCallbacks(LeaderCallbacks{
		OnStartedLeading: func() {
			log.Info("leader acquired", "instanceId", cm.instanceId)
			mutex.Lock()
			defer mutex.Unlock()
			leaderStopCh = make(chan struct{})

			cm.Lock()
			defer cm.Unlock()
			for _, component := range cm.components {
				if component.NeedLeaderElection() {
					go cm.runComponent(component, leaderStopCh, errCh)
				}
			}
		},
		OnStoppedLeading: func() {
			log.Info("leader lost", "instanceId", cm.instanceId)
			closeLeaderCh()
		},
	})
	go cm.leaderElector.Start(stop)
	go func() {
		<-stop
		closeLeaderCh()
	}()
}
