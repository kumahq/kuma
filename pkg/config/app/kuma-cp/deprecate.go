package kuma_cp

import (
	"io"
	"time"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
)

var deprecations = []config.Deprecation{
	{
		Env:    "KUMA_METRICS_MESH_MIN_RESYNC_TIMEOUT",
		EnvMsg: "Use KUMA_METRICS_MESH_MIN_RESYNC_INTERVAL instead.",
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaCPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaCPConfig.Metrics == nil || kumaCPConfig.Metrics.Mesh == nil {
				return "", false
			}
			if kumaCPConfig.Metrics.Mesh.MinResyncTimeout.Duration == time.Duration(0) {
				return "", false
			}
			return "Metrics.Mesh.MinResyncTimeout", true
		},
		ConfigValueMsg: "Use Metrics.Mesh.MinResyncInterval instead.",
	},
	{
		Env:    "KUMA_METRICS_MESH_FULL_RESYNC_TIMEOUT",
		EnvMsg: "Use KUMA_METRICS_MESH_FULL_RESYNC_INTERVAL instead.",
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaCPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaCPConfig.Metrics == nil || kumaCPConfig.Metrics.Mesh == nil {
				return "", false
			}
			if kumaCPConfig.Metrics.Mesh.MaxResyncTimeout.Duration == time.Duration(0) {
				return "", false
			}
			return "Metrics.Mesh.MaxResyncTimeout", true
		},
		ConfigValueMsg: "Use Metrics.Mesh.FullResyncInterval instead.",
	},
	{
		Env:    "KUMA_STORE_POSTGRES_MIN_RECONNECT_INTERVAL",
		EnvMsg: "The env is specific to 'pq' driver which is marked as deprecated.",
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaCPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaCPConfig.Store == nil || kumaCPConfig.Store.Postgres == nil {
				return "", false
			}
			if kumaCPConfig.Store.Postgres.MinReconnectInterval.Duration == postgres.DefaultMinReconnectInterval.Duration {
				return "", false
			}
			return "Store.Postgres.MinReconnectInterval", true
		},
		ConfigValueMsg: "The config is specific to 'pq' driver which is marked as deprecated.",
	},
	{
		Env:    "KUMA_STORE_POSTGRES_MAX_IDLE_CONNECTIONS",
		EnvMsg: "The env is specific to 'pq' driver which is marked as deprecated.",
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaCPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaCPConfig.Store == nil || kumaCPConfig.Store.Postgres == nil {
				return "", false
			}
			if kumaCPConfig.Store.Postgres.MaxIdleConnections == postgres.DefaultMaxIdleConnections {
				return "", false
			}
			return "Store.Postgres.MaxIdleConnections", true
		},
		ConfigValueMsg: "The config is specific to 'pq' driver which is marked as deprecated.",
	},
	{
		Env:    "KUMA_STORE_POSTGRES_MAX_RECONNECT_INTERVAL",
		EnvMsg: "The env is specific to 'pq' driver which is marked as deprecated.",
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaCPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaCPConfig.Store == nil || kumaCPConfig.Store.Postgres == nil {
				return "", false
			}
			if kumaCPConfig.Store.Postgres.MaxReconnectInterval.Duration == postgres.DefaultMaxReconnectInterval.Duration {
				return "", false
			}
			return "Store.Postgres.MaxReconnectInterval", true
		},
		ConfigValueMsg: "The config is specific to 'pq' driver which is marked as deprecated.",
	},
}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}
