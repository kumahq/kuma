package universal

import (
	"context"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/dns"
)

func UpdateOutbounds(ctx context.Context, rm manager.ResourceManager, vips dns.VIPList) error {
	dpList := &mesh.DataplaneResourceList{}
	if err := rm.List(ctx, dpList); err != nil {
		return err
	}
	for _, dp := range dpList.Items {
		if dp.Spec.Networking.GetTransparentProxying() == nil {
			continue
		}
		dp.Spec.Networking.Outbound = dns.VIPOutbounds(dp.Meta.GetName(), dpList.Items, vips)
		if err := rm.Update(ctx, dp); err != nil {
			log.Error(err, "failed to update VIP outbounds", "dataplane", dp.GetMeta())
			continue
		}
		log.V(0).Info("outbounds updated", "dataplane", dp)
	}
	return nil
}
