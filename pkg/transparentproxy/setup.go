package transparentproxy

import (
	"context"

	"github.com/kumahq/kuma/v2/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/v2/pkg/transparentproxy/ebpf"
	"github.com/kumahq/kuma/v2/pkg/transparentproxy/iptables"
)

func Setup(ctx context.Context, cfg config.Config) (string, error) {
	if cfg.Ebpf.Enabled {
		return ebpf.Setup(cfg)
	}

	return iptables.Setup(ctx, cfg)
}

func Cleanup(cfg config.Config) (string, error) {
	if cfg.Ebpf.Enabled {
		return ebpf.Cleanup(cfg)
	}
	return iptables.Cleanup(cfg)
}
