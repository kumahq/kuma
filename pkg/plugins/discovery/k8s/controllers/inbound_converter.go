package controllers

import (
	"fmt"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	util_k8s "github.com/Kong/kuma/pkg/plugins/discovery/k8s/util"
)

func InboundInterfacesFor(clusterName string, pod *kube_core.Pod, services []*kube_core.Service, isGateway bool) ([]*mesh_proto.Dataplane_Networking_Inbound, error) {
	var ifaces []*mesh_proto.Dataplane_Networking_Inbound
	for _, svc := range services {
		for _, svcPort := range svc.Spec.Ports {
			if svcPort.Protocol != "" && svcPort.Protocol != kube_core.ProtocolTCP {
				// ignore non-TCP ports
				continue
			}
			containerPort, err := util_k8s.FindPort(pod, &svcPort)
			if err != nil {
				converterLog.Error(err, "failed to find a container port in a given Pod that would match a given Service port", "namespace", pod.Namespace, "podName", pod.Name, "serviceName", svc.Name, "servicePortName", svcPort.Name)
				// ignore those cases where a Pod doesn't have all the ports a Service has
				continue
			}

			tags := InboundTagsFor(clusterName, pod, svc, &svcPort, isGateway)

			ifaces = append(ifaces, &mesh_proto.Dataplane_Networking_Inbound{
				Port: uint32(containerPort),
				Tags: tags,
			})
		}
	}
	if len(ifaces) == 0 {
		// Notice that here we return an error immediately
		// instead of leaving Dataplane validation up to a ValidatingAdmissionWebHook.
		// We do it this way in order to provide the most descriptive error message.
		cause := "However, there are no Services that select this Pod."
		if len(services) > 0 {
			cause = "However, this Pod doesn't have any container ports that would satisfy matching Service(s)."
		}
		return nil, errors.Errorf("Kuma requires every Pod in a Mesh to be a part of at least one Service. %s", cause)
	}
	return ifaces, nil
}

func InboundTagsFor(clusterName string, pod *kube_core.Pod, svc *kube_core.Service, svcPort *kube_core.ServicePort, isGateway bool) map[string]string {
	tags := util_k8s.CopyStringMap(pod.Labels)
	if tags == nil {
		tags = make(map[string]string)
	}
	tags[mesh_proto.ServiceTag] = ServiceTagFor(svc, svcPort)
	if clusterName != "" {
		tags[mesh_proto.ClusterTag] = clusterName
	}
	// notice that in case of a gateway it might be confusing to see a protocol tag
	// since gateway proxies multiple services each with its own protocol
	if !isGateway {
		tags[mesh_proto.ProtocolTag] = ProtocolTagFor(svc, svcPort)
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
