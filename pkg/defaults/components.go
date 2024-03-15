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
)

var log = core.Log.WithName("defaults")

func Setup(runtime runtime.Runtime) error {
	if runtime.Config().Defaults.SkipTenantResources {
		log.V(1).Info("skipping default tenant resources because KUMA_DEFAULTS_SKIP_TENANT_RESOURCES is set to true")
		return nil
	}
	if !runtime.Config().IsFederatedZoneCP() { // Don't run defaults in Zone connected to global (it's done in Global)
		defaultsComponent := NewDefaultsComponent(
			runtime.Config().Defaults,
			runtime.ResourceManager(),
			runtime.ResourceStore(),
			runtime.Extensions(),
		)

		zoneSigningKeyManager := tokens.NewSigningKeyManager(runtime.ResourceManager(), zone.SigningKeyPrefix)
		if err := runtime.Add(tokens.NewDefaultSigningKeyComponent(
			runtime.AppContext(),
			zoneSigningKeyManager,
			log.WithValues("secretPrefix", zone.SigningKeyPrefix),
			runtime.Extensions(),
		)); err != nil {
			return err
		}
		if err := runtime.Add(defaultsComponent); err != nil {
			return err
		}
	}

	if runtime.Config().Mode != config_core.Global { // Envoy Admin CA is not synced in multizone and not needed in Global CP.
		envoyAdminCaDefault := &EnvoyAdminCaDefaultComponent{
			ResManager: runtime.ResourceManager(),
			Extensions: runtime.Extensions(),
		}
		zoneDefault := &ZoneDefaultComponent{
			ResManager: runtime.ResourceManager(),
			Extensions: runtime.Extensions(),
			ZoneName:   runtime.Config().Multizone.Zone.Name,
		}
		if err := runtime.Add(envoyAdminCaDefault, zoneDefault); err != nil {
			return err
		}
	}
	return nil
}

func NewDefaultsComponent(
	config *kuma_cp.Defaults,
	resManager core_manager.ResourceManager,
	resStore store.ResourceStore,
	extensions context.Context,
) component.Component {
	return &defaultsComponent{
		config:     config,
		resManager: resManager,
		resStore:   resStore,
		extensions: extensions,
	}
}

var _ component.Component = &defaultsComponent{}

type defaultsComponent struct {
	config     *kuma_cp.Defaults
	resManager core_manager.ResourceManager
	resStore   store.ResourceStore
	extensions context.Context
}

func (d *defaultsComponent) NeedLeaderElection() bool {
	// If you spin many instances without default resources at once, many of them would create them, therefore only leader should create default resources.
	return true
}

func (d *defaultsComponent) Start(stop <-chan struct{}) error {
	// todo(jakubdyszkiewicz) once this https://github.com/kumahq/kuma/issues/1001 is done. Wait for all the components to be ready.
	ctx, cancelFn := context.WithCancel(user.Ctx(context.Background(), user.ControlPlane))
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
				return retry.RetryableError(func() error {
					_, err := CreateMeshIfNotExist(ctx, d.resManager, d.extensions)
					return err
				}()) // retry all errors
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
