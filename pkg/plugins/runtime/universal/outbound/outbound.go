package outbound

import (
	"context"
	"time"

	"github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/dns/resolver"

	"github.com/golang/protobuf/proto"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
)

var log = core.Log.WithName("vip-outbounds-reconciler")

type VIPOutboundsReconciler struct {
	rorm      manager.ReadOnlyResourceManager
	rm        manager.ResourceManager
	resolver  resolver.DNSResolver
	newTicker func() *time.Ticker
}

func NewVIPOutboundsReconciler(rorm manager.ReadOnlyResourceManager, rm manager.ResourceManager, resolver resolver.DNSResolver, refresh time.Duration) (*VIPOutboundsReconciler, error) {
	return &VIPOutboundsReconciler{
		rorm:     rorm,
		rm:       rm,
		resolver: resolver,
		newTicker: func() *time.Ticker {
			return time.NewTicker(refresh)
		},
	}, nil
}

func (v *VIPOutboundsReconciler) NeedLeaderElection() bool {
	return true
}

func (v *VIPOutboundsReconciler) Start(stop <-chan struct{}) error {
	ticker := v.newTicker()
	defer ticker.Stop()

	log.Info("starting the VIP outbounds reconciler")
	for {
		select {
		case <-ticker.C:
			if err := v.UpdateVIPOutbounds(context.Background()); err != nil {
				log.Error(err, "errors in the VIP outbounds reconciler")
			}
		case <-stop:
			log.Info("stopping")
			return nil
		}
	}
}

func (v *VIPOutboundsReconciler) UpdateVIPOutbounds(ctx context.Context) error {
	// First get all ingresses
	// backward compatibility
	var dpIngresses []*core_mesh.DataplaneResource
	dpList := &core_mesh.DataplaneResourceList{}
	if err := v.rorm.List(ctx, dpList); err != nil {
		return err
	}
	for _, dp := range dpList.Items {
		if dp.Spec.IsIngress() {
			dpIngresses = append(dpIngresses, dp)
		}
	}

	zoneIngresses := &core_mesh.ZoneIngressResourceList{}
	if err := v.rorm.List(ctx, zoneIngresses); err != nil {
		return err
	}

	// Then add outbounds to each Dataplane
	meshes := &core_mesh.MeshResourceList{}
	if err := v.rorm.List(ctx, meshes); err != nil {
		return err
	}
	for _, m := range meshes.Items {
		dpList := &core_mesh.DataplaneResourceList{}
		if err := v.rorm.List(ctx, dpList, store.ListByMesh(m.Meta.GetName())); err != nil {
			return err
		}
		externalServices := &core_mesh.ExternalServiceResourceList{}
		if err := v.rorm.List(ctx, externalServices, store.ListByMesh(m.Meta.GetName())); err != nil {
			return err
		}
		dpsUpdated := 0

		allDps := make([]*core_mesh.DataplaneResource, len(dpIngresses)+len(dpList.Items))
		copy(allDps[:len(dpIngresses)], dpIngresses)
		copy(allDps[len(dpIngresses):], dpList.Items)

		for _, dp := range dpList.Items {
			if dp.Spec.Networking.GetTransparentProxying() == nil || dp.Spec.IsIngress() {
				continue
			}
			newOutbounds := dns.VIPOutbounds(model.MetaToResourceKey(dp.Meta), allDps, zoneIngresses.Items, v.resolver.GetVIPs(), externalServices.Items)

			if outboundsEqual(newOutbounds, dp.Spec.Networking.Outbound) {
				continue
			}
			dp.Spec.Networking.Outbound = newOutbounds
			if err := v.rm.Update(ctx, dp); err != nil {
				log.Error(err, "failed to update VIP outbounds", "dataplane", dp.GetMeta())
				continue
			}
			dpsUpdated++
			log.V(1).Info("outbounds updated", "mesh", m.Meta.GetName(), "dataplane", dp)
		}
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
