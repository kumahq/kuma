package v1alpha1

import (
	"fmt"
	"strings"

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

func (m *MeshMultiZoneServiceResource) FindPort(port uint32) (meshservice_api.Port, bool) {
	for _, p := range m.Spec.Ports {
		if p.Port == port {
			return p, true
		}
	}
	return meshservice_api.Port{}, false
}

func (m *MeshMultiZoneServiceResource) DestinationName(port uint32) string {
	return fmt.Sprintf("%s_mzsvc_%d", strings.ReplaceAll(m.GetMeta().GetName(), ".", "_"), port)
}
