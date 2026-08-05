package v1alpha1

import (
	"fmt"

	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core/vip"
	"github.com/kumahq/kuma/v3/pkg/core/resources/sni"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
)

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

func (s *MeshExternalService) SNIs() []sni.Section {
	if s == nil {
		return nil
	}
	return []sni.Section{{Port: s.Match.Port, SectionName: s.Match.GetName()}}
}

func (l *MeshExternalServiceResourceList) GetDestinations() []core.Destination {
	var result []core.Destination
	for _, item := range l.Items {
		result = append(result, item)
	}
	return result
}
