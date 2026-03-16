package controllers

import (
	"fmt"

	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/util"
)

const (
	// KumaZoneProxyTypeIngress marks a Service as generating a ZoneIngress
	// listener on the Dataplane.
	KumaZoneProxyTypeIngress = "ingress"

	// KumaZoneProxyTypeEgress marks a Service as generating a ZoneEgress
	// listener on the Dataplane.
	KumaZoneProxyTypeEgress = "egress"
)

// ListenersForService generates Dataplane zone proxy listeners from a Service
// labeled with k8s.kuma.io/zone-proxy-type. The logic mirrors inboundForService
// but produces Listener objects instead of Inbound objects.
// Returns nil, nil when the service has no zone-proxy-type label.
func ListenersForService(pod *kube_core.Pod, svc *kube_core.Service) ([]*mesh_proto.Dataplane_Networking_Listener, error) {
	proxyTypeStr, ok := svc.Labels[metadata.KumaZoneProxyTypeLabel]
	if !ok {
		return nil, nil
	}

	var listenerType mesh_proto.Dataplane_Networking_Listener_Type
	switch proxyTypeStr {
	case KumaZoneProxyTypeIngress:
		listenerType = mesh_proto.Dataplane_Networking_Listener_ZoneIngress
	case KumaZoneProxyTypeEgress:
		listenerType = mesh_proto.Dataplane_Networking_Listener_ZoneEgress
	default:
		return nil, fmt.Errorf("service %s/%s has invalid %s label value %q; allowed: %q, %q",
			svc.Namespace, svc.Name, metadata.KumaZoneProxyTypeLabel, proxyTypeStr,
			KumaZoneProxyTypeIngress, KumaZoneProxyTypeEgress)
	}

	var listeners []*mesh_proto.Dataplane_Networking_Listener
	for idx := range svc.Spec.Ports {
		svcPort := svc.Spec.Ports[idx]
		if svcPort.Protocol != "" && svcPort.Protocol != kube_core.ProtocolTCP {
			continue
		}

		containerPort, _, container, err := util_k8s.FindPort(pod, &svcPort)
		if err != nil {
			converterLog.Error(err, "failed to find container port for zone proxy listener",
				"namespace", pod.Namespace, "podName", pod.Name,
				"serviceName", svc.Name, "servicePortName", svcPort.Name)
			continue
		}

		state := mesh_proto.Dataplane_Networking_Listener_Ready

		if container != nil {
			if cs := util_k8s.FindContainerStatus(container.Name, pod.Status.ContainerStatuses); cs != nil && !cs.Ready {
				state = mesh_proto.Dataplane_Networking_Listener_NotReady
			}
		}

		// also we're checking whether kuma-sidecar container is ready
		if cs := util_k8s.FindContainerOrInitContainerStatus(
			util_k8s.KumaSidecarContainerName,
			pod.Status.ContainerStatuses,
			pod.Status.InitContainerStatuses,
		); cs != nil && !cs.Ready {
			state = mesh_proto.Dataplane_Networking_Listener_NotReady
		}

		if pod.DeletionTimestamp != nil {
			state = mesh_proto.Dataplane_Networking_Listener_NotReady
		}

		name := svcPort.Name
		if name == "" {
			name = fmt.Sprintf("%s-%d", svc.Name, svcPort.Port)
		}

		listeners = append(listeners, &mesh_proto.Dataplane_Networking_Listener{
			Type:    listenerType,
			Address: pod.Status.PodIP,
			Port:    uint32(containerPort),
			Name:    name,
			State:   state,
		})
	}

	return listeners, nil
}
