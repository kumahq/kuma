package defaults

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/envoy/admin/tls"
)

type EnvoyAdminCaDefaultComponent struct {
	ResManager manager.ResourceManager
}

var _ component.Component = &EnvoyAdminCaDefaultComponent{}

func (e *EnvoyAdminCaDefaultComponent) Start(stop <-chan struct{}) error {
	ctx, cancelFn := context.WithCancel(user.Ctx(context.Background(), user.ControlPlane))
	go func() {
		<-stop
		cancelFn()
	}()
	return retry.Do(ctx, retry.WithMaxDuration(10*time.Minute, retry.NewConstant(5*time.Second)), func(ctx context.Context) error {
		if err := e.ensureEnvoyAdminCaExist(ctx); err != nil {
			log.V(1).Info("could not ensure that Envoy Admin CA exists. Retrying.", "err", err)
			return retry.RetryableError(err)
		}
		return nil
	})
}

func (e EnvoyAdminCaDefaultComponent) NeedLeaderElection() bool {
	return true
}

func (e *EnvoyAdminCaDefaultComponent) ensureEnvoyAdminCaExist(ctx context.Context) error {
	_, err := tls.LoadCA(ctx, e.ResManager)
	if err == nil {
		log.V(1).Info("Envoy Admin CA already exists. Skip creating Envoy Admin CA.")
		return nil
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "error while loading admin client certificate")
	}
	log.V(1).Info("trying to create Envoy Admin CA")
	pair, err := tls.GenerateCA()
	if err != nil {
		return errors.Wrap(err, "could not generate admin client certificate")
	}
	if err := tls.CreateCA(ctx, *pair, e.ResManager); err != nil {
		return errors.Wrap(err, "could not create admin client certificate")
	}
	log.Info("Envoy Admin CA created")
	return nil
}
