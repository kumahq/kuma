package cmd

import (
	"github.com/Kong/kuma/pkg/config"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/config/core/resources/store"
	core_plugins "github.com/Kong/kuma/pkg/core/plugins"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
)

var migrateLog = controlPlaneLog.WithName("migrate")

func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Launch Control Plane",
		Long:  `Launch Control Plane.`,
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
		Short: "Launch Control Plane",
		Long:  `Launch Control Plane.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := kuma_cp.DefaultConfig()
			err := config.Load(args.configPath, &cfg)
			if err != nil {
				migrateLog.Error(err, "could not load the configuration")
				return err
			}

			ver, err := migrate(cfg)
			if err != nil {
				if err == core_plugins.AlreadyMigrated {
					cmd.Printf("DB already migrated to the newest version: %d", ver)
				} else {
					return err
				}
			} else {
				cmd.Printf("DB migrated to %d version", ver)
			}

			return nil
		},
	}
	cmd.PersistentFlags().StringVarP(&args.configPath, "config-file", "c", "", "configuration file")
	return cmd
}

func migrate(cfg kuma_cp.Config) (uint, error) {
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
		return 0, errors.Errorf("unknown store type %s", cfg.Store.Type)
	}
	plugin, err := core_plugins.Plugins().ResourceStore(pluginName)
	if err != nil {
		return 0, errors.Wrapf(err, "could not retrieve store %s plugin", pluginName)
	}
	return plugin.Migrate(nil, pluginConfig)
}
