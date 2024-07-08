package test

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

func InitializeConfig(cfg config.Config) config.InitializedConfig {
	inbound, err := cfg.Redirect.Inbound.Initialize(consts.IptablesChainsPrefix)
	if err != nil {
		panic(err)
	}

	outbound, err := cfg.Redirect.Outbound.Initialize(consts.IptablesChainsPrefix)
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
