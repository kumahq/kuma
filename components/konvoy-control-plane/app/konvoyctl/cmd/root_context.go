package cmd

import (
	"time"

	"github.com/Kong/konvoy/components/konvoy-control-plane/app/konvoyctl/pkg/config"
	config_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoyctl/v1alpha1"
	core_store "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/pkg/errors"
)

func (rc *rootContext) LoadConfig() error {
	return config.Load(rc.args.configFile, &rc.runtime.config)
}

func (rc *rootContext) SaveConfig() error {
	return config.Save(rc.args.configFile, &rc.runtime.config)
}

func (rc *rootContext) Config() *config_proto.Configuration {
	return &rc.runtime.config
}

func (rc *rootContext) CurrentContext() (*config_proto.Context, error) {
	if rc.Config().CurrentContext == "" {
		return nil, errors.Errorf("Current context is not set. Use `konvoyctl config control-planes add` to add an existing Control Plane.")
	}
	_, currentContext := rc.Config().GetContext(rc.Config().CurrentContext)
	if currentContext == nil {
		return nil, errors.Errorf("Current context is broken. Use `konvoyctl config control-planes add` to add an existing Control Plane once again.")
	}
	return currentContext, nil
}

func (rc *rootContext) CurrentControlPlane() (*config_proto.ControlPlane, error) {
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

func (rc *rootContext) Now() time.Time {
	return rc.runtime.now()
}

func (rc *rootContext) NewResourceStore(cp *config_proto.ControlPlane) (core_store.ResourceStore, error) {
	return rc.runtime.newResourceStore(cp)
}
