package kumadp

import (
	"io"

	"github.com/kumahq/kuma/pkg/config"
)

const DeprecateAdminPortMsg = "Please set adminPort directly in Dataplane resource, in the field 'networking.admin.port'."

var deprecations = []config.Deprecation{
	{
		Env:    "KUMA_DATAPLANE_ADMIN_PORT",
		EnvMsg: DeprecateAdminPortMsg,
		ConfigValuePath: func(cfg config.Config) (string, bool) {
			kumaDPConfig, ok := cfg.(*Config)
			if !ok {
				panic("wrong config type")
			}
			if kumaDPConfig.Dataplane.AdminPort.Empty() {
				return "", false
			}
			return "Dataplane.AdminPort", true
		},
		ConfigValueMsg: DeprecateAdminPortMsg,
	},
}

func PrintDeprecations(cfg *Config, out io.Writer) {
	config.PrintDeprecations(deprecations, cfg, out)
}
