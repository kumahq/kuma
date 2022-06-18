package clusterid

import (
	"context"
	"fmt"

	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

type clusterIDCreator struct {
	configManager config_manager.ConfigManager
}

func (c *clusterIDCreator) Start(_ <-chan struct{}) error {
	return c.create()
}

func (c *clusterIDCreator) NeedLeaderElection() bool {
	return true
}

func (c *clusterIDCreator) create() error {
	resource := config_model.NewConfigResource()
	err := c.configManager.Get(context.Background(), resource, store.GetByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh))
	if err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}
		resource.Spec.Config = core.NewUUID()
		log.Info("creating cluster ID", "clusterID", resource.Spec.Config)
		if err := c.configManager.Create(context.Background(), resource, store.CreateByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh)); err != nil {
			return fmt.Errorf("could not create config: %w", err)
		}
	}

	return nil
}
