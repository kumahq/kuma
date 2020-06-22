package component

import (
	"sync"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/util/channels"
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

// Component of Kuma, i.e. gRPC Server, HTTP server, reconciliation loop.
var _ Component = ComponentFunc(nil)

type ComponentFunc func(<-chan struct{}) error

func (f ComponentFunc) NeedLeaderElection() bool {
	return false
}

func (f ComponentFunc) Start(stop <-chan struct{}) error {
	return f(stop)
}

type Manager interface {

	// Add registers a component, i.e. gRPC Server, HTTP server, reconciliation loop.
	Add(...Component) error

	// Start starts registered components and blocks until the Stop channel is closed.
	// Returns an error if there is an error starting any component.
	Start(<-chan struct{}) error
}

var _ Manager = &manager{}

func NewManager(leaderElector LeaderElector) Manager {
	return &manager{
		leaderElector: leaderElector,
	}
}

type manager struct {
	components    []Component
	leaderElector LeaderElector
}

func (cm *manager) Add(c ...Component) error {
	cm.components = append(cm.components, c...)
	return nil
}

func (cm *manager) Start(stop <-chan struct{}) error {
	errCh := make(chan error)

	cm.startNonLeaderComponents(stop, errCh)
	cm.startLeaderComponents(stop, errCh)

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
		if channels.IsClosed(leaderStopCh) {
			close(leaderStopCh)
		}
	}

	cm.leaderElector.AddCallbacks(LeaderCallbacks{
		OnStartedLeading: func() {
			log.Info("Leader acquired")
			mutex.Lock()
			defer mutex.Unlock()
			leaderStopCh = make(chan struct{})
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
			log.Info("Leader lost")
			closeLeaderCh()
		},
	})
	log.Info("Starting leader election")
	go cm.leaderElector.Start(stop)
	go func() {
		<-stop
		closeLeaderCh()
	}()
}
