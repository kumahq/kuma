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

func (r *MeshExternalServiceResource) IsReachableFromZone(zone string) bool {
	return r.GetMeta().GetLabels() == nil || r.GetMeta().GetLabels()[mesh_proto.ZoneTag] == "" || r.GetMeta().GetLabels()[mesh_proto.ZoneTag] == zone
}

var _ vip.ResourceHoldingVIPs = &MeshExternalServiceResource{}

func (r *MeshExternalServiceResource) VIPs() []string {
	if r.Status.VIP.IP == "" {
		return nil
	}
	return []string{r.Status.VIP.IP}
}

func (r *MeshExternalServiceResource) AllocateVIP(vip string) {
	r.Status.VIP = VIP{
		IP: vip,
	}
}

func (r *MeshExternalServiceResource) AsOutbounds() xds_types.Outbounds {
	if r.Status.VIP.IP != "" {
		return xds_types.Outbounds{{
			Address:  r.Status.VIP.IP,
			Port:     uint32(r.Spec.Match.Port),
			Resource: kri.WithSectionName(kri.From(r), r.Spec.Match.GetName()),
		}}
	}
	return nil
}

func (r *MeshExternalServiceResource) Domains() *xds_types.VIPDomains {
	if r.Status.VIP.IP != "" {
		var domains []string
		for _, address := range r.Status.Addresses {
			domains = append(domains, address.Hostname)
		}
		return &xds_types.VIPDomains{
			Address: r.Status.VIP.IP,
			Domains: domains,
		}
	}
	return nil
}

func (r *MeshExternalServiceResource) GetPorts() []core.Port {
	return []core.Port{r.Spec.Match}
}

func (r *MeshExternalServiceResource) FindPortByName(name string) (core.Port, bool) {
	return r.Spec.Match, true
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
