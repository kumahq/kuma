package defaults

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"
	"go.uber.org/multierr"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/tokens"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zone"
	"github.com/kumahq/kuma/pkg/tokens/builtin/zoneingress"
	"github.com/kumahq/kuma/pkg/util/iterator"
)

var log = core.Log.WithName("defaults")

func Setup(runtime runtime.Runtime) error {
	if runtime.Config().Mode != config_core.Zone { // Don't run defaults in Zone (it's done in Global)
		// todo(jakubdyszkiewicz) once this https://github.com/kumahq/kuma/issues/1001 is done. Wait for all the components to be ready.
		contexts, err := iterator.CustomIterator()
		if err != nil {
			return err
		}
		var allErrors error
		for _, ctx := range contexts {
			ctx = user.Ctx(ctx, user.ControlPlane)
			defaultsComponent := NewDefaultsComponent(ctx, runtime.Config().Defaults, runtime.Config().Mode, runtime.Config().Environment, runtime.ResourceManager(), runtime.ResourceStore())
			if err := runtime.Add(defaultsComponent); err != nil {
				allErrors = multierr.Append(allErrors, err)
			}

			zoneIngressSigningKeyManager := tokens.NewSigningKeyManager(runtime.ResourceManager(), zoneingress.ZoneIngressSigningKeyPrefix)
			if err := runtime.Add(tokens.NewDefaultSigningKeyComponent(ctx, zoneIngressSigningKeyManager, log.WithValues("secretPrefix", zoneingress.ZoneIngressSigningKeyPrefix))); err != nil {
				allErrors = multierr.Append(allErrors, err)
			}

			zoneSigningKeyManager := tokens.NewSigningKeyManager(runtime.ResourceManager(), zone.SigningKeyPrefix)
			if err := runtime.Add(tokens.NewDefaultSigningKeyComponent(ctx, zoneSigningKeyManager, log.WithValues("secretPrefix", zoneingress.ZoneIngressSigningKeyPrefix))); err != nil {
				allErrors = multierr.Append(allErrors, err)
			}
		}
		if allErrors != nil {
			return allErrors
		}
	}

	if runtime.Config().Mode != config_core.Global { // Envoy Admin CA is not synced in multizone and not needed in Global CP.
		contexts, err := iterator.CustomIterator()
		if err != nil {
			return err
		}
		var allErrors error
		for _, ctx := range contexts {
			if err := runtime.Add(NewEnvoyAdminCaDefaultComponent(ctx, runtime.ResourceManager())); err != nil {
				allErrors = multierr.Append(allErrors, err)
			}
		}
		if allErrors != nil {
			return allErrors
		}
	}
	return nil
}

func NewDefaultsComponent(ctx context.Context, config *kuma_cp.Defaults, cpMode config_core.CpMode, environment config_core.EnvironmentType, resManager core_manager.ResourceManager, resStore store.ResourceStore) component.Component {
	return &defaultsComponent{
		cpMode:      cpMode,
		environment: environment,
		config:      config,
		resManager:  resManager,
		resStore:    resStore,
		ctx:         ctx,
	}
}

var _ component.Component = &defaultsComponent{}

type defaultsComponent struct {
	cpMode      config_core.CpMode
	environment config_core.EnvironmentType
	config      *kuma_cp.Defaults
	resManager  core_manager.ResourceManager
	resStore    store.ResourceStore
	ctx         context.Context
}

func (d *defaultsComponent) NeedLeaderElection() bool {
	// If you spin many instances without default resources at once, many of them would create them, therefore only leader should create default resources.
	return true
}

func (d *defaultsComponent) Start(stop <-chan struct{}) error {
	// todo(jakubdyszkiewicz) once this https://github.com/kumahq/kuma/issues/1001 is done. Wait for all the components to be ready.
	ctx, cancelFn := context.WithCancel(d.ctx)
	defer cancelFn()
	wg := &sync.WaitGroup{}
	errChan := make(chan error)

	if d.config.SkipMeshCreation {
		log.V(1).Info("skipping default Mesh creation because KUMA_DEFAULTS_SKIP_MESH_CREATION is set to true")
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// if after this time we cannot create a resource - something is wrong and we should return an error which will restart CP.
			err := retry.Do(ctx, retry.WithMaxDuration(10*time.Minute, retry.NewConstant(5*time.Second)), func(ctx context.Context) error {
				return retry.RetryableError(CreateMeshIfNotExist(ctx, d.resManager)) // retry all errors
			})
			if err != nil {
				// Retry this operation since on Kubernetes Mesh needs to be validated and set default values.
				// This code can execute before the control plane is ready therefore hooks can fail.
				errChan <- errors.Wrap(err, "could not create the default Mesh")
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
		close(errChan)
	}()

	var errs error
	for {
		select {
		case <-stop:
			return errs
		case err := <-errChan:
			errs = multierr.Append(errs, err)
		case <-done:
			return errs
		}
	}
}
