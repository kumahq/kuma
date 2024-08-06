package tokens

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
)

func EnsureDefaultSigningKeyExist(signingKeyPrefix string, ctx context.Context, resManager core_manager.ResourceManager, logger logr.Logger) error {
	logger = logger.WithValues("prefix", signingKeyPrefix)
	signingKeyManager := NewSigningKeyManager(resManager, signingKeyPrefix)
	_, _, err := signingKeyManager.GetLatestSigningKey(ctx)
	if err == nil {
		logger.V(1).Info("signing key already exists. Skip creating.")
		return nil
	}
	if _, ok := err.(*SigningKeyNotFound); !ok {
		return err
	}
	if err := signingKeyManager.CreateDefaultSigningKey(ctx); err != nil {
		return errors.Wrapf(err, "could not create signing key with prefix %s", signingKeyPrefix)
	}
	logger.Info("default signing key created")
	return nil
}
