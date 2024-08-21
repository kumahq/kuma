package topology

import (
	"strings"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
)

func VIPOutbounds(
	virtualOutboundView *vips.VirtualOutboundMeshView,
	tldomain string,
	vipPort uint32,
) ([]xds.VIPDomains, []*xds.Outbound) {
	var vipDomains []xds.VIPDomains
	var outbounds []*xds.Outbound
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
				seenGlobalVip = seenGlobalVip || ob.Port == vipPort
				if ob.Port != 0 {
					outbounds = append(outbounds, &xds.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
						Address: voutbound.Address,
						Port:    ob.Port,
						Tags:    ob.TagSet,
					}})
				}
			}
			// TODO remove the `vips.Host` on the next major version it's there for backward compatibility
			if key.Type == vips.Host && !seenGlobalVip && len(voutbound.Outbounds) > 0 && len(domain.Domains) > 0 {
				outbounds = append(outbounds, &xds.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
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
				outbounds = append(outbounds, &xds.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    ob.Port,
					Tags:    ob.TagSet,
				}})
			}
			// TODO this should be a else once we remove backward compatibility
			if ob.Port != vipPort {
				outbounds = append(outbounds, &xds.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
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

func MeshServiceOutbounds(meshServices []*meshservice_api.MeshServiceResource) ([]xds.VIPDomains, []*xds.Outbound) {
	var outbounds []*xds.Outbound
	var vipDomains []xds.VIPDomains
	for _, svc := range meshServices {
		for _, vip := range svc.Status.VIPs {
			for _, port := range svc.Spec.Ports {
				outbounds = append(outbounds, &xds.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: vip.IP,
					Port:    port.Port,
					BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
						Kind: string(meshservice_api.MeshServiceType),
						Name: svc.Meta.GetName(),
						Port: port.Port,
					},
				}})
			}
		}
		if len(svc.Status.VIPs) > 0 {
			var domains []string
			for _, addr := range svc.Status.Addresses {
				domains = append(domains, addr.Hostname)
			}
			vipDomains = append(vipDomains, xds.VIPDomains{
				Address: svc.Status.VIPs[0].IP,
				Domains: domains,
			})
		}
	}
	return vipDomains, outbounds
}

func MeshExternalServiceOutbounds(meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource) ([]xds.VIPDomains, []*xds.Outbound) {
	var vipDomains []xds.VIPDomains
	var outbounds []*xds.Outbound

	for _, meshExternalService := range meshExternalServices {
		if meshExternalService.Status.VIP.IP != "" {
			outbound := &mesh_proto.Dataplane_Networking_Outbound{
				Address: meshExternalService.Status.VIP.IP,
				Port:    uint32(meshExternalService.Spec.Match.Port),
				BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
					Kind: string(meshexternalservice_api.MeshExternalServiceType),
					Name: meshExternalService.Meta.GetName(),
					Port: uint32(meshExternalService.Spec.Match.Port),
				},
			}
			outbounds = append(outbounds, &xds.Outbound{LegacyOutbound: outbound})

			var domains []string
			for _, address := range meshExternalService.Status.Addresses {
				domains = append(domains, address.Hostname)
			}
			vipDomains = append(vipDomains, xds.VIPDomains{Address: meshExternalService.Status.VIP.IP, Domains: domains})
		}
	}

	return vipDomains, outbounds
}

func MeshMultiZoneServiceOutbounds(services []*v1alpha1.MeshMultiZoneServiceResource) ([]xds.VIPDomains, []*xds.Outbound) {
	var vipDomains []xds.VIPDomains
	var outbounds []*xds.Outbound

	for _, svc := range services {
		for _, vip := range svc.Status.VIPs {
			for _, port := range svc.Spec.Ports {
				outbound := &mesh_proto.Dataplane_Networking_Outbound{
					Address: vip.IP,
					Port:    port.Port,
					BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
						Kind: string(v1alpha1.MeshMultiZoneServiceType),
						Name: svc.Meta.GetName(),
						Port: port.Port,
					},
				}
				outbounds = append(outbounds, &xds.Outbound{LegacyOutbound: outbound})
			}
		}
		if len(svc.Status.VIPs) > 0 {
			var domains []string
			for _, addr := range svc.Status.Addresses {
				domains = append(domains, addr.Hostname)
			}
			vipDomains = append(vipDomains, xds.VIPDomains{
				Address: svc.Status.VIPs[0].IP,
				Domains: domains,
			})
		}
	}

	return vipDomains, outbounds
}
