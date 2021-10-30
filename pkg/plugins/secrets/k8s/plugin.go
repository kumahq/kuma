package k8s

import (
	"github.com/pkg/errors"

	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	k8s_extensions "github.com/kumahq/kuma/pkg/plugins/extensions/k8s"
)

var _ core_plugins.SecretStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Kubernetes, &plugin{})
}

func (p *plugin) NewSecretStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (secret_store.SecretStore, error) {
	client, ok := k8s_extensions.FromSecretClientContext(pc.Extensions())
	if !ok {
		return nil, errors.Errorf("secret client hasn't been configured")
	}
	return NewStore(client, client, pc.Config().Store.Kubernetes.SystemNamespace)
}
