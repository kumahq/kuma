package clusterid

import (
	"context"
	"github.com/hashicorp/go-multierror"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/util/iterator"
	"github.com/pkg/errors"

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
	contexts, err := iterator.CustomIterator()
	if err != nil {
		log.Error(err, "could not get contexts")
	}
	var allErrors error
	for _, ctx := range contexts {
		allErrors = multierror.Append(allErrors, Create(user.Ctx(ctx, user.ControlPlane), c.configManager))
	}
	return allErrors
}

func Create(ctx context.Context, configManager config_manager.ConfigManager) error {
	resource := config_model.NewConfigResource()
	err := configManager.Get(ctx, resource, store.GetByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh))
	if err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}
		resource.Spec.Config = core.NewUUID()
		log.Info("creating cluster ID", "clusterID", resource.Spec.Config)
		if err := configManager.Create(ctx, resource, store.CreateByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh)); err != nil {
			return errors.Wrap(err, "could not create config")
		}
	}
	return nil
}
