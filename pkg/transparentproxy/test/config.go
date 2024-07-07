package test

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

func InitializeConfig(cfg config.Config) config.InitializedConfig {
	inbound, err := cfg.Redirect.Inbound.Initialize()
	if err != nil {
		panic(err)
	}

	outbound, err := cfg.Redirect.Outbound.Initialize()
	if err != nil {
		panic(err)
	}

	vnet, err := cfg.Redirect.VNet.Initialize()
	if err != nil {
		panic(err)
	}

	return config.InitializedConfig{
		Config: cfg,
		Redirect: config.InitializedRedirect{
			Redirect: cfg.Redirect,
			DNS: config.InitializedDNS{
				DNS:         cfg.Redirect.DNS,
				ServersIPv4: nil,
				ServersIPv6: nil,
			},
			Inbound:  inbound,
			Outbound: outbound,
			VNet:     vnet,
		},
		LoopbackInterfaceName: "",
	}
}
