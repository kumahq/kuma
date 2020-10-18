package universal

import (
	"context"
	"reflect"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns"
)

func UpdateOutbounds(ctx context.Context, rm manager.ResourceManager, vips dns.VIPList) error {
	meshes := &mesh.MeshResourceList{}
	if err := rm.List(ctx, meshes); err != nil {
		return err
	}
	for _, m := range meshes.Items {
		dpList := &mesh.DataplaneResourceList{}
		if err := rm.List(ctx, dpList, store.ListByMesh(m.Meta.GetName())); err != nil {
			return err
		}
		externalServices := &mesh.ExternalServiceResourceList{}
		if err := rm.List(ctx, externalServices, store.ListByMesh(m.Meta.GetName())); err != nil {
			return err
		}
		dpsUpdated := 0
		for _, dp := range dpList.Items {
			if dp.Spec.Networking.GetTransparentProxying() == nil {
				continue
			}
			newOutbounds := dns.VIPOutbounds(dp.Meta.GetName(), dpList.Items, vips, externalServices.Items)
			if reflect.DeepEqual(newOutbounds, dp.Spec.Networking.Outbound) {
				continue
			}
			dp.Spec.Networking.Outbound = newOutbounds
			if err := rm.Update(ctx, dp); err != nil {
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
