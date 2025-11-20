package generate

import (
	"slices"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/core"
	"github.com/kumahq/kuma/v2/pkg/core/runtime"
	"github.com/kumahq/kuma/v2/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.GetMode() == config_core.Global {
		return nil
	}
	if storeType := rt.Config().Store.Type; storeType == store.KubernetesStore {
		return nil
	}
	logger := core.Log.WithName("workload").WithName("generator")
	if !slices.Contains(rt.Config().CoreResources.Enabled, "workloads") {
		logger.Info("Workload is not enabled. Skip starting generator for Workload.")
		return nil
	}
	generator, err := New(
		logger,
		rt.Config().Workload.GenerationInterval.Duration,
		rt.Config().Workload.DeletionGracePeriod.Duration,
		rt.Metrics(),
		rt.ResourceManager(),
		rt.MeshCache(),
		rt.Config().Multizone.Zone.Name,
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
