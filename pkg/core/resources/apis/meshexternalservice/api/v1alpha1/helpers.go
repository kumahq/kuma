package v1alpha1

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/pkg/core/metadata"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
)

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

func (t *MeshExternalServiceResource) AsOutbounds() xds_types.Outbounds {
	if t.Status.VIP.IP != "" {
		return xds_types.Outbounds{{
			Address:  t.Status.VIP.IP,
			Port:     uint32(t.Spec.Match.Port),
			Resource: kri.WithSectionName(kri.From(t), t.Spec.Match.GetName()),
		}}
	}
	return nil
}

func (t *MeshExternalServiceResource) Domains() *xds_types.VIPDomains {
	if t.Status.VIP.IP != "" {
		var domains []string
		for _, address := range t.Status.Addresses {
			domains = append(domains, address.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: t.Status.VIP.IP,
			Domains: domains,
		}
	}
	return nil
}

func (t *MeshExternalServiceResource) GetPorts() []core.Port {
	return []core.Port{t.Spec.Match}
}

func (t *MeshExternalServiceResource) FindPortByName(name string) (core.Port, bool) {
	return t.Spec.Match, true
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

func (l *MeshExternalServiceResourceList) GetDestinations() []core.Destination {
	var result []core.Destination
	for _, item := range l.Items {
		result = append(result, item)
	}
	return result
}
