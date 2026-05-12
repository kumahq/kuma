package kuma_cp

import (
	"io"
	"time"

	"github.com/kumahq/kuma/v2/pkg/config"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
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
		Env:    "",
		EnvMsg: "",
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaCPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaCPConfig.MonitoringAssignmentServer == nil {
				return "", false
			}
			if kumaCPConfig.MonitoringAssignmentServer.Enabled &&
				kumaCPConfig.Environment == config_core.KubernetesEnvironment {
				return "MonitoringAssignmentServer", true
			}
			return "", false
		},
		ConfigValueMsg: "MADS is enabled on Kubernetes. It will be removed on Kubernetes in Kuma 3.0. " +
			"Set KUMA_MONITORING_ASSIGNMENT_SERVER_ENABLED=false to disable it.",
	},
}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}
