package transparentproxy

import (
	"context"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/ebpf"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables"
)

func Setup(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	if cfg.IPv4.Ebpf.Enabled {
		return ebpf.Setup(cfg.IPv4)
	}

	return iptables.Setup(ctx, cfg)
}

func Cleanup(ctx context.Context, cfg config.InitializedConfig) error {
	if cfg.IPv4.Ebpf.Enabled {
		_, err := ebpf.Cleanup(cfg.IPv4)
		return err
	}

	return iptables.Cleanup(ctx, cfg)
}
