package component

type Component interface {
	// Start blocks until the channel is closed or an error occurs.
	// The component will stop running when the channel is closed.
	Start(<-chan struct{}) error
}

// Component of Kuma, i.e. gRPC Server, HTTP server, reconciliation loop.
var _ Component = ComponentFunc(nil)

type ComponentFunc func(<-chan struct{}) error

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

func NewManager() Manager {
	return &manager{}
}

type manager struct {
	components []Component
}

func (cm *manager) Add(c ...Component) error {
	cm.components = append(cm.components, c...)
	return nil
}

func (cm *manager) Start(stop <-chan struct{}) error {
	errCh := make(chan error)
	for _, component := range cm.components {
		go func(c Component) {
			if err := c.Start(stop); err != nil {
				errCh <- err
			}
		}(component)
	}
	select {
	case <-stop:
		return nil
	case err := <-errCh:
		return err
	}
}
