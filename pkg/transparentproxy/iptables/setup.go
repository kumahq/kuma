package iptables

import (
	"context"
	"errors"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/builder"
)

func Setup(ctx context.Context, cfg config.Config) (string, error) {
	if cfg.DryRun {
		// TODO (bartsmykla): we should generate IPv4 and IPv6 when cfg.IPv6 is
		//  set, but currently in DryRun mode we would just display IPv6
		//  configuration when cfg.IPv6 is set
		// TODO (bartsmykla): I think dns servers should be provided as a config
		//  value instead of explicit function parameter here
		output, err := builder.BuildIPTables(cfg, nil, cfg.IPv6)
		if err != nil {
			return "", err
		}

		_, _ = cfg.RuntimeStdout.Write([]byte(output))

		return output, nil
	}

	return builder.RestoreIPTables(ctx, cfg)
}

func Cleanup(cfg config.Config) (string, error) {
	return "", errors.New("cleanup is not supported")
}
