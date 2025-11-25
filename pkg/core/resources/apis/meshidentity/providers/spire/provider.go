package spire

import (
	"fmt"

	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	core_plugins "github.com/kumahq/kuma/v2/pkg/core/plugins"
	meshidentity_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshidentity/providers"
)

var _ core_plugins.IdentityProviderPlugin = &plugin{}

type plugin struct{}

func InitProvider() {
	core_plugins.Register(core_plugins.PluginName(meshidentity_api.SpireType), &plugin{})
}

func (p plugin) NewIdentityProvider(context core_plugins.PluginContext, config core_plugins.PluginConfig) (providers.IdentityProvider, error) {
	var socketPath string
	switch context.Config().Environment {
	case config_core.UniversalEnvironment:
		socketPath = context.Config().Runtime.Universal.Spire.SocketPath
	case config_core.KubernetesEnvironment:
		socketPath = fmt.Sprintf(
			"%s/%s",
			context.Config().Runtime.Kubernetes.Injector.Spire.MountPath,
			context.Config().Runtime.Kubernetes.Injector.Spire.SocketFileName,
		)
	default:
		return nil, fmt.Errorf("cannot initialize spire identity provider, unsupported environment: %s")
	}
	return NewSpireIdentityProvider(
		socketPath,
		context.Config().Multizone.Zone.Name,
		context.Config().Environment,
	), nil
}
