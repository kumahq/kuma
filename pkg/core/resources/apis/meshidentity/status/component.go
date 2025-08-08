package status

import (
	"slices"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

func Setup(rt runtime.Runtime) error {
	// currently we support only k8s
	if rt.GetMode() == config_core.Global || rt.Config().Environment != config_core.KubernetesEnvironment {
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
