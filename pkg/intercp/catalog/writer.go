package catalog

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var writerLog = core.Log.WithName("intercp").WithName("catalog").WithName("writer")

type catalogWriter struct {
	catalog    Catalog
	heartbeats *Heartbeats
	instance   Instance
	interval   time.Duration
	ctx        context.Context
}

var _ component.Component = &catalogWriter{}

func NewWriter(ctx context.Context, catalog Catalog, heartbeats *Heartbeats, instance Instance, interval time.Duration) component.Component {
	leaderInstance := instance
	leaderInstance.Leader = true
	return &catalogWriter{
		catalog:    catalog,
		heartbeats: heartbeats,
		instance:   leaderInstance,
		interval:   interval,
		ctx: ctx,
	}
}

func (r *catalogWriter) Start(stop <-chan struct{}) error {
	heartbeatLog.Info("starting catalog writer")
	writerLog.Info("replacing a leader in the catalog")
	if err := r.catalog.ReplaceLeader(r.ctx, r.instance); err != nil {
		writerLog.Error(err, "could not replace leader") // continue, it will be replaced in ticker anyways
	}
	ticker := time.NewTicker(r.interval)
	for {
		select {
		case <-ticker.C:
			instances := r.heartbeats.ResetAndCollect()
			instances = append(instances, r.instance)
			updated, err := r.catalog.Replace(r.ctx, instances)
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
