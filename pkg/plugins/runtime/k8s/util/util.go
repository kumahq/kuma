package util

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"golang.org/x/exp/maps"
	kube_core "k8s.io/api/core/v1"
	kube_labels "k8s.io/apimachinery/pkg/labels"
	kube_types "k8s.io/apimachinery/pkg/types"
	kube_intstr "k8s.io/apimachinery/pkg/util/intstr"
	kube_client "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

type ServicePredicate func(*kube_core.Service) bool

func MatchServiceThatSelectsPod(pod *kube_core.Pod, ignoredLabels []string) ServicePredicate {
	return func(svc *kube_core.Service) bool {
		selector := maps.Clone(svc.Spec.Selector)
		for _, ignoredLabel := range ignoredLabels {
			delete(selector, ignoredLabel)
		}
		return kube_labels.SelectorFromSet(selector).Matches(kube_labels.Set(pod.Labels))
	}
}

// According to K8S docs about Service#selector:
// Route service traffic to pods with label keys and values matching this selector.
// If empty or not present, the service is assumed to have an external process managing its endpoints, which Kubernetes will not modify.
// Only applies to types ClusterIP, NodePort, and LoadBalancer. Ignored if type is ExternalName. More info: https://kubernetes.io/docs/concepts/services-networking/service/
//
// When converting Pod to Dataplane, we don't want to take into account Services that has no Selector, otherwise any Pod will match this service
// and since we just take any int target port in #util.FindPort every Dataplane in the same namespace as this service would get an extra inbound for it.
func AnySelector() ServicePredicate {
	return func(svc *kube_core.Service) bool {
		return len(svc.Spec.Selector) > 0
	}
}

func Not(predicate ServicePredicate) ServicePredicate {
	return func(svc *kube_core.Service) bool {
		return !predicate(svc)
	}
}

func Ignored() ServicePredicate {
	return func(svc *kube_core.Service) bool {
		if svc.Annotations == nil {
			return false
		}
		ignore, _, _ := metadata.Annotations(svc.Annotations).GetEnabled(metadata.KumaIgnoreAnnotation)
		return ignore
	}
}

func MatchService(svc *kube_core.Service, predicates ...ServicePredicate) bool {
	for _, predicate := range predicates {
		if !predicate(svc) {
			return false
		}
	}
	return true
}

// FindPort locates the container port for the given pod and portName.  If the
// targetPort is a number, use that.  If the targetPort is a string, look that
// string up in all named ports in all containers in the target pod.  If no
// match is found, fail.
func FindPort(pod *kube_core.Pod, svcPort *kube_core.ServicePort) (int, *kube_core.Container, error) {
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
					return int(port.ContainerPort), &container, nil
				}
			}
		}
	case kube_intstr.Int:
		// According to K8S docs about Container#ports:
		// List of ports to expose from the container. Exposing a port here gives the system additional information about the network connections a container uses, but is primarily informational.
		// Not specifying a port here DOES NOT prevent that port from being exposed. Any port which is listening on the default "0.0.0.0" address inside a container will be accessible from the network
		//
		// Therefore we cannot match service port to the container port.
		for _, container := range pod.Spec.Containers {
			for _, port := range container.Ports {
				if port.ContainerPort == portName.IntVal && givenOrDefault(port.Protocol) == givenOrDefault(svcPort.Protocol) {
					return int(port.ContainerPort), &container, nil
				}
			}
		}
		return portName.IntValue(), nil, nil
	}

	return 0, nil, fmt.Errorf("no suitable port for manifest: %s", pod.UID)
}

func findContainerStatus(containerName string, status []kube_core.ContainerStatus, initStatus []kube_core.ContainerStatus) *kube_core.ContainerStatus {
	for _, cs := range append(status, initStatus...) {
		if cs.Name == containerName {
			return &cs
		}
	}
	return nil
}

func FindContainerStatus(containerName string, status []kube_core.ContainerStatus) *kube_core.ContainerStatus {
	return findContainerStatus(containerName, status, nil)
}

func FindContainerOrInitContainerStatus(containerName string, status []kube_core.ContainerStatus, initStatus []kube_core.ContainerStatus) *kube_core.ContainerStatus {
	return findContainerStatus(containerName, status, initStatus)
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

// MeshOfByLabelOrAnnotation returns the mesh of the given object according to its own
// annotations or labels or the annotations of its namespace. It treats the annotation
// directly on the object as deprecated.
func MeshOfByLabelOrAnnotation(log logr.Logger, obj kube_client.Object, namespace *kube_core.Namespace) string {
	if mesh, exists := metadata.Annotations(obj.GetLabels()).GetString(metadata.KumaMeshLabel); exists && mesh != "" {
		return mesh
	}
	if mesh, exists := metadata.Annotations(obj.GetAnnotations()).GetString(metadata.KumaMeshAnnotation); exists && mesh != "" {
		log.Info("WARNING: The kuma.io/mesh annotation is deprecated for this object kind. Use label instead", "name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", obj.GetObjectKind().GroupVersionKind().Kind)
		return mesh
	}

	// Label wasn't found on the object, let's look on the namespace instead
	if mesh, exists := metadata.Annotations(namespace.GetLabels()).GetString(metadata.KumaMeshLabel); exists && mesh != "" {
		return mesh
	}

	if mesh, exists := metadata.Annotations(namespace.GetAnnotations()).GetString(metadata.KumaMeshAnnotation); exists && mesh != "" {
		log.Info("WARNING: The kuma.io/mesh annotation is deprecated for this object kind. Use label instead", "name", obj.GetName(), "namespace", obj.GetNamespace(), "kind", obj.GetObjectKind().GroupVersionKind().Kind)
		return mesh
	}

	return model.DefaultMesh
}

// ServiceTag returns the canonical service name for a Kubernetes service,
// optionally with a specific port.
func ServiceTag(name kube_types.NamespacedName, svcPort *int32) string {
	port := ""
	if svcPort != nil {
		port = fmt.Sprintf("_%d", *svcPort)
	}
	return fmt.Sprintf("%s_%s_svc%s", name.Name, name.Namespace, port)
}

func NamespacesNameFromServiceTag(serviceName string) (kube_types.NamespacedName, error) {
	split := strings.Split(serviceName, "_")
	if len(split) >= 2 {
		return kube_types.NamespacedName{
			Name:      split[0],
			Namespace: split[1],
		}, nil
	}
	return kube_types.NamespacedName{}, fmt.Errorf("incorrect service name: %s", serviceName)
}
