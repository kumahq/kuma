package controllers

import (
	discovery_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/discovery/v1alpha1"
	mesh_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/mesh/v1alpha1"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	mesh_k8s "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/k8s/native/api/v1alpha1"

	util_proto "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/proto"
)

func DataplaneToWorkload(dataplane *mesh_k8s.Dataplane) (*core_discovery.WorkloadInfo, error) {
	endpoints, err := EndpointsFor(dataplane)
	if err != nil {
		return nil, err
	}
	return &core_discovery.WorkloadInfo{
		Workload: &discovery_proto.Workload{
			Id: &discovery_proto.Id{
				Namespace: dataplane.ObjectMeta.Namespace,
				Name:      dataplane.ObjectMeta.Name,
			},
			Meta: &discovery_proto.Meta{
				Labels: dataplane.ObjectMeta.Labels,
			},
		},
		Desc: &core_discovery.WorkloadDescription{
			Version:   dataplane.ObjectMeta.ResourceVersion,
			Endpoints: endpoints,
		},
	}, nil
}

func EndpointsFor(dataplane *mesh_k8s.Dataplane) ([]core_discovery.WorkloadEndpoint, error) {
	proto := &mesh_proto.Dataplane{}
	if err := util_proto.FromMap(dataplane.Spec, proto); err != nil {
		return nil, err
	}
	endpoints := make([]core_discovery.WorkloadEndpoint, 0, len(proto.Networking.GetInbound()))
	for _, inbound := range proto.Networking.GetInbound() {
		iface, err := mesh_proto.ParseInboundInterface(inbound.Interface)
		if err != nil {
			return nil, err
		}
		endpoints = append(endpoints, core_discovery.WorkloadEndpoint{
			Address: iface.WorkloadAddress,
			Port:    iface.WorkloadPort,
		})
	}
	return endpoints, nil
}
