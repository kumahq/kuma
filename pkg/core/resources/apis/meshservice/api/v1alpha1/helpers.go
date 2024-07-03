package v1alpha1

import (
	"fmt"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
)

func (m *MeshServiceResource) DestinationName(port uint32) string {
	return fmt.Sprintf("%s_svc_%d", strings.ReplaceAll(m.GetMeta().GetName(), ".", "_"), port)
}

func (m *MeshServiceResource) FindPort(port uint32) (Port, bool) {
	for _, p := range m.Spec.Ports {
		if p.Port == port {
			return p, true
		}
	}
	return Port{}, false
}

func (m *MeshServiceResource) IsLocalMeshService(localZone string) bool {
	if len(m.GetMeta().GetLabels()) == 0 {
		return true // no labels mean that it's a local resource
	}
	resZone, ok := m.GetMeta().GetLabels()[mesh_proto.ZoneTag]
	if !ok {
		return true // no zone label mean that it's a local resource
	}
	return resZone == localZone
}

var _ vip.ResourceHoldingVIPs = &MeshServiceResource{}

func (t *MeshServiceResource) VIPs() []string {
	vips := make([]string, len(t.Status.VIPs))
	for i, vip := range t.Status.VIPs {
		vips[i] = vip.IP
	}
	return vips
}

func (t *MeshServiceResource) AllocateVIP(vip string) {
	t.Status.VIPs = append(t.Status.VIPs, VIP{
		IP: vip,
	})
}
