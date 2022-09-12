package gc

import (
	"context"
	"fmt"
	"time"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var (
	gcLog = core.Log.WithName("dataplane-gc")
)

type collector struct {
	rm         manager.ResourceManager
	cleanupAge time.Duration
	newTicker  func() *time.Ticker
}

func NewCollector(rm manager.ResourceManager, newTicker func() *time.Ticker, cleanupAge time.Duration) component.Component {
	return &collector{
		cleanupAge: cleanupAge,
		rm:         rm,
		newTicker:  newTicker,
	}
}

func (d *collector) Start(stop <-chan struct{}) error {
	ticker := d.newTicker()
	defer ticker.Stop()
	gcLog.Info("started")
	ctx := context.Background()
	for {
		select {
		case now := <-ticker.C:
			if err := d.cleanup(ctx, now); err != nil {
				gcLog.Error(err, "unable to cleanup")
				continue
			}
		case <-stop:
			gcLog.Info("stopped")
			return nil
		}
	}
}

func (d *collector) cleanup(ctx context.Context, now time.Time) error {
	dataplaneInsights := &core_mesh.DataplaneInsightResourceList{}
	if err := d.rm.List(ctx, dataplaneInsights); err != nil {
		return err
	}
	onDelete := []model.ResourceKey{}
	for _, di := range dataplaneInsights.Items {
		if di.Spec.IsOnline() {
			continue
		}
		if s := di.Spec.GetLastSubscription().(*mesh_proto.DiscoverySubscription); s != nil {
			if err := s.GetDisconnectTime().CheckValid(); err != nil {
				gcLog.Error(err, "unable to parse DisconnectTime", "disconnect time", s.GetDisconnectTime(), "mesh", di.GetMeta().GetMesh(), "dataplane", di.GetMeta().GetName())
				continue
			}
			if now.Sub(s.GetDisconnectTime().AsTime()) > d.cleanupAge {
				onDelete = append(onDelete, model.ResourceKey{Name: di.GetMeta().GetName(), Mesh: di.GetMeta().GetMesh()})
			}
		}
	}
	for _, rk := range onDelete {
		gcLog.Info(fmt.Sprintf("deleting dataplane which is offline for %v", d.cleanupAge), "name", rk.Name, "mesh", rk.Mesh)
		if err := d.rm.Delete(ctx, core_mesh.NewDataplaneResource(), store.DeleteBy(rk)); err != nil {
			gcLog.Error(err, "unable to delete dataplane", "name", rk.Name, "mesh", rk.Mesh)
			continue
		}
	}
	return nil
}

func (d *collector) NeedLeaderElection() bool {
	return true
}
