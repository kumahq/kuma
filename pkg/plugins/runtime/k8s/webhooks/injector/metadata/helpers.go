package metadata

import (
	"strconv"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"

	kube_core "k8s.io/api/core/v1"
)

func GetMesh(pod *kube_core.Pod) string {
	if mesh := pod.Annotations[KumaMeshAnnotation]; mesh != "" {
		return mesh
	}
	return core_model.DefaultMesh
}

func HasKumaSidecar(pod *kube_core.Pod) bool {
	return pod.Annotations[KumaSidecarInjectedAnnotation] == KumaSidecarInjected
}

func HasTransparentProxyingEnabled(pod *kube_core.Pod) bool {
	return pod.Annotations[KumaTransparentProxyingAnnotation] == KumaTransparentProxyingEnabled
}

func HasGatewayEnabled(pod *kube_core.Pod) bool {
	return pod.Annotations[KumaGatewayAnnotation] == KumaGatewayEnabled
}

func GetTransparentProxyingInboundPort(pod *kube_core.Pod) uint32 {
	port, err := strconv.ParseUint(pod.Annotations[KumaTransparentProxyingInboundPortAnnotation], 10, 32)
	if err != nil {
		return 0
	}
	return uint32(port)
}

func GetTransparentProxyingOutboundPort(pod *kube_core.Pod) uint32 {
	port, err := strconv.ParseUint(pod.Annotations[KumaTransparentProxyingOutboundPortAnnotation], 10, 32)
	if err != nil {
		return 0
	}
	return uint32(port)
}
