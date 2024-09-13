package v1alpha1

import (
	"fmt"

	core_vip "github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

var _ core_vip.ResourceHoldingVIPs = &MeshMultiZoneServiceResource{}

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

func (m *MeshMultiZoneServiceResource) FindPortByName(name string) (meshservice_api.Port, bool) {
	for _, p := range m.Spec.Ports {
		if p.Name == name {
			return p, true
		}
		if fmt.Sprintf("%d", p.Port) == name {
			return p, true
		}
	}
	return meshservice_api.Port{}, false
}

func (m *MeshMultiZoneServiceResource) DestinationName(port uint32) string {
	id := model.NewResourceIdentifier(m)
	return fmt.Sprintf("%s_%s_%s_%s_mzsvc_%d", id.Mesh, id.Name, id.Namespace, id.Zone, port)
}

func (m *MeshMultiZoneServiceResource) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range m.Status.VIPs {
		for _, port := range m.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     port.Port,
				Resource: pointer.To(model.NewTypedResourceIdentifier(m, model.WithSectionName(port.GetName()))),
			})
		}
	}
	return outbounds
}

func (t *MeshMultiZoneServiceResource) Domains() *xds_types.VIPDomains {
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
