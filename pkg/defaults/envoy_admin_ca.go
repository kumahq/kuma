package defaults

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/envoy/admin/tls"
)

func EnsureEnvoyAdminCaExists(
	ctx context.Context,
	resManager manager.ResourceManager,
	logger logr.Logger,
	cfg kuma_cp.Config,
) error {
	if cfg.Mode == config_core.Global {
		return nil // Envoy Admin CA is not synced in multizone env and not needed in Global CP.
	}
	_, err := tls.LoadCA(ctx, resManager)
	if err == nil {
		logger.V(1).Info("Envoy Admin CA already exists. Skip creating Envoy Admin CA.")
		return nil
	}
	if !store.IsNotFound(err) {
		return errors.Wrap(err, "error while loading envoy admin CA")
	}
	pair, err := tls.GenerateCA()
	if err != nil {
		return errors.Wrap(err, "could not generate envoy admin CA")
	}
	if err := tls.CreateCA(ctx, *pair, resManager); err != nil {
		return errors.Wrap(err, "could not create envoy admin CA")
	}
	logger.Info("Envoy Admin CA created")
	return nil
}
