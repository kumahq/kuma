package spire

import (
	"github.com/kumahq/kuma/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	meshidentity_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshidentity/providers"
)

var _ core_plugins.IdentityProviderPlugin = &plugin{}

type plugin struct{}

func InitProvider() {
	core_plugins.Register(core_plugins.PluginName(meshidentity_api.SpireType), &plugin{})
}

func (p plugin) NewIdentityProvider(context core_plugins.PluginContext, config core_plugins.PluginConfig) (providers.IdentityProvider, error) {
	if context.Config().Environment == core.KubernetesEnvironment {
		return NewSpireIdentityProvider(
			context.Config().Runtime.Kubernetes.Injector.Spire.MountPath,
			context.Config().Runtime.Kubernetes.Injector.Spire.SocketFileName,
			context.Config().Multizone.Zone.Name,
		), nil
	}
	return nil, nil
}
