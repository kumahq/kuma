package postgres

import (
	"errors"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/plugins/resources/postgres"
	core_plugins "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/plugins"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
)

var _ core_plugins.ResourceStorePlugin = &plugin{}

type plugin struct{}

func init() {
	core_plugins.Register(core_plugins.Postgres, &plugin{})
}

func (p *plugin) NewResourceStore(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_store.ResourceStore, error) {
	cfg, ok := config.(*postgres.PostgresStoreConfig)
	if !ok {
		return nil, errors.New("invalid type of the config. Passed config should be a PostgresStoreConfig")
	}
	return NewStore(*cfg)
}
