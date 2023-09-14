package universal

import (
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	secret_store "github.com/kumahq/kuma/pkg/core/secrets/store"
)

var _ core_plugins.SecretStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) NewSecretStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (secret_store.SecretStore, error) {
	return secret_store.NewSecretStore(pc.ResourceStore().DefaultResourceStore()), nil
}
