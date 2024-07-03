package test

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

func InitializeConfig(cfg config.Config) config.InitializedConfig {
	return config.InitializedConfig{
		Config: cfg,
		Redirect: config.InitializedRedirect{
			Redirect: cfg.Redirect,
			DNS: config.InitializedDNS{
				DNS:         cfg.Redirect.DNS,
				ServersIPv4: nil,
				ServersIPv6: nil,
			},
		},
		LoopbackInterfaceName: "",
	}
}
