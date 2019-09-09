package universal

import (
	"github.com/Kong/kuma/pkg/core/runtime"
)

var _ runtime.ComponentManager = &componentManager{}

func NewComponentManager() runtime.ComponentManager {
	return &componentManager{}
}

type componentManager struct {
	components []runtime.Component
}

func (cm *componentManager) Add(c runtime.Component) error {
	cm.components = append(cm.components, c)
	return nil
}

func (cm *componentManager) Start(stop <-chan struct{}) error {
	errCh := make(chan error)
	for _, component := range cm.components {
		go func(c runtime.Component) {
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
