package test

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/consts"
)

func InitializeConfigIPvX(
	cfg config.Config,
	ipv6 bool,
) config.InitializedConfigIPvX {
	inbound, err := cfg.Redirect.Inbound.Initialize(
		ipv6,
		consts.IptablesChainsPrefix,
	)
	if err != nil {
		panic(err)
	}

	outbound, err := cfg.Redirect.Outbound.Initialize(
		ipv6,
		consts.IptablesChainsPrefix,
	)
	if err != nil {
		panic(err)
	}

	vnet, err := cfg.Redirect.VNet.Initialize(ipv6)
	if err != nil {
		panic(err)
	}

	return config.InitializedConfigIPvX{
		Config: cfg,
		Redirect: config.InitializedRedirect{
			Redirect: cfg.Redirect,
			DNS:      config.InitializedDNS{DNS: cfg.Redirect.DNS},
			Inbound:  inbound,
			Outbound: outbound,
			VNet:     vnet,
		},
		LoopbackInterfaceName: "",
	}
}
