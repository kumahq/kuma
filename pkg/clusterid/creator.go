package clusterid

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/multitenant"
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
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ctx = multitenant.WithTenant(ctx, multitenant.GlobalTenantID)
	resource := config_model.NewConfigResource()
	err := c.configManager.Get(ctx, resource, store.GetByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh))
	if err != nil {
		if !store.IsNotFound(err) {
			return err
		}
		resource.Spec.Config = core.NewUUID()
		log.Info("creating cluster ID", "clusterID", resource.Spec.Config)
		if err := c.configManager.Create(ctx, resource, store.CreateByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh)); err != nil {
			return errors.Wrap(err, "could not create config")
		}
	}

	return nil
}
