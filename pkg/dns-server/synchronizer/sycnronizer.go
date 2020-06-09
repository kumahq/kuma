package synchronizer

import (
	"context"
	"time"

	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
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
		domain    string
		rm        manager.ResourceManager
		resolver  resolver.DNSResolver
		newTicker func() *time.Ticker
	}
)

const (
	tickInterval = 500 * time.Millisecond
)

func NewResourceSynchronizer(domain string, rm manager.ResourceManager, resolver resolver.DNSResolver) (Synchronizer, error) {
	return &ResourceSynchronizer{
		domain:   domain,
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
			d.synchronise()
			return nil
		}
	}
}

func (d *ResourceSynchronizer) synchronise() {
	meshes := mesh.MeshResourceList{}

	err := d.rm.List(context.Background(), &meshes)
	if err != nil {
		synchroniserLog.Error(err, "unable to synchronise")
		return
	}

	for _, m := range meshes.Items {
		dataplanes := mesh.DataplaneResourceList{}

		err := d.rm.List(context.Background(), &dataplanes, store.ListByMesh(m.Meta.GetName()))
		if err != nil {
			synchroniserLog.Error(err, "unable to synchronise", "mesh", m.Meta.GetName())
		}

		serviceMap := make(map[string]bool)

		// TODO: Do we need to reflect somehow the fact this service belongs to a particular `mesh`
		for _, dp := range dataplanes.Items {
			for _, inbound := range dp.Spec.Networking.Inbound {
				serviceMap[inbound.GetService()] = true
			}
		}

		err = d.resolver.SyncServicesForDomain(serviceMap, d.domain)
		if err != nil {
			synchroniserLog.Error(err, "unable to synchronise", "serviceMap", serviceMap)
		}
	}
}
