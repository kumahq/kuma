package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core"
	core_vip "github.com/kumahq/kuma/v2/pkg/core/resources/apis/core/vip"
	meshservice_api "github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshservice/api/v1alpha1"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var _ core_vip.ResourceHoldingVIPs = &MeshMultiZoneServiceResource{}

func (r *MeshMultiZoneServiceResource) VIPs() []string {
	vips := make([]string, len(r.Status.VIPs))
	for i, vip := range r.Status.VIPs {
		vips[i] = vip.IP
	}
	return vips
}

func (r *MeshMultiZoneServiceResource) AllocateVIP(vip string) {
	r.Status.VIPs = append(r.Status.VIPs, meshservice_api.VIP{
		IP: vip,
	})
}

func (r *MeshMultiZoneServiceResource) FindPortByName(name string) (core.Port, bool) {
	for _, p := range r.Spec.Ports {
		if pointer.Deref(p.Name) == name {
			return p, true
		}
		if fmt.Sprintf("%d", p.Port) == name {
			return p, true
		}
	}
	return Port{}, false
}

func (r *MeshMultiZoneServiceResource) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range r.Status.VIPs {
		for _, port := range r.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     uint32(port.Port),
				Resource: kri.WithSectionName(kri.From(r), port.GetName()),
			})
		}
	}
	return outbounds
}

func (r *MeshMultiZoneServiceResource) Domains() *xds_types.VIPDomains {
	if len(r.Status.VIPs) > 0 {
		var domains []string
		for _, addr := range r.Status.Addresses {
			domains = append(domains, addr.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: r.Status.VIPs[0].IP,
			Domains: domains,
		}
	}
	return nil
}

func (r *MeshMultiZoneServiceResource) GetPorts() []core.Port {
	var ports []core.Port
	for _, port := range r.Spec.Ports {
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

func (p Port) GetProtocol() core_meta.Protocol {
	return p.AppProtocol
}

func (l *MeshMultiZoneServiceResourceList) GetDestinations() []core.Destination {
	var result []core.Destination
	for _, item := range l.Items {
		result = append(result, item)
	}
	return result
}
