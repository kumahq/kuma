package insights

import (
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.Config().IsFederatedZoneCP() {
		return nil
	}
	minResyncInterval := rt.Config().Metrics.Mesh.MinResyncInterval.Duration
	if rt.Config().Metrics.Mesh.MinResyncTimeout.Duration != 0 {
		minResyncInterval = rt.Config().Metrics.Mesh.MinResyncTimeout.Duration
	}
	fullResyncInterval := rt.Config().Metrics.Mesh.FullResyncInterval.Duration
	if rt.Config().Metrics.Mesh.MaxResyncTimeout.Duration != 0 {
		fullResyncInterval = rt.Config().Metrics.Mesh.MaxResyncTimeout.Duration
	}
	resyncer := NewResyncer(&Config{
		ResourceManager:     rt.ResourceManager(),
		EventReaderFactory:  rt.EventBus(),
		MinResyncInterval:   minResyncInterval,
		FullResyncInterval:  fullResyncInterval,
		Registry:            registry.Global(),
		TenantFn:            rt.Tenants(),
		EventBufferCapacity: rt.Config().Metrics.Mesh.BufferSize,
		EventProcessors:     rt.Config().Metrics.Mesh.EventProcessors,
		Metrics:             rt.Metrics(),
		Extensions:          rt.Extensions(),
	})
	return rt.Add(component.NewResilientComponent(log, resyncer))
}
