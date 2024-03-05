package transparentproxy

import (
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/ebpf"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables"
)

func Setup(cfg config.Config) (string, error) {
	if cfg.Ebpf.Enabled {
		return ebpf.Setup(cfg)
	}

	return iptables.Setup(cfg)
}

func Cleanup(cfg config.Config) (string, error) {
	if cfg.Ebpf.Enabled {
		return ebpf.Cleanup(cfg)
	}
	return iptables.Cleanup(cfg)
}
