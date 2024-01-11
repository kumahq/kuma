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
	kuma_log "github.com/kumahq/kuma/pkg/log"
)

type EnvoyAdminCaDefaultComponent struct {
	ResManager manager.ResourceManager
	Extensions context.Context
}

var _ component.Component = &EnvoyAdminCaDefaultComponent{}

func (e *EnvoyAdminCaDefaultComponent) Start(stop <-chan struct{}) error {
	ctx, cancelFn := context.WithCancel(user.Ctx(context.Background(), user.ControlPlane))
	defer cancelFn()
	logger := kuma_log.AddFieldsFromCtx(log, ctx, e.Extensions)
	errChan := make(chan error)
	go func() {
		errChan <- retry.Do(ctx, retry.WithMaxDuration(10*time.Minute, retry.NewConstant(5*time.Second)), func(ctx context.Context) error {
			if err := EnsureEnvoyAdminCaExist(ctx, e.ResManager, e.Extensions); err != nil {
				logger.V(1).Info("could not ensure that Envoy Admin CA exists. Retrying.", "err", err)
				return retry.RetryableError(err)
			}
			return nil
		})
	}()
	select {
	case <-stop:
		return nil
	case err := <-errChan:
		return err
	}
}

func (e EnvoyAdminCaDefaultComponent) NeedLeaderElection() bool {
	return true
}

func EnsureEnvoyAdminCaExist(
	ctx context.Context,
	resManager manager.ResourceManager,
	extensions context.Context,
) error {
	logger := kuma_log.AddFieldsFromCtx(log, ctx, extensions)
	_, err := tls.LoadCA(ctx, resManager)
	if err == nil {
		logger.V(1).Info("Envoy Admin CA already exists. Skip creating Envoy Admin CA.")
		return nil
	}
	if !store.IsResourceNotFound(err) {
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
