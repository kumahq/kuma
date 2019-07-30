package postgres

import (
	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
)

var _ core_plugins.ResourceStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Postgres, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, _ core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	return NewStore(*pc.Config().Store.Postgres)
}
