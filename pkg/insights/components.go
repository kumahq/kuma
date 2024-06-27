package insights

import (
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/status"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.GetMode() == config_core.Zone {
		logger := core.Log.WithName("meshservice").WithName("status-updater")
		updater, err := status.NewStatusUpdater(
			logger,
			rt.ReadOnlyResourceManager(),
			rt.ResourceManager(),
			rt.Config().CoreResources.Status.MeshServiceInterval.Duration,
			rt.Metrics(),
			rt.Config().Multizone.Zone.Name,
		)
		if err != nil {
			return err
		}
		if err := rt.Add(component.NewResilientComponent(
			logger,
			updater,
			rt.Config().General.ResilientComponentBaseBackoff.Duration,
			rt.Config().General.ResilientComponentMaxBackoff.Duration),
		); err != nil {
			return err
		}

		logger = core.Log.WithName("meshmultizoneservice").WithName("status-updater")
		meshmzsvcUpdater, err := meshmultizoneservice.NewStatusUpdater(
			logger,
			rt.ReadOnlyResourceManager(),
			rt.ResourceManager(),
			rt.Config().CoreResources.Status.MeshMultiZoneServiceInterval.Duration,
			rt.Metrics(),
		)
		if err != nil {
			return err
		}
		if err := rt.Add(component.NewResilientComponent(
			logger,
			meshmzsvcUpdater,
			rt.Config().General.ResilientComponentBaseBackoff.Duration,
			rt.Config().General.ResilientComponentMaxBackoff.Duration),
		); err != nil {
			return err
		}
	}
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
	return rt.Add(component.NewResilientComponent(log, resyncer, rt.Config().General.ResilientComponentBaseBackoff.Duration, rt.Config().General.ResilientComponentMaxBackoff.Duration))
}
