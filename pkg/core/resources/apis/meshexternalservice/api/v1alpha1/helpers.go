package v1alpha1

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq
	"github.com/kumahq/kuma/pkg/core/resources/apis/core/vip"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (m *MeshExternalServiceResource) DestinationName(port uint32) string {
	id := kri.From(m, "")
	if port == 0 {
		port = uint32(m.Spec.Match.Port)
	}
	return fmt.Sprintf("%s_%s_%s_%s_%s_%d", id.Mesh, id.Name, id.Namespace, id.Zone, MeshExternalServiceResourceTypeDescriptor.ShortName, port)
}

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
			Resource: pointer.To(kri.From(t, "")),
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
