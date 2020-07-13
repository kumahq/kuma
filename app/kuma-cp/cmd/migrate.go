package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	core_plugins "github.com/kumahq/kuma/pkg/core/plugins"
	"github.com/kumahq/kuma/pkg/version"
)

var migrateLog = controlPlaneLog.WithName("migrate")

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate database to which Control Plane is connected",
		Long:  `Migrate database to which Control Plane is connected. The database contains all policies, dataplanes and secrets. The schema has to be in sync with version of Kuma CP to properly work. Make sure to run "kuma-cp migrate up" before running new version of Kuma.`,
	}
	cmd.AddCommand(newMigrateUpCmd())
	return cmd
}

func newMigrateUpCmd() *cobra.Command {
	args := struct {
		configPath string
	}{}
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply the newest schema changes to the database.",
		Long:  `Apply the newest schema changes to the database.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := kuma_cp.DefaultConfig()
			err := config.Load(args.configPath, &cfg)
			if err != nil {
				migrateLog.Error(err, "could not load the configuration")
				return err
			}

			if err := migrate(cfg); err != nil {
				if err == core_plugins.AlreadyMigrated {
					cmd.Printf("DB has already been migrated for Kuma %s\n", version.Build.Version)
				} else {
					return err
				}
			} else {
				cmd.Printf("DB has been migrated for Kuma %s\n", version.Build.Version)
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&args.configPath, "config-file", "c", "", "configuration file")
	return cmd
}

func migrate(cfg kuma_cp.Config) error {
	var pluginName core_plugins.PluginName
	var pluginConfig core_plugins.PluginConfig
	switch cfg.Store.Type {
	case store.KubernetesStore:
		pluginName = core_plugins.Kubernetes
		pluginConfig = nil
	case store.MemoryStore:
		pluginName = core_plugins.Memory
		pluginConfig = nil
	case store.PostgresStore:
		pluginName = core_plugins.Postgres
		pluginConfig = cfg.Store.Postgres
	default:
		return errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().ResourceStore(pluginName)
	if err != nil {
		return errors.Wrapf(err, "could not retrieve store %s plugin", pluginName)
	}
	_, err = plugin.Migrate(nil, pluginConfig)
	return err
}
