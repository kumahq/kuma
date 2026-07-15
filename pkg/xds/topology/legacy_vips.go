package topology

import (
	"encoding/json"
	"maps"
	"strings"

	"github.com/asaskevich/govalidator"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmzservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	system_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/system"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/dns/vips"
)

// LegacyVIPCompatibility restores legacy service VIPs from the persisted DNS
// config and remaps MeshService-family hostnames onto those VIPs when the
// dataplane intentionally has no Mesh*Service outbounds.
func LegacyVIPCompatibility(
	configs []*system_api.ConfigResource,
	domain string,
	serviceVIPPort uint32,
	meshServices []*meshservice_api.MeshServiceResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
) ([]xds_types.VIPDomains, xds_types.Outbounds, error) {
	serviceVIPs := map[string]string{}
	var domains []xds_types.VIPDomains
	var outbounds xds_types.Outbounds

	for _, config := range configs {
		view, err := virtualOutboundViewFromConfig(config.Spec.GetConfig())
		if err != nil {
			return nil, nil, err
		}

		viewDomains, viewOutbounds := VIPOutbounds(view, domain, serviceVIPPort)
		domains = append(domains, viewDomains...)
		outbounds = append(outbounds, viewOutbounds...)

		maps.Copy(serviceVIPs, serviceEntryAddresses(view))
	}

	domains = append(domains, resourceDomainsForLegacyVIPs(meshServices, meshExternalServices, meshMultiZoneServices, serviceVIPs)...)

	return domains, outbounds, nil
}

func resourceDomainsForLegacyVIPs(
	meshServices []*meshservice_api.MeshServiceResource,
	meshExternalServices []*meshexternalservice_api.MeshExternalServiceResource,
	meshMultiZoneServices []*meshmzservice_api.MeshMultiZoneServiceResource,
	serviceVIPs map[string]string,
) []xds_types.VIPDomains {
	var result []xds_types.VIPDomains

	for _, ms := range meshServices {
		if address := firstMeshServiceLegacyVIP(ms, serviceVIPs); address != "" {
			if domains := meshServiceStatusDomains(ms); len(domains) > 0 {
				result = append(result, xds_types.VIPDomains{Address: address, Domains: domains})
			}
		}

		for _, port := range ms.Spec.Ports {
			if address := serviceVIPs[legacyMeshServiceEntryName(ms, port.Port)]; address != "" {
				result = append(result, xds_types.VIPDomains{
					Address: address,
					Domains: legacyMeshServicePortDomains(ms, port.Port),
				})
			}
		}
	}

	for _, mes := range meshExternalServices {
		if address := serviceVIPs[legacyMeshExternalServiceEntryName(mes)]; address != "" {
			if domains := meshExternalServiceStatusDomains(mes); len(domains) > 0 {
				result = append(result, xds_types.VIPDomains{Address: address, Domains: domains})
			}
		}
	}

	for _, mzs := range meshMultiZoneServices {
		if address := serviceVIPs[legacyMeshMultiZoneServiceEntryName(mzs)]; address != "" {
			if domains := meshMultiZoneServiceStatusDomains(mzs); len(domains) > 0 {
				result = append(result, xds_types.VIPDomains{Address: address, Domains: domains})
			}
		}
	}

	return result
}

func serviceEntryAddresses(view *vips.VirtualOutboundMeshView) map[string]string {
	addresses := map[string]string{}
	for _, entry := range view.HostnameEntries() {
		if entry.Type != vips.Service {
			continue
		}
		if outbound := view.Get(entry); outbound != nil && outbound.Address != "" {
			addresses[entry.Name] = outbound.Address
		}
	}
	return addresses
}

func firstMeshServiceLegacyVIP(ms *meshservice_api.MeshServiceResource, serviceVIPs map[string]string) string {
	for _, port := range ms.Spec.Ports {
		if address := serviceVIPs[legacyMeshServiceEntryName(ms, port.Port)]; address != "" {
			return address
		}
	}
	return serviceVIPs[legacyMeshServiceEntryName(ms, 0)]
}

func legacyMeshServicePortDomains(ms *meshservice_api.MeshServiceResource, port int32) []string {
	serviceName := legacyMeshServiceEntryName(ms, port)
	if serviceName == "" {
		return nil
	}

	domains := []string{serviceName + "." + legacyMeshDomain}
	if cleanedDomain := strings.ReplaceAll(serviceName, "_", ".") + "." + legacyMeshDomain; cleanedDomain != domains[0] {
		domains = append(domains, cleanedDomain)
	}
	return domains
}

func meshServiceStatusDomains(ms *meshservice_api.MeshServiceResource) []string {
	var domains []string
	for _, address := range ms.Status.Addresses {
		if address.Hostname != "" {
			domains = append(domains, address.Hostname)
		}
	}
	return domains
}

func legacyMeshExternalServiceEntryName(mes *meshexternalservice_api.MeshExternalServiceResource) string {
	domains := legacyMeshExternalServiceDomains(mes)
	if len(domains) == 0 {
		return ""
	}
	return strings.TrimSuffix(domains[0], "."+legacyMeshDomain)
}

func meshExternalServiceStatusDomains(mes *meshexternalservice_api.MeshExternalServiceResource) []string {
	var domains []string
	for _, address := range mes.Status.Addresses {
		if address.Hostname != "" {
			domains = append(domains, address.Hostname)
		}
	}
	return domains
}

func legacyMeshMultiZoneServiceEntryName(mzs *meshmzservice_api.MeshMultiZoneServiceResource) string {
	domains := legacyMeshMultiZoneServiceDomains(mzs)
	if len(domains) == 0 {
		return ""
	}
	return strings.TrimSuffix(domains[0], "."+legacyMeshDomain)
}

func meshMultiZoneServiceStatusDomains(mzs *meshmzservice_api.MeshMultiZoneServiceResource) []string {
	var domains []string
	for _, address := range mzs.Status.Addresses {
		if address.Hostname != "" {
			domains = append(domains, address.Hostname)
		}
	}
	return domains
}

func virtualOutboundViewFromConfig(config string) (*vips.VirtualOutboundMeshView, error) {
	if config == "" {
		return vips.NewEmptyVirtualOutboundView(), nil
	}

	tagFirstView := vips.NewEmptyTagFirstOutboundView()
	if err := json.Unmarshal([]byte(config), tagFirstView); err != nil {
		return nil, err
	}
	if len(tagFirstView.PerService) != 0 {
		return tagFirstView.ToVirtualOutboundView(), nil
	}

	view := map[vips.HostnameEntry]vips.VirtualOutbound{}
	if err := json.Unmarshal([]byte(config), &view); err != nil {
		return nil, err
	}
	return vips.NewVirtualOutboundView(view)
}

func isLegacyVirtualOutboundEntry(origin string) bool {
	return strings.HasPrefix(origin, vips.VirtualOutboundPrefix)
}

func VIPOutbounds(
	virtualOutboundView *vips.VirtualOutboundMeshView,
	topLevelDomain string,
	vipPort uint32,
) ([]xds_types.VIPDomains, xds_types.Outbounds) {
	var vipDomains []xds_types.VIPDomains
	var outbounds xds_types.Outbounds
	for _, key := range virtualOutboundView.HostnameEntries() {
		voutbound := virtualOutboundView.Get(key)
		if voutbound == nil || voutbound.Address == "" {
			continue
		}

		switch key.Type {
		case vips.Host, vips.FullyQualifiedDomain:
			var kept []vips.OutboundEntry
			for _, ob := range voutbound.Outbounds {
				if !isLegacyVirtualOutboundEntry(ob.Origin) {
					kept = append(kept, ob)
				}
			}
			if len(kept) == 0 {
				continue
			}

			domain := xds_types.VIPDomains{Address: voutbound.Address}
			if govalidator.IsDNSName(key.Name) {
				domain.Domains = []string{key.Name}
				vipDomains = append(vipDomains, domain)
			}
			seenGlobalVip := false
			for _, ob := range kept {
				seenGlobalVip = seenGlobalVip || ob.Port == vipPort
				if ob.Port != 0 {
					outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
						Address: voutbound.Address,
						Port:    ob.Port,
						Tags:    ob.TagSet,
					}})
				}
			}
			if key.Type == vips.Host && !seenGlobalVip && len(domain.Domains) > 0 {
				outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    vipPort,
					Tags:    kept[0].TagSet,
				}})
			}
		case vips.Service:
			if len(voutbound.Outbounds) == 0 {
				continue
			}

			domain := xds_types.VIPDomains{Address: voutbound.Address}

			outbound := voutbound.Outbounds[0]
			service := outbound.TagSet[mesh_proto.ServiceTag]
			if service == "" {
				service = key.Name
			}

			domain.Domains = []string{service + "." + topLevelDomain}
			cleanedDomain := strings.ReplaceAll(service, "_", ".") + "." + topLevelDomain
			if cleanedDomain != domain.Domains[0] {
				domain.Domains = append(domain.Domains, cleanedDomain)
			}
			if outbound.Port != 0 {
				outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    outbound.Port,
					Tags:    outbound.TagSet,
				}})
			}
			if outbound.Port != vipPort {
				outbounds = append(outbounds, &xds_types.Outbound{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{
					Address: voutbound.Address,
					Port:    vipPort,
					Tags:    outbound.TagSet,
				}})
			}
			vipDomains = append(vipDomains, domain)
		}
	}
	return vipDomains, outbounds
}
