package metadata

import (
	kube_core "k8s.io/api/core/v1"
)

func HasKonvoySidecar(pod *kube_core.Pod) bool {
	return pod.Annotations[KonvoySidecarInjectedAnnotation] == KonvoySidecarInjectedAnnotationValue
}
