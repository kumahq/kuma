package outbound

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/dns/resolver"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
)

var log = core.Log.WithName("dns-vips-allocator")

type OutboundsLoop struct {
	rorm      manager.ReadOnlyResourceManager
	rm        manager.ResourceManager
	resolver  resolver.DNSResolver
	newTicker func() *time.Ticker
}

const (
	tickInterval = 500 * time.Millisecond
)

func NewOutboundsLoop(rorm manager.ReadOnlyResourceManager, rm manager.ResourceManager, resolver resolver.DNSResolver) (*OutboundsLoop, error) {
	return &OutboundsLoop{
		rorm:     rorm,
		rm:       rm,
		resolver: resolver,
		newTicker: func() *time.Ticker {
			return time.NewTicker(tickInterval)
		},
	}, nil
}

func (o *OutboundsLoop) NeedLeaderElection() bool {
	return true
}

func (o *OutboundsLoop) Start(stop <-chan struct{}) error {
	ticker := o.newTicker()
	defer ticker.Stop()

	log.Info("starting the outbounds loop")
	for {
		select {
		case <-ticker.C:
			if err := o.UpdateOutbounds(context.Background()); err != nil {
				log.Error(err, "errors in the outbounds loop")
			}
		case <-stop:
			log.Info("stopping")
			return nil
		}
	}
}

func (o *OutboundsLoop) UpdateOutbounds(ctx context.Context) error {
	meshes := &mesh.MeshResourceList{}
	if err := o.rorm.List(ctx, meshes); err != nil {
		return err
	}
	for _, m := range meshes.Items {
		dpList := &mesh.DataplaneResourceList{}
		if err := o.rorm.List(ctx, dpList, store.ListByMesh(m.Meta.GetName())); err != nil {
			return err
		}
		externalServices := &mesh.ExternalServiceResourceList{}
		if err := o.rorm.List(ctx, externalServices, store.ListByMesh(m.Meta.GetName())); err != nil {
			return err
		}
		dpsUpdated := 0
		for _, dp := range dpList.Items {
			if dp.Spec.Networking.GetTransparentProxying() == nil {
				continue
			}
			newOutbounds := dns.VIPOutbounds(dp.Meta.GetName(), dpList.Items, o.resolver.GetVIPs(), externalServices.Items)

			if outboundsEqual(newOutbounds, dp.Spec.Networking.Outbound) {
				continue
			}
			dp.Spec.Networking.Outbound = newOutbounds
			if err := o.rm.Update(ctx, dp); err != nil {
				log.Error(err, "failed to update VIP outbounds", "dataplane", dp.GetMeta())
				continue
			}
			dpsUpdated++
			log.V(1).Info("outbounds updated", "mesh", m.Meta.GetName(), "dataplane", dp)
		}
		log.Info("outbounds updated due to VIP changes", "mesh", m.Meta.GetName(), "dpsUpdated", dpsUpdated)
	}
	return nil
}

func outboundsEqual(outbounds []*mesh_proto.Dataplane_Networking_Outbound, other []*mesh_proto.Dataplane_Networking_Outbound) bool {
	if len(outbounds) != len(other) {
		return false
	}
	for i := range outbounds {
		if !proto.Equal(outbounds[i], other[i]) {
			return false
		}
	}
	return true
}
