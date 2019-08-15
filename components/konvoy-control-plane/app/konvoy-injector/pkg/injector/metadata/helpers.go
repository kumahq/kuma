package metadata

import (
	"strconv"

	kube_core "k8s.io/api/core/v1"
)

func HasKonvoySidecar(pod *kube_core.Pod) bool {
	return pod.Annotations[KonvoySidecarInjectedAnnotation] == KonvoySidecarInjected
}

func HasTransparentProxyingEnabled(pod *kube_core.Pod) bool {
	return pod.Annotations[KonvoyTransparentProxyingAnnotation] == KonvoyTransparentProxyingEnabled
}

func GetTransparentProxyingPort(pod *kube_core.Pod) uint32 {
	port, err := strconv.ParseUint(pod.Annotations[KonvoyTransparentProxyingPortAnnotation], 10, 32)
	if err != nil {
		return 0
	}
	return uint32(port)
}
