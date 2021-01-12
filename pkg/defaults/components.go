package defaults

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var log = core.Log.WithName("defaults")

func Setup(runtime runtime.Runtime) error {
	defaultsComponent := NewDefaultsComponent(runtime.Config().Defaults, runtime.Config().Mode, runtime.Config().Environment, runtime.ResourceManager(), runtime.ResourceStore())
	return runtime.Add(defaultsComponent)
}

func NewDefaultsComponent(config *kuma_cp.Defaults, cpMode config_core.CpMode, environment config_core.EnvironmentType, resManager core_manager.ResourceManager, resStore store.ResourceStore) component.Component {
	return &defaultsComponent{
		cpMode:      cpMode,
		environment: environment,
		config:      config,
		resManager:  resManager,
		resStore:    resStore,
	}
}

var _ component.Component = &defaultsComponent{}

type defaultsComponent struct {
	cpMode      config_core.CpMode
	environment config_core.EnvironmentType
	config      *kuma_cp.Defaults
	resManager  core_manager.ResourceManager
	resStore    store.ResourceStore
}

func (d *defaultsComponent) NeedLeaderElection() bool {
	// If you spin many instances without default resources at once, many of them would create them, therefore only leader should create default resources.
	return true
}

func (d *defaultsComponent) Start(_ <-chan struct{}) error {
	// todo(jakubdyszkiewicz) once this https://github.com/kumahq/kuma/issues/1001 is done. Wait for all the components to be ready.
	if d.config.SkipMeshCreation {
		log.V(1).Info("skipping default Mesh creation because KUMA_DEFAULTS_SKIP_MESH_CREATION is set to true")
	} else if err := doWithRetry(d.createMeshIfNotExist); err != nil {
		// Retry this operation since on Kubernetes Mesh needs to be validated and set default values.
		// This code can execute before the control plane is ready therefore hooks can fail.
		return errors.Wrap(err, "could not create the default Mesh")
	}
	return nil
}

func doWithRetry(fn func() error) error {
	backoff, _ := retry.NewConstant(5 * time.Second)
	backoff = retry.WithMaxDuration(10*time.Minute, backoff) // if after this time we cannot create a resource - something is wrong and we should return an error which will restart CP.
	return retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		return retry.RetryableError(fn()) // retry all errors
	})
}
