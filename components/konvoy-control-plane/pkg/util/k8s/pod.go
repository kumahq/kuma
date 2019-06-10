package k8s

import (
	k8s_core "k8s.io/api/core/v1"
)

func GetTcpPorts(pod *k8s_core.Pod) []uint32 {
	ports := make([]uint32, 0, 1)
	for _, c := range pod.Spec.Containers {
		for _, p := range c.Ports {
			if p.Protocol == "" || p.Protocol == k8s_core.ProtocolTCP {
				ports = append(ports, uint32(p.ContainerPort))
			}
		}
	}
	return ports
}
