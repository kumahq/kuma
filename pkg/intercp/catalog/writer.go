package catalog

import (
	"context"
	"sort"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	"github.com/kumahq/kuma/pkg/core/user"
)

var writerLog = core.Log.WithName("intercp").WithName("catalog").WithName("writer")

type catalogWriter struct {
	catalog    Catalog
	heartbeats *Heartbeats
	instance   Instance
	interval   time.Duration
}

var _ component.Component = &catalogWriter{}

func NewWriter(catalog Catalog, heartbeats *Heartbeats, instance Instance, interval time.Duration) component.Component {
	return &catalogWriter{
		catalog:    catalog,
		heartbeats: heartbeats,
		instance:   instance,
		interval:   interval,
	}
}

func (r *catalogWriter) Start(stop <-chan struct{}) error {
	heartbeatLog.Info("starting catalog writer")
	ctx := user.Ctx(context.Background(), user.ControlPlane)
	ticker := time.NewTicker(r.interval)
	for {
		select {
		case <-ticker.C:
			instances := r.heartbeats.ResetAndCollect()
			r.instance.Leader = true
			instances = append(instances, r.instance)
			sort.Slice(instances, func(i, j int) bool {
				return instances[i].Id < instances[j].Id
			})
			updated, err := r.catalog.Replace(ctx, instances)
			if err != nil {
				writerLog.Error(err, "could not update catalog")
				continue
			}
			if updated {
				writerLog.Info("instances catalog updated", "instances", instances)
			} else {
				writerLog.V(1).Info("no need to update instances, because the catalog is the same", "instances", instances)
			}
		case <-stop:
			return nil
		}
	}
}

func (r *catalogWriter) NeedLeaderElection() bool {
	return true
}
