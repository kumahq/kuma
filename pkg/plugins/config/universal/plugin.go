package universal

import (
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
)

var _ core_plugins.ConfigStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) NewConfigStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	return pc.ResourceStore().DefaultResourceStore(), nil
}
