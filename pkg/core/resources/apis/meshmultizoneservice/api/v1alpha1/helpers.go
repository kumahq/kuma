package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	core_vip "github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
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

func (m *MeshMultiZoneServiceResource) FindPortByName(name string) (core.Port, bool) {
	for _, p := range m.Spec.Ports {
		if pointer.Deref(p.Name) == name {
			return p, true
		}
		if fmt.Sprintf("%d", p.Port) == name {
			return p, true
		}
	}
	return Port{}, false
}

func (m *MeshMultiZoneServiceResource) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range m.Status.VIPs {
		for _, port := range m.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     uint32(port.Port),
				Resource: pointer.To(kri.From(m, port.GetName())),
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

func (t *MeshMultiZoneServiceResource) GetPorts() []core.Port {
	var ports []core.Port
	for _, port := range t.Spec.Ports {
		ports = append(ports, core.Port(port))
	}
	return ports
}

func (p Port) GetName() string {
	if pointer.Deref(p.Name) != "" {
		return pointer.Deref(p.Name)
	}
	return fmt.Sprintf("%d", p.Port)
}

func (p Port) GetValue() int32 {
	return p.Port
}

func (p Port) GetProtocol() core_mesh.Protocol {
	return p.AppProtocol
}
