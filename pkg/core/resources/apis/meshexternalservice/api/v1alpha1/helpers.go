package v1alpha1

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v2/pkg/core/metadata"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/core/vip"
	xds_types "github.com/kumahq/kuma/v2/pkg/core/xds/types"
)

type Destination struct {
	MeshExternalServiceResource
}

func ToDst(mes *MeshExternalServiceResource) *Destination {
	return &Destination{*mes}
}

func IsReachableFromZone(mes *MeshExternalServiceResource, zone string) bool {
	return mes.GetMeta().GetLabels() == nil || mes.GetMeta().GetLabels()[mesh_proto.ZoneTag] == "" || mes.GetMeta().GetLabels()[mesh_proto.ZoneTag] == zone
}

var _ vip.ResourceHoldingVIPs = &Destination{}

func (d *Destination) VIPs() []string {
	if d.Status.VIP.IP == "" {
		return nil
	}
	return []string{d.Status.VIP.IP}
}

func (d *Destination) AllocateVIP(vip string) {
	d.Status.VIP = VIP{
		IP: vip,
	}
}

func (d *Destination) AsOutbounds() xds_types.Outbounds {
	if d.Status.VIP.IP != "" {
		return xds_types.Outbounds{{
			Address:  d.Status.VIP.IP,
			Port:     uint32(d.Spec.Match.Port),
			Resource: kri.WithSectionName(kri.From(d), d.Spec.Match.GetName()),
		}}
	}
	return nil
}

func (d *Destination) Domains() *xds_types.VIPDomains {
	if d.Status.VIP.IP != "" {
		var domains []string
		for _, address := range d.Status.Addresses {
			domains = append(domains, address.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: d.Status.VIP.IP,
			Domains: domains,
		}
	}
	return nil
}

func (d *Destination) GetPorts() []core.Port {
	return []core.Port{d.Spec.Match}
}

func (d *Destination) FindPortByName(name string) (core.Port, bool) {
	return d.Spec.Match, true
}

func (m Match) GetName() string {
	return fmt.Sprintf("%d", m.Port)
}

func (m Match) GetValue() int32 {
	return m.Port
}

func (m Match) GetProtocol() core_meta.Protocol {
	return m.Protocol
}

type DestinationList struct {
	MeshExternalServiceResourceList
}

func ToDstList(mesList *MeshExternalServiceResourceList) *DestinationList {
	return &DestinationList{*mesList}
}

func (l *DestinationList) GetDestinations() []core.Destination {
	var result []core.Destination
	for _, item := range l.Items {
		result = append(result, &Destination{*item})
	}
	return result
}
