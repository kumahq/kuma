package dns

import (
	"sort"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/dns/vips"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

const VIPListenPort = uint32(80)

func VIPOutbounds(
	name string,
	dataplanes []*core_mesh.DataplaneResource,
	vips vips.List,
	externalServices []*core_mesh.ExternalServiceResource,
) []*mesh_proto.Dataplane_Networking_Outbound {
	serviceVIPMap := map[string]string{}
	services := []string{}
	for _, dataplane := range dataplanes {
		if dataplane.Meta.GetName() == name {
			continue
		}

		if dataplane.Spec.IsIngress() {
			for _, service := range dataplane.Spec.Networking.Ingress.AvailableServices {
				if service.Mesh == dataplane.Meta.GetMesh() {
					// Only add outbounds for services in the same mesh
					inService := service.Tags[mesh_proto.ServiceTag]
					if _, found := serviceVIPMap[inService]; !found {
						vip, err := ForwardLookup(vips, inService)
						if err == nil {
							serviceVIPMap[inService] = vip
							services = append(services, inService)
						}
					}
				}
			}
		} else {
			for _, inbound := range dataplane.Spec.Networking.Inbound {
				inService := inbound.GetTags()[mesh_proto.ServiceTag]
				if _, found := serviceVIPMap[inService]; !found {
					vip, err := ForwardLookup(vips, inService)
					if err == nil {
						serviceVIPMap[inService] = vip
						services = append(services, inService)
					}
				}
			}
		}
	}

	for _, externalService := range externalServices {
		inService := externalService.Spec.Tags[mesh_proto.ServiceTag]
		if _, found := serviceVIPMap[inService]; !found {
			vip, err := ForwardLookup(vips, inService)
			if err == nil {
				serviceVIPMap[inService] = vip
				services = append(services, inService)
			}
		}
	}

	sort.Strings(services)
	outbounds := []*mesh_proto.Dataplane_Networking_Outbound{}
	for _, service := range services {
		outbounds = append(outbounds,
			&mesh_proto.Dataplane_Networking_Outbound{
				Address: serviceVIPMap[service],
				Port:    VIPListenPort,
				Tags: map[string]string{
					mesh_proto.ServiceTag: service,
				},
			})
	}

	return outbounds
}

func ForwardLookup(vips vips.List, service string) (string, error) {
	ip, found := vips[service]
	if !found {
		return "", errors.Errorf("service [%s] not found", service)
	}
	return ip, nil
}
