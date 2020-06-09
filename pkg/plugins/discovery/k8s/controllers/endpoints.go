package controllers

import (
	"sort"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_k8s "github.com/Kong/kuma/pkg/plugins/resources/k8s/native/api/v1alpha1"
	util_proto "github.com/Kong/kuma/pkg/util/proto"
)

type Endpoint struct {
	Address string
	Port    uint32
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

func endpointsByService(dataplanes []*mesh_k8s.Dataplane) EndpointsByService {
	result := EndpointsByService{}
	for _, other := range dataplanes {
		dataplane := &mesh_proto.Dataplane{}
		if err := util_proto.FromMap(other.Spec, dataplane); err != nil {
			converterLog.Error(err, "failed to parse Dataplane", "dataplane", other.Spec)
			continue // one invalid Dataplane definition should not break the entire mesh
		}
		if dataplane.Networking.Ingress != nil {
			continue
		}
		for _, inbound := range dataplane.Networking.GetInbound() {
			svc, ok := inbound.GetTags()[mesh_proto.ServiceTag]
			if !ok {
				continue
			}
			endpoint := Endpoint{
				Port: inbound.Port,
			}
			if inbound.Address != "" {
				endpoint.Address = inbound.Address
			} else {
				endpoint.Address = dataplane.Networking.Address
			}
			result[svc] = append(result[svc], endpoint)
		}
	}
	return result
}
