package tls

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
)

type DefaultsComponent struct {
	ResManager manager.ResourceManager
	Log        logr.Logger
}

var _ component.Component = &DefaultsComponent{}

func (e *DefaultsComponent) Start(stop <-chan struct{}) error {
	ctx, cancelFn := context.WithCancel(user.Ctx(context.Background(), user.ControlPlane))
	go func() {
		<-stop
		cancelFn()
	}()
	return retry.Do(ctx, retry.WithMaxDuration(10*time.Minute, retry.NewConstant(5*time.Second)), func(ctx context.Context) error {
		if err := e.ensureInterCpCaExist(ctx); err != nil {
			e.Log.V(1).Info("could not ensure that Inter CP CA exists. Retrying.", "err", err)
			return retry.RetryableError(err)
		}
		return nil
	})
}

func (e DefaultsComponent) NeedLeaderElection() bool {
	return true
}

func (e *DefaultsComponent) ensureInterCpCaExist(ctx context.Context) error {
	_, err := LoadCA(ctx, e.ResManager)
	if err == nil {
		e.Log.V(1).Info("Inter CP CA already exists. Skip creating Envoy Admin CA.")
		return nil
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "error while loading admin client certificate")
	}
	e.Log.V(1).Info("trying to create Inter CP CA")
	pair, err := GenerateCA()
	if err != nil {
		return errors.Wrap(err, "could not generate admin client certificate")
	}
	if err := CreateCA(ctx, *pair, e.ResManager); err != nil {
		return errors.Wrap(err, "could not create admin client certificate")
	}
	e.Log.Info("Inter CP CA created")
	return nil
}
