package v1alpha1

import (
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
)

var _ vip.ResourceHoldingVIPs = &MeshMultiZoneServiceResource{}

func (t *MeshMultiZoneServiceResource) VIPs() []string {
	vips := make([]string, len(t.Status.VIPs))
	for i, vip := range t.Status.VIPs {
		vips[i] = vip.IP
	}
	return vips
}

func (t *MeshMultiZoneServiceResource) AllocateVIP(vip string) {
	t.Status.VIPs = append(t.Status.VIPs, meshservice_api.VIP{
		IP: vip,
	})
}
