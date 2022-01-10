package tokens

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

type defaultSigningKeyComponent struct {
	signingKeyManager SigningKeyManager
	log               logr.Logger
}

var _ component.Component = &defaultSigningKeyComponent{}

func NewDefaultSigningKeyComponent(signingKeyManager SigningKeyManager, log logr.Logger) component.Component {
	return &defaultSigningKeyComponent{
		signingKeyManager: signingKeyManager,
		log:               log,
	}
}

func (d *defaultSigningKeyComponent) Start(stop <-chan struct{}) error {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		backoff := retry.WithMaxDuration(10*time.Minute, retry.NewConstant(5*time.Second)) // if after this time we cannot create a resource - something is wrong and we should return an error which will restart CP.
		err := retry.Do(ctx, backoff, func(ctx context.Context) error {
			return retry.RetryableError(d.createDefaultSigningKeyIfNotExist(ctx)) // retry all errors
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

func (d *defaultSigningKeyComponent) createDefaultSigningKeyIfNotExist(ctx context.Context) error {
	_, _, err := d.signingKeyManager.GetLatestSigningKey(ctx)
	if err == nil {
		d.log.V(1).Info("signing key already exists. Skip creating.")
		return nil
	}
	if _, ok := err.(*SigningKeyNotFound); !ok {
		return err
	}
	d.log.Info("trying to create signing key")
	if err := d.signingKeyManager.CreateDefaultSigningKey(ctx); err != nil {
		d.log.V(1).Info("could not create signing key", "err", err)
		return err
	}
	d.log.Info("default signing key created")
	return nil
}

func (d *defaultSigningKeyComponent) NeedLeaderElection() bool {
	return true
}
