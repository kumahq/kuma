package controllers

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func InboundInterfacesFor(zone string, pod *kube_core.Pod, services []*kube_core.Service) ([]*mesh_proto.Dataplane_Networking_Inbound, error) {
	var ifaces []*mesh_proto.Dataplane_Networking_Inbound
	for _, svc := range services {
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Protocol != "" && svcPort.Protocol != kube_core.ProtocolTCP {
				// ignore non-TCP ports
				continue
			}
			containerPort, container, err := util_k8s.FindPort(pod, &svcPort)
			if err != nil {
				converterLog.Error(err, "failed to find a container port in a given Pod that would match a given Service port", "namespace", pod.Namespace, "podName", pod.Name, "serviceName", svc.Name, "servicePortName", svcPort.Name)
				// ignore those cases where a Pod doesn't have all the ports a Service has
				continue
			}

			tags := InboundTagsForService(zone, pod, svc, &svcPort)
			var health *mesh_proto.Dataplane_Networking_Inbound_Health

			// if container is not equal nil then port is explicitly defined as containerPort so we're able
			// to figure out which container implements which service. Since we know container we can check its status
			// and map it to the Dataplane health
			if container != nil {
				if cs := util_k8s.FindContainerStatus(pod, container.Name); cs != nil {
					health = &mesh_proto.Dataplane_Networking_Inbound_Health{
						Ready: cs.Ready,
					}
				}
			}

			// also we're checking whether kuma-sidecar container is ready
			if cs := util_k8s.FindContainerStatus(pod, util_k8s.KumaSidecarContainerName); cs != nil {
				if health != nil {
					health.Ready = health.Ready && cs.Ready
				} else {
					health = &mesh_proto.Dataplane_Networking_Inbound_Health{
						Ready: cs.Ready,
					}
				}
			}

			ifaces = append(ifaces, &mesh_proto.Dataplane_Networking_Inbound{
				Port:   uint32(containerPort),
				Tags:   tags,
				Health: health,
			})
		}
	}

	if len(ifaces) == 0 {
		if len(services) > 0 {
			// Notice that here we return an error immediately
			// instead of leaving Dataplane validation up to a ValidatingAdmissionWebHook.
			// We do it this way in order to provide the most descriptive error message.
			return nil, errors.Errorf("Kuma requires every Pod in a Mesh to be a part of at least one Service. However, this Pod doesn't have any container ports that would satisfy matching Service(s).")
		}

		// The Pod does not have any services associated with it, just get the data from the Pod itself
		tags := InboundTagsForPod(zone, pod)
		var health *mesh_proto.Dataplane_Networking_Inbound_Health

		for _, container := range pod.Spec.Containers {
			if container.Name != util_k8s.KumaSidecarContainerName {
				if cs := util_k8s.FindContainerStatus(pod, container.Name); cs != nil {
					health = &mesh_proto.Dataplane_Networking_Inbound_Health{
						Ready: cs.Ready,
					}
				}
			}
		}

		// also we're checking whether kuma-sidecar container is ready
		if cs := util_k8s.FindContainerStatus(pod, util_k8s.KumaSidecarContainerName); cs != nil {
			if health != nil {
				health.Ready = health.Ready && cs.Ready
			} else {
				health = &mesh_proto.Dataplane_Networking_Inbound_Health{
					Ready: cs.Ready,
				}
			}
		}

		ifaces = append(ifaces, &mesh_proto.Dataplane_Networking_Inbound{
			Port:   mesh_core.TCPPortReserved,
			Tags:   tags,
			Health: health,
		})
	}
	return ifaces, nil
}

func InboundTagsForService(zone string, pod *kube_core.Pod, svc *kube_core.Service, svcPort *kube_core.ServicePort) map[string]string {
	tags := util_k8s.CopyStringMap(pod.Labels)
	for key, value := range tags {
		if value == "" {
			delete(tags, key)
		}
	}
	if tags == nil {
		tags = make(map[string]string)
	}
	tags[mesh_proto.ServiceTag] = ServiceTagFor(svc, svcPort)
	if zone != "" {
		tags[mesh_proto.ZoneTag] = zone
	}
	tags[mesh_proto.ProtocolTag] = ProtocolTagFor(svc, svcPort)
	if isHeadlessService(svc) {
		tags[mesh_proto.InstanceTag] = pod.Name
	}
	return tags
}

func ServiceTagFor(svc *kube_core.Service, svcPort *kube_core.ServicePort) string {
	return fmt.Sprintf("%s_%s_svc_%d", svc.Name, svc.Namespace, svcPort.Port)
}

// ProtocolTagFor infers service protocol from a `<port>.service.kuma.io/protocol` annotation.
func ProtocolTagFor(svc *kube_core.Service, svcPort *kube_core.ServicePort) string {
	protocolAnnotation := fmt.Sprintf("%d.service.kuma.io/protocol", svcPort.Port)
	protocolValue := svc.Annotations[protocolAnnotation]
	if protocolValue == "" {
		// if `<port>.service.kuma.io/protocol` annotation is missing or has an empty value
		// we want Dataplane to have a `protocol: tcp` tag in order to get user's attention
		return mesh_core.ProtocolTCP
	}
	// if `<port>.service.kuma.io/protocol` annotation is present but has an invalid value
	// we still want Dataplane to have a `protocol: <value as is>` tag in order to make it clear
	// to a user that at least `<port>.service.kuma.io/protocol` has an effect
	return protocolValue
}

func InboundTagsForPod(zone string, pod *kube_core.Pod) map[string]string {
	tags := util_k8s.CopyStringMap(pod.Labels)
	for key, value := range tags {
		if value == "" {
			delete(tags, key)
		}
	}
	if tags == nil {
		tags = make(map[string]string)
	}
	tags[mesh_proto.ServiceTag] = fmt.Sprintf("%s_%s_svc", nameFromPod(pod), pod.Namespace)
	if zone != "" {
		tags[mesh_proto.ZoneTag] = zone
	}
	tags[mesh_proto.ProtocolTag] = mesh_core.ProtocolTCP
	tags[mesh_proto.InstanceTag] = pod.Name

	return tags
}

func nameFromPod(pod *kube_core.Pod) string {
	// the name is in format <name>-<replica set id>-<pod id>
	split := strings.Split(pod.Name, "-")
	if len(split) > 2 {
		split = split[:len(split)-2]
	}

	return strings.Join(split, "-")
}
