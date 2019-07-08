package controllers

import (
	"fmt"

	discovery_proto "github.com/Kong/konvoy/components/konvoy-control-plane/api/discovery/v1alpha1"
	core_discovery "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/discovery"
	kube_core "k8s.io/api/core/v1"
)

func ToWorkload(pod *kube_core.Pod) (*core_discovery.WorkloadInfo, error) {
	return &core_discovery.WorkloadInfo{
		Workload: &discovery_proto.Workload{
			Id: &discovery_proto.Id{
				Namespace: pod.ObjectMeta.Namespace,
				Name:      pod.ObjectMeta.Name,
			},
			Meta: &discovery_proto.Meta{
				Labels: pod.ObjectMeta.Labels,
			},
		},
		Desc: &core_discovery.WorkloadDescription{
			Version:   fmt.Sprintf("v%d", pod.Generation),
			Endpoints: GetEndpoints(pod),
		},
	}, nil
}

func GetEndpoints(pod *kube_core.Pod) []core_discovery.WorkloadEndpoint {
	address := pod.Status.PodIP
	ports := GetTcpPorts(pod)
	endpoints := make([]core_discovery.WorkloadEndpoint, 0, len(ports))
	for _, port := range ports {
		endpoints = append(endpoints, core_discovery.WorkloadEndpoint{
			Address: address,
			Port:    port,
		})
	}
	return endpoints
}

func GetTcpPorts(pod *kube_core.Pod) []uint32 {
	ports := make([]uint32, 0, 1)
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.Protocol == "" || p.Protocol == kube_core.ProtocolTCP {
				ports = append(ports, uint32(p.ContainerPort))
			}
		}
	}
	return ports
}
