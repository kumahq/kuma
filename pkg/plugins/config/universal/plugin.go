package universal

import (
	config_store "github.com/Kong/kuma/pkg/core/config/store"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
)

var _ core_plugins.ConfigStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Universal, &plugin{})
}

func (p *plugin) NewConfigStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (config_store.ConfigStore, error) {
	return config_store.NewConfigStore(pc.ResourceStore()), nil
}
