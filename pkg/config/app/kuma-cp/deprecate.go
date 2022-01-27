package kuma_cp

import (
	"io"

	"github.com/kumahq/kuma/pkg/config"
)

var deprecations = []config.Deprecation{
	{
		Env:    "KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT",
		EnvMsg: "Use KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT instead.",
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaCPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaCPConfig.Runtime == nil || kumaCPConfig.Runtime.Kubernetes == nil {
				return "", false
			}
			if kumaCPConfig.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort == 0 {
				return "", false
			}
			return "Runtime.Kubernetes.Injector.SidecarContainer.AdminPort", true
		},
		ConfigValueMsg: "Use BootstrapServer.Params.AdminPort instead.",
	},
}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}

func (c Config) GetEnvoyAdminPort() uint32 {
	adminPort := func() uint32 {
		if c.BootstrapServer == nil || c.BootstrapServer.Params == nil {
			return 0
		}
		return c.BootstrapServer.Params.AdminPort
	}

	// start of the backwards compatibility code
	deprecatedAdminPort := func() uint32 {
		if c.Runtime == nil || c.Runtime.Kubernetes == nil {
			return 0
		}
		return c.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort
	}

	if deprecatedAdminPort() != 0 && adminPort() == 0 {
		return deprecatedAdminPort()
	}
	// end of the backwards compatibility code

	return adminPort()
}
