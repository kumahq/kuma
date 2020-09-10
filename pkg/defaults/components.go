package defaults

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var log = core.Log.WithName("defaults")

func Setup(runtime runtime.Runtime) error {
	defaultsComponent := NewDefaultsComponent(runtime.Config().Defaults, runtime.ResourceManager())
	return runtime.Add(defaultsComponent)
}

func NewDefaultsComponent(config *kuma_cp.Defaults, resManager core_manager.ResourceManager) component.Component {
	return &defaultsComponent{
		config:     config,
		resManager: resManager,
	}
}

var _ component.Component = &defaultsComponent{}

type defaultsComponent struct {
	config *kuma_cp.Defaults
	resManager core_manager.ResourceManager
}

func (d *defaultsComponent) NeedLeaderElection() bool {
	// If you spin many instances without default resources at once, many of them would create them, therefore only leader should create default resources.
	return true
}

func (d *defaultsComponent) Start(_ <-chan struct{}) error {
	// todo(jakubdyszkiewicz) once this https://github.com/kumahq/kuma/issues/1001 is done. Wait for all the components to be ready.
	if d.config.SkipMeshCreation {
		log.V(1).Info("skipping default Mesh creation because KUMA_DEFAULTS_SKIP_MESH_CREATION is set to true")
	} else {
		// Retry this operation since on Kubernetes Mesh needs to be validated and set default values.
		// This code can execute before the control plane is ready therefore hooks can fail.
		if err := retryOperation(d.createMeshIfNotExist); err != nil {
			return errors.Wrap(err, "could not create the default Mesh")
		}
	}
	return nil
}

func retryOperation(fn func() error) error {
	backoff, err := retry.NewConstant(1 * time.Second)
	if err != nil {
		return errors.Wrap(err, "invalid backoff")
	}
	backoff = retry.WithMaxDuration(1 * time.Minute, backoff)
	return retry.Do(context.Background(), backoff, func(ctx context.Context) error {
		return retry.RetryableError(fn()) // retry all errors
	})
}
