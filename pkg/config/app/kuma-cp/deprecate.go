package kuma_cp

import (
	"io"
	"time"

	"github.com/kumahq/kuma/pkg/config"
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
			if kumaCPConfig.Metrics.Mesh.MinResyncTimeout.Duration != time.Duration(0) {
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
			if kumaCPConfig.Metrics.Mesh.MaxResyncTimeout.Duration != time.Duration(0) {
				return "", false
			}
			return "Metrics.Mesh.MaxResyncTimeout", true
		},
		ConfigValueMsg: "Use Metrics.Mesh.FullResyncInterval instead.",
	},
}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}
