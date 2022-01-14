package controllers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

const (
	KubeNamespaceTag = "k8s.kuma.io/namespace"
	KubeServiceTag   = "k8s.kuma.io/service-name"
	KubePortTag      = "k8s.kuma.io/service-port"
)

func inboundForService(zone string, pod *kube_core.Pod, service *kube_core.Service) (ifaces []*mesh_proto.Dataplane_Networking_Inbound) {
	for _, svcPort := range service.Spec.Ports {
		if svcPort.Protocol != "" && svcPort.Protocol != kube_core.ProtocolTCP {
			// ignore non-TCP ports
			continue
		}
		containerPort, container, err := util_k8s.FindPort(pod, &svcPort)
		if err != nil {
			converterLog.Error(err, "failed to find a container port in a given Pod that would match a given Service port", "namespace", pod.Namespace, "podName", pod.Name, "serviceName", service.Name, "servicePortName", svcPort.Name)
			// ignore those cases where a Pod doesn't have all the ports a Service has
			continue
		}

		tags := InboundTagsForService(zone, pod, service, &svcPort)
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

	return
}

func inboundForServiceless(zone string, pod *kube_core.Pod) (ifaces []*mesh_proto.Dataplane_Networking_Inbound) {
	// The Pod does not have any services associated with it, just get the data from the Pod itself

	// We still need that extra listener with a service because it is required in many places of the code (e.g. mTLS)
	// TCPPortReserved, is a special port that will never be allocated from the TCP/IP stack. We use it as special
	// designator that this is actually a service-less inbound.

	// NOTE: It is cleaner to implement an equivalent of Gateway which is inbound-less dataplane. However such approch
	// will create lots of code changes to account for this other type of dataplne (we already have GW and Ingress),
	// including GUI and CLI changes

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
		Port:   mesh_proto.TCPPortReserved,
		Tags:   tags,
		Health: health,
	})

	return
}

func InboundInterfacesFor(zone string, pod *kube_core.Pod, services []*kube_core.Service) ([]*mesh_proto.Dataplane_Networking_Inbound, error) {
	ifaces := []*mesh_proto.Dataplane_Networking_Inbound{}
	for _, svc := range services {
		svcIfaces := inboundForService(zone, pod, svc)
		ifaces = append(ifaces, svcIfaces...)
	}

	if len(ifaces) == 0 {
		if len(services) > 0 {
			return nil, errors.Errorf("A service that selects pod %s was found, but it doesn't match any container ports.", pod.GetName())
		}

		ifaces = append(ifaces, inboundForServiceless(zone, pod)...)
	}
	return ifaces, nil
}

func InboundTagsForService(zone string, pod *kube_core.Pod, svc *kube_core.Service, svcPort *kube_core.ServicePort) map[string]string {
	tags := util_k8s.CopyStringMap(pod.Labels)
	for key, value := range tags {
		if key == metadata.KumaSidecarInjectionAnnotation || value == "" {
			delete(tags, key)
		} else if strings.Contains(key, "kuma.io/") {
			// we don't want to convert labels like
			// kuma.io/sidecar-injection, kuma.io/service, k8s.kuma.io/namespace etc.
			converterLog.Info("ignoring label when converting labels to tags, because it uses reserved Kuma prefix", "label", key, "pod", pod.Name)
			delete(tags, key)
		}
	}
	if tags == nil {
		tags = make(map[string]string)
	}
	tags[KubeNamespaceTag] = pod.Namespace
	tags[KubeServiceTag] = svc.Name
	tags[KubePortTag] = strconv.Itoa(int(svcPort.Port))
	tags[mesh_proto.ServiceTag] = util_k8s.ServiceTagFor(svc, &svcPort.Port)
	if zone != "" {
		tags[mesh_proto.ZoneTag] = zone
	}
	tags[mesh_proto.ProtocolTag] = ProtocolTagFor(svc, svcPort)
	if isHeadlessService(svc) {
		tags[mesh_proto.InstanceTag] = pod.Name
	}
	return tags
}

// ProtocolTagFor infers service protocol from a `<port>.service.kuma.io/protocol` annotation or `appProtocol`.
func ProtocolTagFor(svc *kube_core.Service, svcPort *kube_core.ServicePort) string {
	var protocolValue string
	protocolAnnotation := fmt.Sprintf("%d.service.kuma.io/protocol", svcPort.Port)

	if svcPort.AppProtocol != nil {
		protocolValue = *svcPort.AppProtocol
	} else {
		protocolValue = svc.Annotations[protocolAnnotation]
	}

	if protocolValue == "" {
		// if `appProtocol` or `<port>.service.kuma.io/protocol` is missing or has an empty value
		// we want Dataplane to have a `protocol: tcp` tag in order to get user's attention
		return core_mesh.ProtocolTCP
	}
	// if `appProtocol` or `<port>.service.kuma.io/protocol` field is present but has an invalid value
	// we still want Dataplane to have a `protocol: <value as is>` tag in order to make it clear
	// to a user that at least `appProtocol` or `<port>.service.kuma.io/protocol` has an effect
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
	tags[KubeNamespaceTag] = pod.Namespace
	tags[mesh_proto.ServiceTag] = fmt.Sprintf("%s_%s_svc", nameFromPod(pod), pod.Namespace)
	if zone != "" {
		tags[mesh_proto.ZoneTag] = zone
	}
	tags[mesh_proto.ProtocolTag] = core_mesh.ProtocolTCP
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
