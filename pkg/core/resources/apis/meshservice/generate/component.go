package generate

import (
	"slices"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.GetMode() == config_core.Global {
		return nil
	}
	if storeType := rt.Config().Store.Type; storeType == store.KubernetesStore {
		return nil
	}
	logger := core.Log.WithName("meshservice").WithName("generator")
	if !slices.Contains(rt.Config().CoreResources.Enabled, "meshservices") {
		logger.Info("MeshService is not enabled. Skip starting generator for MeshService.")
		return nil
	}
	generator, err := New(
		logger,
		rt.Config().MeshService.GenerationInterval.Duration,
		rt.Config().MeshService.DeletionGracePeriod.Duration,
		rt.Metrics(),
		rt.ResourceManager(),
		rt.MeshCache(),
	)
	if err != nil {
		return err
	}
	return rt.Add(component.NewResilientComponent(
		logger,
		generator,
		rt.Config().General.ResilientComponentBaseBackoff.Duration,
		rt.Config().General.ResilientComponentMaxBackoff.Duration,
	))
}
