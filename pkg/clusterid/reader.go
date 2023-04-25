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
)

var log = core.Log.WithName("clusterID")

// clusterIDReader tries to read cluster ID and sets it in the runtime. Cluster ID does not change during CP lifecycle
// therefore once cluster ID is read and set, the component exits.
// In standalone setup, followers are waiting until leader creates a cluster ID
// In Multi-zone setup, the global followers and all zones waits until global leader creates a cluster ID
type clusterIDReader struct {
	rt core_runtime.Runtime
}

func (c *clusterIDReader) Start(stop <-chan struct{}) error {
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			clusterID, err := c.read(ctx)
			if err != nil {
				log.Error(err, "could not read cluster ID") // just log, do not exit to retry operation
			}
			if clusterID != "" {
				log.Info("setting cluster ID", "clusterID", clusterID)
				c.rt.SetClusterId(clusterID)
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

func (c *clusterIDReader) read(ctx context.Context) (string, error) {
	resource := config_model.NewConfigResource()
	err := c.rt.ConfigManager().Get(ctx, resource, store.GetByKey(config_manager.ClusterIdConfigKey, core_model.NoMesh))
	if err != nil {
		if store.IsResourceNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return resource.Spec.Config, nil
}
