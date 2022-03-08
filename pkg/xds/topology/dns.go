package topology

import (
	"strings"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
)

const VIPListenPort = uint32(80)

func VIPOutbounds(
	virtualOutboundView *vips.VirtualOutboundMeshView,
	tldomain string,
) ([]xds.VIPDomains, []*mesh_proto.Dataplane_Networking_Outbound) {
	var vipDomains []xds.VIPDomains
	var outbounds []*mesh_proto.Dataplane_Networking_Outbound
	for _, key := range virtualOutboundView.HostnameEntries() {
		voutbound := virtualOutboundView.Get(key)
		if voutbound.Address == "" {
			continue
		}
		domain := xds.VIPDomains{Address: voutbound.Address}
		switch key.Type {
		case vips.Host, vips.FullyQualifiedDomain:
			if govalidator.IsDNSName(key.Name) {
				domain.Domains = []string{key.Name}
				vipDomains = append(vipDomains, domain)
			}
			seenGlobalVip := false
			for _, ob := range voutbound.Outbounds {
				seenGlobalVip = seenGlobalVip || ob.Port == VIPListenPort
				if ob.Port != 0 {
					outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
						Address: voutbound.Address,
						Port:    ob.Port,
						Tags:    ob.TagSet,
					})
				}
			}
			// TODO remove the `vips.Host` on the next major version it's there for backward compatibility
			if key.Type == vips.Host && !seenGlobalVip && len(voutbound.Outbounds) > 0 && len(domain.Domains) > 0 {
				outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    VIPListenPort,
					Tags:    voutbound.Outbounds[0].TagSet,
				})
			}
		case vips.Service:
			ob := voutbound.Outbounds[0]
			service := ob.TagSet[mesh_proto.ServiceTag]
			domain.Domains = []string{service + "." + tldomain}
			cleanedDomain := strings.ReplaceAll(service, "_", ".") + "." + tldomain
			if cleanedDomain != domain.Domains[0] {
				domain.Domains = append(domain.Domains, cleanedDomain)
			}
			if ob.Port != 0 {
				outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    ob.Port,
					Tags:    ob.TagSet,
				})
			}
			// TODO this should be a else once we remove backward compatibility
			if ob.Port != VIPListenPort {
				outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    VIPListenPort,
					Tags:    ob.TagSet,
				})
			}
			vipDomains = append(vipDomains, domain)
		}
	}
	return vipDomains, outbounds
}
