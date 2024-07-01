package v1alpha1

import (
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
)

func (m *MeshExternalServiceResource) DestinationName(port uint32) string {
	return m.GetMeta().GetName()
}

func (m *MeshExternalServiceResource) IsReachableFromZone(zone string) bool {
	return m.GetMeta().GetLabels() == nil || m.GetMeta().GetLabels()[mesh_proto.ZoneTag] == "" || m.GetMeta().GetLabels()[mesh_proto.ZoneTag] == zone
}

var _ vip.ResourceHoldingVIPs = &MeshExternalServiceResource{}

func (t *MeshExternalServiceResource) VIPs() []string {
	if t.Status.VIP.IP == "" {
		return nil
	}
	return []string{t.Status.VIP.IP}
}

func (t *MeshExternalServiceResource) AllocateVIP(vip string) {
	t.Status.VIP = VIP{
		IP: vip,
	}
}
