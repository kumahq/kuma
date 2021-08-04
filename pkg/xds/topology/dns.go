package topology

import (
	"sort"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
)

const VIPListenPort = uint32(80)

func VIPOutbounds(
	resourceKey model.ResourceKey,
	dataplanes []*core_mesh.DataplaneResource,
	zoneIngresses []*core_mesh.ZoneIngressResource,
	vipList vips.List,
	tldomain string,
	externalServices []*core_mesh.ExternalServiceResource,
) ([]xds.VIPDomains, []*mesh_proto.Dataplane_Networking_Outbound) {
	type vipEntry struct {
		ip        string
		port      uint32
		entryType vips.EntryType
	}
	serviceVIPMap := map[string][]vipEntry{}
	services := []string{}
	for _, dataplane := range dataplanes {
		// backwards compatibility
		if dataplane.Spec.IsIngress() {
			for _, service := range dataplane.Spec.GetNetworking().GetIngress().GetAvailableServices() {
				if service.Mesh == resourceKey.Mesh {
					// Only add outbounds for services in the same mesh
					inService := service.Tags[mesh_proto.ServiceTag]
					if _, found := serviceVIPMap[inService]; !found {
						vip, err := ForwardLookup(vipList, vips.NewServiceEntry(inService))
						if err == nil {
							serviceVIPMap[inService] = append(serviceVIPMap[inService], vipEntry{vip, VIPListenPort, vips.Service})
							services = append(services, inService)
						}
					}
				}
			}
		} else {
			for _, inbound := range dataplane.Spec.GetNetworking().GetInbound() {
				inService := inbound.GetTags()[mesh_proto.ServiceTag]
				if _, found := serviceVIPMap[inService]; !found {
					vip, err := ForwardLookup(vipList, vips.NewServiceEntry(inService))
					if err == nil {
						serviceVIPMap[inService] = append(serviceVIPMap[inService], vipEntry{vip, VIPListenPort, vips.Service})
						services = append(services, inService)
					}
				}
			}
		}
	}

	for _, zi := range zoneIngresses {
		for _, service := range zi.Spec.GetAvailableServices() {
			if service.Mesh == resourceKey.Mesh {
				// Only add outbounds for services in the same mesh
				inService := service.Tags[mesh_proto.ServiceTag]
				if _, found := serviceVIPMap[inService]; !found {
					vip, err := ForwardLookup(vipList, vips.NewServiceEntry(inService))
					if err == nil {
						serviceVIPMap[inService] = append(serviceVIPMap[inService], vipEntry{vip, VIPListenPort, vips.Service})
						services = append(services, inService)
					}
				}
			}
		}
	}

	externalServicesByServiceName := map[string]*core_mesh.ExternalServiceResource{}
	for _, externalService := range externalServices {
		inService := externalService.Spec.Tags[mesh_proto.ServiceTag]
		externalServicesByServiceName[inService] = externalService
		host := externalService.Spec.GetHost()
		if _, found := serviceVIPMap[inService]; !found {
			vip1, err := ForwardLookup(vipList, vips.NewHostEntry(host))
			if err == nil {
				port := externalService.Spec.GetPort()
				var p32 uint32
				if p64, err := strconv.ParseUint(port, 10, 32); err != nil {
					p32 = VIPListenPort
				} else {
					p32 = uint32(p64)
				}
				serviceVIPMap[inService] = append(serviceVIPMap[inService], vipEntry{vip1, p32, vips.Host})
				services = append(services, inService)
			}
			vip2, err := ForwardLookup(vipList, vips.NewServiceEntry(inService))
			if err == nil {
				port := externalService.Spec.GetPort()
				var p32 uint32
				if p64, err := strconv.ParseUint(port, 10, 32); err != nil {
					p32 = VIPListenPort
				} else {
					p32 = uint32(p64)
				}
				serviceVIPMap[inService] = append(serviceVIPMap[inService], vipEntry{vip2, p32, vips.Service})
				services = append(services, inService)
			}
		}
	}

	sort.Strings(services)
	var vipDomains []xds.VIPDomains
	var outbounds []*mesh_proto.Dataplane_Networking_Outbound
	for _, service := range services {
		entries := serviceVIPMap[service]
		for _, entry := range entries {
			outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
				Address: entry.ip,
				Tags:    map[string]string{mesh_proto.ServiceTag: service},
				Port:    entry.port,
			})
			vip := xds.VIPDomains{
				Address: entry.ip,
			}
			switch entry.entryType {
			case vips.Service:
				// add regular .mesh domain
				vip.Domains = []string{service + "." + tldomain}
				cleanedDomain := strings.ReplaceAll(service, "_", ".") + "." + tldomain
				if cleanedDomain != vip.Domains[0] {
					vip.Domains = append(vip.Domains, cleanedDomain)
				}
				// todo (lobkovilya): backwards compatibility, could be deleted in the next major release Kuma 1.2.x
				if entry.port != VIPListenPort {
					outbounds = append(outbounds, &mesh_proto.Dataplane_Networking_Outbound{
						Address: entry.ip,
						Tags:    map[string]string{mesh_proto.ServiceTag: service},
						Port:    VIPListenPort,
					})
				}
			case vips.Host:
				host := externalServicesByServiceName[service].Spec.GetHost()
				if govalidator.IsDNSName(host) {
					vip.Domains = append(vip.Domains, host)
				}
			}
			vipDomains = append(vipDomains, vip)
		}
	}

	return vipDomains, outbounds
}

func ForwardLookup(vips vips.List, entry vips.Entry) (string, error) {
	ip, found := vips[entry]
	if !found {
		return "", errors.Errorf("entry name [%s] not found", entry.Name)
	}
	return ip, nil
}
