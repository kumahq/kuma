package tokens

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

type defaultSigningKeyComponent struct {
	signingKeyManager SigningKeyManager
	ctx               context.Context
}

var _ component.Component = &defaultSigningKeyComponent{}

func NewDefaultSigningKeyComponent(ctx context.Context, signingKeyManager SigningKeyManager) component.Component {
	return &defaultSigningKeyComponent{
		signingKeyManager: signingKeyManager,
		ctx:               ctx,
	}
}

func (d *defaultSigningKeyComponent) Start(stop <-chan struct{}) error {
	ctx, cancelFn := context.WithCancel(user.Ctx(d.ctx, user.ControlPlane))
	defer cancelFn()
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		backoff := retry.WithMaxDuration(10*time.Minute, retry.NewConstant(5*time.Second)) // if after this time we cannot create a resource - something is wrong and we should return an error which will restart CP.
		err := retry.Do(ctx, backoff, func(ctx context.Context) error {
			return retry.RetryableError(CreateDefaultSigningKeyIfNotExist(ctx, d.signingKeyManager)) // retry all errors
		})
		if err != nil {
			// Retry this operation since on Kubernetes, secrets are validated.
			// This code can execute before the control plane is ready therefore hooks can fail.
			errChan <- errors.Wrap(err, "could not create the default signing key")
		}
	}()
	select {
	case <-stop:
		return nil
	case err := <-errChan:
		return err
	}
}

func CreateDefaultSigningKeyIfNotExist(ctx context.Context, signingKeyManager SigningKeyManager) error {
	logger, err := kuma_log.FromContextWithNameAndOptionalValues(ctx, "defaults")
	if err != nil {
		return err
	}

	_, _, err = signingKeyManager.GetLatestSigningKey(ctx)
	if err == nil {
		logger.V(1).Info("signing key already exists. Skip creating.")
		return nil
	}
	if _, ok := err.(*SigningKeyNotFound); !ok {
		return err
	}
	logger.Info("trying to create signing key")
	if err := signingKeyManager.CreateDefaultSigningKey(ctx); err != nil {
		logger.V(1).Info("could not create signing key", "err", err)
		return err
	}
	logger.Info("default signing key created")
	return nil
}

func (d *defaultSigningKeyComponent) NeedLeaderElection() bool {
	return true
}
