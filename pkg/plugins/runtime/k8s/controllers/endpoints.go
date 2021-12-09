package controllers

import (
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
)

type Endpoint struct {
	Address  string
	Port     uint32
	Instance string
}

type EndpointsByService map[string][]Endpoint

func (e EndpointsByService) Services() []string {
	list := make([]string, 0, len(e))
	for key := range e {
		list = append(list, key)
	}
	sort.Strings(list)
	return list
}

func endpointsByService(dataplanes []*core_mesh.DataplaneResource) EndpointsByService {
	result := EndpointsByService{}
	for _, other := range dataplanes {
		for _, inbound := range other.Spec.Networking.GetInbound() {
			svc, ok := inbound.GetTags()[mesh_proto.ServiceTag]
			if !ok {
				continue
			}
			endpoint := Endpoint{
				Port:     inbound.Port,
				Instance: inbound.GetTags()[mesh_proto.InstanceTag],
			}
			if inbound.Address != "" {
				endpoint.Address = inbound.Address
			} else {
				endpoint.Address = other.Spec.Networking.Address
			}
			result[svc] = append(result[svc], endpoint)
		}
	}
	return result
}
