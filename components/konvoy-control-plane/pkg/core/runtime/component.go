package runtime

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Component of the Control Plane, i.e. gRPC Server, HTTP server, reconciliation loop.
type Component = manager.Runnable

var _ Component = ComponentFunc(nil)

type ComponentFunc func(<-chan struct{}) error

func (f ComponentFunc) Start(stop <-chan struct{}) error {
	return f(stop)
}

type ComponentManager interface {

	// Add registers a component, i.e. gRPC Server, HTTP server, reconciliation loop.
	Add(c Component) error

	// Start starts registered components and blocks until the Stop channel is closed.
	// Returns an error if there is an error starting any component.
	Start(<-chan struct{}) error
}

func Add(cm ComponentManager, cs ...Component) error {
	for _, c := range cs {
		if err := cm.Add(c); err != nil {
			return err
		}
	}
	return nil
}
