package issuer

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var log = core.Log.WithName("plugins").WithName("authn").WithName("api-server").WithName("tokens")

type defaultSigningKeyComponent struct {
	signingKeyManager SigningKeyManager
}

var _ component.Component = &defaultSigningKeyComponent{}

func NewDefaultSigningKeyComponent(signingKeyManager SigningKeyManager) component.Component {
	return &defaultSigningKeyComponent{
		signingKeyManager: signingKeyManager,
	}
}

func (d *defaultSigningKeyComponent) Start(stop <-chan struct{}) error {
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		if err := doWithRetry(ctx, d.createDefaultSigningKeyIfNotExist); err != nil {
			// Retry this operation since on Kubernetes, secrets are validated.
			// This code can execute before the control plane is ready therefore hooks can fail.
			errChan <- errors.Wrap(err, "could not create the default user token's signing key")
		}
	}()
	select {
	case <-stop:
		return nil
	case err := <-errChan:
		return err
	}
}

func (d *defaultSigningKeyComponent) createDefaultSigningKeyIfNotExist() error {
	_, _, err := d.signingKeyManager.GetLatestSigningKey()
	if err == nil {
		log.V(1).Info("user token's signing key already exists. Skip creating.")
		return nil
	}
	if err != SigningKeyNotFound {
		return err
	}
	log.Info("trying to create user token's signing key")
	if err := d.signingKeyManager.CreateDefaultSigningKey(); err != nil {
		log.V(1).Info("could not create user token's signing key", "err", err)
		return err
	}
	log.Info("default user token's signing key created")
	return nil
}

func (d *defaultSigningKeyComponent) NeedLeaderElection() bool {
	return true
}

func doWithRetry(ctx context.Context, fn func() error) error {
	backoff, _ := retry.NewConstant(5 * time.Second)
	backoff = retry.WithMaxDuration(10*time.Minute, backoff) // if after this time we cannot create a resource - something is wrong and we should return an error which will restart CP.
	return retry.Do(ctx, backoff, func(ctx context.Context) error {
		return retry.RetryableError(fn()) // retry all errors
	})
}
