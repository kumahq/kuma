package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/destinationname"
	core_vip "github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
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

func (m *MeshMultiZoneServiceResource) findPort(port uint32) (Port, bool) {
	for _, p := range m.Spec.Ports {
		if p.Port == port {
			return p, true
		}
	}
	return Port{}, false
}

func (m *MeshMultiZoneServiceResource) FindSectionNameByPort(port uint32) (string, bool) {
	if port, found := m.findPort(port); found {
		return port.GetNameOrStringifyPort(), true
	}
	return "", false
}

func (m *MeshMultiZoneServiceResource) FindPortByName(name string) (Port, bool) {
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

func (m *MeshMultiZoneServiceResource) DestinationName(port uint32) string {
	return destinationname.LegacyName(kri.From(m, ""), MeshMultiZoneServiceResourceTypeDescriptor.ShortName, port)
}

func (m *MeshMultiZoneServiceResource) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range m.Status.VIPs {
		for _, port := range m.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     port.Port,
				Resource: pointer.To(kri.From(m, port.GetNameOrStringifyPort())),
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

func (p Port) GetNameOrStringifyPort() string {
	if pointer.Deref(p.Name) != "" {
		return pointer.Deref(p.Name)
	}
	return fmt.Sprintf("%d", p.Port)
}

func (p Port) GetName() string {
	return pointer.Deref(p.Name)
}

func (p Port) GetValue() uint32 {
	return p.Port
}
