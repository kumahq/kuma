package clusterid

import (
	"context"

	"github.com/pkg/errors"

	config_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
)

func Setup(rt core_runtime.Runtime) error {
	return rt.Add(&clusterID{rt: rt})
}

type clusterID struct {
	rt core_runtime.Runtime
}

func (c *clusterID) Start(_ <-chan struct{}) error {
	return createClusterID(c.rt)
}

func (c *clusterID) NeedLeaderElection() bool {
	return true
}

func createClusterID(runtime core_runtime.Runtime) error {
	resource := &config_model.ConfigResource{
		Spec: &config_proto.Config{},
	}

	err := runtime.ConfigManager().Get(context.Background(), resource, store.GetByKey(config_manager.ClusterIdConfigKey, ""))
	if err != nil {
		if !store.IsResourceNotFound(err) {
			return err
		}
		resource.Spec.Config = core.NewUUID()
		if err := runtime.ConfigManager().Create(context.Background(), resource, store.CreateByKey(config_manager.ClusterIdConfigKey, "")); err != nil {
			return errors.Wrap(err, "could not create config")
		}
	}
	clusterId := resource.Spec.Config
	runtime.SetClusterId(clusterId)

	return nil
}
