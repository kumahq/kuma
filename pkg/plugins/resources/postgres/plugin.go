package postgres

import (
	"errors"

	"github.com/Kong/kuma/pkg/config/plugins/resources/postgres"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
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
	migrated, err := isDbMigrated(*cfg)
	if err != nil {
		return nil, err
	}
	if !migrated {
		return nil, errors.New(`database is not migrated. Run "kuma-cp migrate up" to update database to the newest schema`)
	}
	return NewStore(*cfg)
}

func (p *plugin) Migrate(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_plugins.DbVersion, error) {
	cfg, ok := config.(*postgres.PostgresStoreConfig)
	if !ok {
		return 0, errors.New("invalid type of the config. Passed config should be a PostgresStoreConfig")
	}
	return migrateDb(*cfg)
}
