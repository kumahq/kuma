package topology

import (
	"strings"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/dns/vips"
)

func VIPOutbounds(
	virtualOutboundView *vips.VirtualOutboundMeshView,
	tldomain string,
	vipPort uint32,
) ([]xds_types.VIPDomains, []*xds_types.Outbound) {
	var vipDomains []xds_types.VIPDomains
	var outbounds []*xds_types.Outbound
	for _, key := range virtualOutboundView.HostnameEntries() {
		voutbound := virtualOutboundView.Get(key)
		if voutbound.Address == "" {
			continue
		}
		domain := xds_types.VIPDomains{Address: voutbound.Address}
		switch key.Type {
		case vips.Host, vips.FullyQualifiedDomain:
			if govalidator.IsDNSName(key.Name) {
				domain.Domains = []string{key.Name}
				vipDomains = append(vipDomains, domain)
			}
			seenGlobalVip := false
			for _, ob := range voutbound.Outbounds {
				seenGlobalVip = seenGlobalVip || ob.Port == vipPort
				if ob.Port != 0 {
					outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
						Address: voutbound.Address,
						Port:    ob.Port,
						Tags:    ob.TagSet,
					}})
				}
			}
			// TODO remove the `vips.Host` on the next major version it's there for backward compatibility
			if key.Type == vips.Host && !seenGlobalVip && len(voutbound.Outbounds) > 0 && len(domain.Domains) > 0 {
				outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    vipPort,
					Tags:    voutbound.Outbounds[0].TagSet,
				}})
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
				outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    ob.Port,
					Tags:    ob.TagSet,
				}})
			}
			// TODO this should be a else once we remove backward compatibility
			if ob.Port != vipPort {
				outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    vipPort,
					Tags:    ob.TagSet,
				}})
			}
			vipDomains = append(vipDomains, domain)
		}
	}
	return vipDomains, outbounds
}

func Outbounds[T interface{ AsOutbounds() xds_types.Outbounds }](list []T) xds_types.Outbounds {
	var outbounds xds_types.Outbounds
	for _, item := range list {
		outbounds = append(outbounds, item.AsOutbounds()...)
	}
	return outbounds
}

func Domains[T interface{ Domains() *xds_types.VIPDomains }](list []T) []xds_types.VIPDomains {
	var domains []xds_types.VIPDomains
	for _, item := range list {
		if d := item.Domains(); d != nil {
			domains = append(domains, *d)
		}
	}
	return domains
}
