package transparentproxy

import (
	"context"

	"github.com/kumahq/kuma/v3/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/v3/pkg/transparentproxy/iptables"
)

func Setup(ctx context.Context, cfg config.InitializedConfig) (string, error) {
	return iptables.Setup(ctx, cfg)
}

func Cleanup(ctx context.Context, cfg config.InitializedConfig) error {
	return iptables.Cleanup(ctx, cfg)
}
