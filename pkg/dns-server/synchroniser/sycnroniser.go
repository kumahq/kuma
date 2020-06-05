package synchroniser

import (
	"context"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/dns-server/resolver"
	"time"
)

var (
	synchroniserLog = core.Log.WithName("dns-server-syncroniser")
)

type (
	Syncroniser interface {
		Start(<-chan struct{}) error
	}

	ResourceSynchroniser struct {
		newTicker func() *time.Ticker
		rm        manager.ResourceManager
		resolver  resolver.DNSResolver
	}
)

const (
	topLevelDomain = ".kuma"
	tickInterval   = 500 * time.Millisecond
)

func NewResourceSynchroniser(rm manager.ResourceManager, resolver resolver.DNSResolver) (Syncroniser, error) {
	return &ResourceSynchroniser{
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
		rm:       rm,
		resolver: resolver,
	}, nil
}

func (d *ResourceSynchroniser) Start(stop <-chan struct{}) error {

	ticker := d.newTicker()
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.synchronise()
		case <-stop:
			d.synchronise()
			break
		}
	}

	return nil
}

func (d *ResourceSynchroniser) synchronise() {
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

		for _, dp := range dataplanes.Items {
			_, err := d.resolver.AddServiceToDomain(dp.Meta.GetName(), topLevelDomain)
			if err != nil {
				synchroniserLog.Error(err, "unable to synchronise")
			}
		}
	}

}
