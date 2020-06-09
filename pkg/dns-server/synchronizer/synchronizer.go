package synchronizer

import (
	"context"
	"time"

	"github.com/Kong/kuma/pkg/core"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/dns-server/resolver"
)

var (
	synchroniserLog = core.Log.WithName("dns-server-synchronizer")
)

type (
	Synchronizer interface {
		Start(<-chan struct{}) error
	}

	ResourceSynchronizer struct {
		rm        manager.ReadOnlyResourceManager
		resolver  resolver.DNSResolver
		newTicker func() *time.Ticker
	}
)

const (
	tickInterval = 500 * time.Millisecond
)

func NewResourceSynchronizer(rm manager.ReadOnlyResourceManager, resolver resolver.DNSResolver) (Synchronizer, error) {
	return &ResourceSynchronizer{
		rm:       rm,
		resolver: resolver,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}, nil
}

func (d *ResourceSynchronizer) Start(stop <-chan struct{}) error {
	ticker := d.newTicker()
	defer ticker.Stop()

	synchroniserLog.Info("Starting the syncroniser")
	for {
		select {
		case <-ticker.C:
			d.synchronise()
		case <-stop:
			return nil
		}
	}
}

func (d *ResourceSynchronizer) synchronise() {
	meshes := core_mesh.MeshResourceList{}

	err := d.rm.List(context.Background(), &meshes)
	if err != nil {
		synchroniserLog.Error(err, "unable to synchronise")
		return
	}

	for _, mesh := range meshes.Items {
		dataplanes := core_mesh.DataplaneResourceList{}

		err := d.rm.List(context.Background(), &dataplanes, store.ListByMesh(mesh.Meta.GetName()))
		if err != nil {
			synchroniserLog.Error(err, "unable to synchronise", "mesh", mesh.Meta.GetName())
		}

		serviceMap := make(map[string]bool)

		// TODO: Do we need to reflect somehow the fact this service belongs to a particular `mesh`
		for _, dp := range dataplanes.Items {
			for _, inbound := range dp.Spec.Networking.Inbound {
				serviceMap[inbound.GetService()] = true
			}
		}

		err = d.resolver.SyncServices(serviceMap)
		if err != nil {
			synchroniserLog.Error(err, "unable to synchronise", "serviceMap", serviceMap)
		}
	}
}
