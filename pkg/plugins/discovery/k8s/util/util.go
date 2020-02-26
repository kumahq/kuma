package util

import (
	"fmt"

	kube_core "k8s.io/api/core/v1"
	kube_labels "k8s.io/apimachinery/pkg/labels"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
)

type ServicePredicate func(*kube_core.Service) bool

func MatchServiceThatSelectsPod(pod *kube_core.Pod) ServicePredicate {
	return func(svc *kube_core.Service) bool {
		selector := kube_labels.SelectorFromSet(svc.Spec.Selector)
		return selector.Matches(kube_labels.Set(pod.Labels))
	}
}

func FindServices(svcs *kube_core.ServiceList, predicate ServicePredicate) []*kube_core.Service {
	matching := make([]*kube_core.Service, 0)
	for i := range svcs.Items {
		svc := &svcs.Items[i]
		if predicate(svc) {
			matching = append(matching, svc)
		}
	}
	return matching
}

// FindPort locates the container port for the given pod and portName.  If the
// targetPort is a number, use that.  If the targetPort is a string, look that
// string up in all named ports in all containers in the target pod.  If no
// match is found, fail.
func FindPort(pod *kube_core.Pod, svcPort *kube_core.ServicePort) (int, error) {
	givenOrDefault := func(value kube_core.Protocol) kube_core.Protocol {
		if value != "" {
			return value
		}
		return kube_core.ProtocolTCP
	}

	portName := svcPort.TargetPort
	switch portName.Type {
	case kube_intstr.String:
		name := portName.StrVal
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				if port.Name == name && givenOrDefault(port.Protocol) == givenOrDefault(svcPort.Protocol) {
					return int(port.ContainerPort), nil
				}
			}
		}
	case kube_intstr.Int:
		return portName.IntValue(), nil
	}

	return 0, fmt.Errorf("no suitable port for manifest: %s", pod.UID)
}

func CopyStringMap(in map[string]string) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string)
	for key, value := range in {
		out[key] = value
	}
	return out
}
