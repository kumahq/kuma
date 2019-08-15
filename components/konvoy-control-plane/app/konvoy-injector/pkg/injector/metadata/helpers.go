package metadata

import (
	"strconv"

	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"

	kube_core "k8s.io/api/core/v1"
)

func GetMesh(pod *kube_core.Pod) string {
	if mesh := pod.Annotations[KonvoyMeshAnnotation]; mesh != "" {
		return mesh
	}
	return core_model.DefaultMesh
}

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
