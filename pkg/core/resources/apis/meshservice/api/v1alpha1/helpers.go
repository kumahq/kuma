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

func (m *MeshServiceResource) FindPortByName(name string) (Port, bool) {
	for _, p := range m.Spec.Ports {
		if p.Name == name {
			return p, true
		}
		if fmt.Sprintf("%d", p.Port) == name {
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

// todo(jakubdyszkiewicz) strongly consider putting this in MeshService object to avoid problems with computation
func (t *MeshServiceResource) SNIName(systemNamespace string) string {
	displayName := t.GetMeta().GetLabels()[mesh_proto.DisplayName]
	namespace := t.GetMeta().GetLabels()[mesh_proto.KubeNamespaceTag]
	origin := t.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if origin == string(mesh_proto.GlobalResourceOrigin) {
		// we need to use original name and namespace for services that were synced from another cluster
		sniName := displayName
		if namespace != "" {
			// when we sync resources from universal to kube, when we retrieve it has KubeNamespaceTag as label value
			if systemNamespace == "" || systemNamespace != namespace {
				sniName += "." + namespace
			}
		}
		return sniName
	}
	if systemNamespace == "" && origin == string(mesh_proto.ZoneResourceOrigin) && namespace != "" {
		// when namespace label was added to Universal MeshService to have a copy of Kubernets MeshService
		return t.GetMeta().GetName() + "." + namespace
	}
	return t.GetMeta().GetName()
}

func (t *MeshServiceResource) Default() error {
	if t.Spec.State == "" {
		t.Spec.State = StateUnavailable
	}
	return nil
}

func (p *Port) GetName() string {
	if p.Name != "" {
		return p.Name
	}
	return fmt.Sprintf("%d", p.Port)
}
