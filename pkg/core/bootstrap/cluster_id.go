package bootstrap

import (
	"context"

	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"

	"github.com/pkg/errors"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"

	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func createClusterID(runtime core_runtime.Runtime) error {
	manager := runtime.ConfigManager()
	resource := &config_model.ConfigResource{}

	err := manager.Get(context.Background(), resource, store.GetByKey(config_manager.ClusterIdConfigKey, ""))
	if err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}

		switch runtime.Config().Mode {
		case config_core.Standalone:
			fallthrough
		case config_core.Global:
			resource.Spec.Config = core.NewUUID()
			if err := manager.Create(context.Background(), resource, store.CreateByKey(config_manager.ClusterIdConfigKey, "")); err != nil {
				return errors.Wrap(err, "could not create config")
			}
		}
	}
	clusterId := resource.Spec.Config
	if err := runtime.SetClusterId(clusterId); err != nil {
		return errors.Wrap(err, "could not set Cluster ID")
	}

	return nil
}
