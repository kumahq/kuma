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

type Destination struct {
	MeshMultiZoneServiceResource
}

func ToDst(mz *MeshMultiZoneServiceResource) *Destination {
	return &Destination{*mz}
}

var _ core_vip.ResourceHoldingVIPs = &Destination{}

func (d *Destination) VIPs() []string {
	vips := make([]string, len(d.Status.VIPs))
	for i, vip := range d.Status.VIPs {
		vips[i] = vip.IP
	}
	return vips
}

func (d *Destination) AllocateVIP(vip string) {
	d.Status.VIPs = append(d.Status.VIPs, meshservice_api.VIP{
		IP: vip,
	})
}

func (d *Destination) FindPortByName(name string) (core.Port, bool) {
	for _, p := range d.Spec.Ports {
		if pointer.Deref(p.Name) == name {
			return p, true
		}
		if fmt.Sprintf("%d", p.Port) == name {
			return p, true
		}
	}
	return Port{}, false
}

func (d *Destination) AsOutbounds() xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, vip := range d.Status.VIPs {
		for _, port := range d.Spec.Ports {
			outbounds = append(outbounds, &xds_types.Outbound{
				Address:  vip.IP,
				Port:     uint32(port.Port),
				Resource: kri.WithSectionName(kri.From(d), port.GetName()),
			})
		}
	}
	return outbounds
}

func (d *Destination) Domains() *xds_types.VIPDomains {
	if len(d.Status.VIPs) > 0 {
		var domains []string
		for _, addr := range d.Status.Addresses {
			domains = append(domains, addr.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: d.Status.VIPs[0].IP,
			Domains: domains,
		}
	}
	return nil
}

func (d *Destination) GetPorts() []core.Port {
	var ports []core.Port
	for _, port := range d.Spec.Ports {
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

type DestinationList struct {
	MeshMultiZoneServiceResourceList
}

func ToDstList(mzList *MeshMultiZoneServiceResourceList) *DestinationList {
	return &DestinationList{*mzList}
}

func (l *DestinationList) GetDestinations() []core.Destination {
	var result []core.Destination
	for _, item := range l.Items {
		result = append(result, &Destination{*item})
	}
	return result
}
