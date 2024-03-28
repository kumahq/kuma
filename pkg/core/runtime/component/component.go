package component

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/util/channels"
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

func NewManager(leaderElector LeaderElector) Manager {
	return &manager{
		leaderElector: leaderElector,
	}
}

var LeaderComponentAddAfterStartErr = errors.New("cannot add leader component after component manager is started")

type manager struct {
	leaderElector LeaderElector

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
			go func(c Component, stopCh <-chan struct{}, errCh chan error) {
				if err := c.Start(stopCh); err != nil {
					errCh <- err
				}
			}(component, cm.stopCh, cm.errCh)
		}
	}
	return nil
}

func (cm *manager) Start(stop <-chan struct{}) error {
	errCh := make(chan error)

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
			go func(c Component) {
				if err := c.Start(stop); err != nil {
					errCh <- err
				}
			}(component)
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
			log.Info("leader acquired")
			mutex.Lock()
			defer mutex.Unlock()
			leaderStopCh = make(chan struct{})

			cm.Lock()
			defer cm.Unlock()
			for _, component := range cm.components {
				if component.NeedLeaderElection() {
					go func(c Component) {
						if err := c.Start(leaderStopCh); err != nil {
							errCh <- err
						}
					}(component)
				}
			}
		},
		OnStoppedLeading: func() {
			log.Info("leader lost")
			closeLeaderCh()
		},
	})
	go cm.leaderElector.Start(stop)
	go func() {
		<-stop
		closeLeaderCh()
	}()
}
