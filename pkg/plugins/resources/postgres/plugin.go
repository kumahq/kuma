package postgres

import (
	"errors"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/events"
	postgres_events "github.com/kumahq/kuma/pkg/plugins/resources/postgres/events"
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
	return NewStore(pc.Metrics(), *cfg)
}

func (p *plugin) Migrate(pc core_plugins.PluginContext, config core_plugins.PluginConfig) (core_plugins.DbVersion, error) {
	cfg, ok := config.(*postgres.PostgresStoreConfig)
	if !ok {
		return 0, errors.New("invalid type of the config. Passed config should be a PostgresStoreConfig")
	}
	return migrateDb(*cfg)
}

func (p *plugin) EventListener(pc core_plugins.PluginContext, out events.Emitter) error {
	return pc.ComponentManager().Add(postgres_events.NewListener(*pc.Config().Store.Postgres, out, true))
}
