package topology

import (
	"sort"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/dns"
)

const VIPListenPort = uint32(80)

func PatchDataplaneWithVIPOutbounds(dataplane *mesh_core.DataplaneResource,
	dataplanes *mesh_core.DataplaneResourceList, resolver dns.DNSResolver) (errs error) {
	serviceVIPMap := map[string]string{}
	services := []string{}
	for _, dp := range dataplanes.Items {
		if dp.Meta.GetName() == dataplane.Meta.GetName() {
			continue
		}

		if dp.Spec.IsIngress() {
			for _, service := range dp.Spec.Networking.Ingress.AvailableServices {
				inService := service.Tags[mesh_proto.ServiceTag]
				if _, found := serviceVIPMap[inService]; !found {
					vip, err := resolver.ForwardLookup(inService)
					if err == nil {
						serviceVIPMap[inService] = vip
						services = append(services, inService)
					}
				}
			}
		} else {
			for _, inbound := range dp.Spec.Networking.Inbound {
				inService := inbound.GetTags()[mesh_proto.ServiceTag]
				if _, found := serviceVIPMap[inService]; !found {
					vip, err := resolver.ForwardLookup(inService)
					if err == nil {
						serviceVIPMap[inService] = vip
						services = append(services, inService)
					}
				}
			}
		}
	}

	sort.Strings(services)

	for _, service := range services {
		dataplane.Spec.Networking.Outbound = append(dataplane.Spec.Networking.Outbound,
			&mesh_proto.Dataplane_Networking_Outbound{
				Address: serviceVIPMap[service],
				Port:    VIPListenPort,
				Service: service,
				Tags: map[string]string{
					mesh_proto.ServiceTag: service,
				},
			})
	}

	return
}
