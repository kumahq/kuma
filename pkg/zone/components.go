package zone

import (
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.GetMode() == config_core.Global {
		return nil
	}
	logger := core.Log.WithName("zone").WithName("components")
	tracker, err := NewZoneAvailableServicesTracker(
		logger,
		rt.Metrics(),
		rt.ResourceManager(),
		rt.MeshCache(),
		rt.Config().Multizone.Zone.IngressUpdateInterval.Duration,
		rt.Config().Experimental.IngressTagFilters,
		rt.Config().Multizone.Zone.Name,
	)
	if err != nil {
		return err
	}
	return rt.Add(component.NewResilientComponent(
		logger,
		tracker,
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration,
	))
}
