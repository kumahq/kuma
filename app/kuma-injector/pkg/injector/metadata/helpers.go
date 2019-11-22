package metadata

import (
	"fmt"
	"strconv"

	core_model "github.com/Kong/kuma/pkg/core/resources/model"

	kube_core "k8s.io/api/core/v1"
)

func GetMesh(pod *kube_core.Pod, namespace string) string {
	if mesh := pod.Annotations[KumaMeshAnnotation]; mesh != "" {
		return mesh
	}
	return fmt.Sprintf("%s.%s", core_model.DefaultMesh, namespace)
}

func HasKumaSidecar(pod *kube_core.Pod) bool {
	return pod.Annotations[KumaSidecarInjectedAnnotation] == KumaSidecarInjected
}

func HasTransparentProxyingEnabled(pod *kube_core.Pod) bool {
	return pod.Annotations[KumaTransparentProxyingAnnotation] == KumaTransparentProxyingEnabled
}

func GetTransparentProxyingPort(pod *kube_core.Pod) uint32 {
	port, err := strconv.ParseUint(pod.Annotations[KumaTransparentProxyingPortAnnotation], 10, 32)
	if err != nil {
		return 0
	}
	return uint32(port)
}
