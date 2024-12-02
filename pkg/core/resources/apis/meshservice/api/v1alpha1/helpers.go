package v1alpha1

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_vip "github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (m *MeshServiceResource) DestinationName(port uint32) string {
	id := model.NewResourceIdentifier(m)
	return fmt.Sprintf("%s_%s_%s_%s_%s_%d", id.Mesh, id.Name, id.Namespace, id.Zone, MeshServiceResourceTypeDescriptor.ShortName, port)
}

func (m *MeshServiceResource) findPort(port uint32) (Port, bool) {
	for _, p := range m.Spec.Ports {
		if p.Port == port {
			return p, true
		}
	}
	return Port{}, false
}

func (m *MeshServiceResource) FindSectionNameByPort(port uint32) (string, bool) {
	if port, found := m.findPort(port); found {
		return port.GetName(), true
	}
	return "", false
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

func (m *MeshServiceResource) IsLocalMeshService() bool {
	if len(m.GetMeta().GetLabels()) == 0 {
		return true // no labels mean that it's a local resource
	}
	origin, ok := m.GetMeta().GetLabels()[mesh_proto.ResourceOriginLabel]
	if !ok {
		return true // no zone label mean that it's a local resource
	}
	return origin == string(mesh_proto.ZoneResourceOrigin)
}

var _ core_vip.ResourceHoldingVIPs = &MeshServiceResource{}

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

func (t *MeshServiceResource) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range t.Status.VIPs {
		for _, port := range t.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     port.Port,
				Resource: pointer.To(model.NewTypedResourceIdentifier(t, model.WithSectionName(port.GetName()))),
			})
		}
	}
	return outbounds
}

func (t *MeshServiceResource) Domains() *xds_types.VIPDomains {
	if len(t.Status.VIPs) > 0 {
		var domains []string
		for _, addr := range t.Status.Addresses {
			domains = append(domains, addr.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: t.Status.VIPs[0].IP,
			Domains: domains,
		}
	}
	return nil
}

func (p *Port) GetName() string {
	if p.Name != "" {
		return p.Name
	}
	return fmt.Sprintf("%d", p.Port)
}
