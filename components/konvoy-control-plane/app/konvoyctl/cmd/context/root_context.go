package context

import (
	"time"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/config"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/pkg/errors"

	konvoyctl_resources "github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/resources"
)

type rootArgs struct {
	ConfigFile string
	Mesh       string
	Debug      bool
}

type RootRuntime struct {
	Config           config_proto.Configuration
	Now              func() time.Time
	NewResourceStore func(*config_proto.ControlPlane) (core_store.ResourceStore, error)
}

type RootContext struct {
	Args    rootArgs
	Runtime RootRuntime
}

func DefaultRootContext() *RootContext {
	return &RootContext{
		Runtime: RootRuntime{
			Now:              time.Now,
			NewResourceStore: konvoyctl_resources.NewResourceStore,
		},
	}
}

func (rc *RootContext) LoadConfig() error {
	return config.Load(rc.Args.ConfigFile, &rc.Runtime.Config)
}

func (rc *RootContext) SaveConfig() error {
	return config.Save(rc.Args.ConfigFile, &rc.Runtime.Config)
}

func (rc *RootContext) Config() *config_proto.Configuration {
	return &rc.Runtime.Config
}

func (rc *RootContext) CurrentContext() (*config_proto.Context, error) {
	if rc.Config().CurrentContext == "" {
		return nil, errors.Errorf("Current context is not set. Use `konvoyctl config control-planes add` to add an existing Control Plane.")
	}
	_, currentContext := rc.Config().GetContext(rc.Config().CurrentContext)
	if currentContext == nil {
		return nil, errors.Errorf("Current context is broken. Use `konvoyctl config control-planes add` to add an existing Control Plane once again.")
	}
	return currentContext, nil
}

func (rc *RootContext) CurrentControlPlane() (*config_proto.ControlPlane, error) {
	currentContext, err := rc.CurrentContext()
	if err != nil {
		return nil, err
	}
	_, controlPlane := rc.Config().GetControlPlane(currentContext.ControlPlane)
	if controlPlane == nil {
		return nil, errors.Errorf("Current context refers to a Control Plane that doesn't exist: %q", currentContext.ControlPlane)
	}
	return controlPlane, nil
}

func (rc *RootContext) CurrentMesh() string {
	if rc.Args.Mesh != "" {
		return rc.Args.Mesh
	}
	return core_model.DefaultMesh
}

func (rc *RootContext) Now() time.Time {
	return rc.Runtime.Now()
}

func (rc *RootContext) NewResourceStore(cp *config_proto.ControlPlane) (core_store.ResourceStore, error) {
	return rc.Runtime.NewResourceStore(cp)
}
