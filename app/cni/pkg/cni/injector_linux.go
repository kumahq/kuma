package cni

import (
	"context"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/v3/pkg/transparentproxy"
	"github.com/kumahq/kuma/v3/pkg/transparentproxy/config"
)

func injectIptables(ctx context.Context, netns string, cfg config.Config) error {
	namespace, err := ns.GetNS(netns)
	if err != nil {
		return errors.Wrapf(err, "failed to open network namespace '%s'", netns)
	}
	defer namespace.Close()

	return namespace.Do(func(ns.NetNS) error {
		initializedConfig, err := cfg.Initialize(ctx)
		if err != nil {
			return errors.Wrap(err, "failed to initialize transparent proxy configuration")
		}

		if _, err := transparentproxy.Setup(ctx, initializedConfig); err != nil {
			return errors.Wrap(err, "failed to set up transparent proxy")
		}

		return nil
	})
}
