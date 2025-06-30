package clusterid

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	config_manager "github.com/kumahq/kuma/pkg/core/config/manager"
	config_model "github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/multitenant"
)

var log = core.Log.WithName("clusterID")

// clusterIDReader tries to read cluster ID and sets it in the runtime. Cluster ID does not change during CP lifecycle
// therefore once cluster ID is read and set, the component exits.
// In single-zone setup, followers are waiting until leader creates a cluster ID
// In multi-zone setup, the global followers and all zones waits until global leader creates a cluster ID
type clusterIDReader struct {
	rt core_runtime.Runtime
}

func (c *clusterIDReader) Start(stop <-chan struct{}) error {
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ctx = multitenant.WithTenant(ctx, multitenant.GlobalTenantID)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			clusterID, clusterCreationTime, err := c.read(ctx)
			if err != nil {
				log.Error(err, "could not read cluster ID") // just log, do not exit to retry operation
			}
			if clusterID != "" {
				log.Info("setting cluster ID", "clusterID", clusterID)
				c.rt.SetClusterInfo(clusterID, clusterCreationTime)
				return nil
			}
		case <-stop:
			return nil
		}
	}
}

func (c *clusterIDReader) NeedLeaderElection() bool {
	return false
}

func (c *clusterIDReader) read(ctx context.Context) (string, time.Time, error) {
	resource := config_model.NewConfigResource()
	err := c.rt.ConfigManager().Get(ctx, resource, store.GetByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh))
	if err != nil {
		if store.IsNotFound(err) {
			return "", time.Time{}, nil
		}
		return "", time.Time{}, err
	}
	return resource.Spec.Config, resource.Meta.GetCreationTime(), nil
}
