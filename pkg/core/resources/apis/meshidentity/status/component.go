package status

import (
	"slices"

	config_core "github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/core"
	resource_labels "github.com/kumahq/kuma/v3/pkg/core/resources/labels"
	"github.com/kumahq/kuma/v3/pkg/core/runtime"
	"github.com/kumahq/kuma/v3/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	if rt.GetMode() == config_core.Global {
		return nil
	}
	logger := core.Log.WithName("meshidentity").WithName("generator")
	if !slices.Contains(rt.Config().CoreResources.Enabled, "meshidentities") {
		logger.Info("MeshIdentity is not enabled. Skip starting generator for MeshIdentity.")
		return nil
	}
	generator, err := New(
		logger,
		rt.Config().CoreResources.Status.MeshIdentityInterval.Duration,
		rt.ResourceManager(),
		rt.ReadOnlyResourceManager(),
		rt.IdentityProviders(),
		resource_labels.ControlPlaneFromConfig(rt.Config()),
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
