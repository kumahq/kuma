package injector

import (
	kube_core "k8s.io/api/core/v1"
)

func InjectKonvoy(pod *kube_core.Pod) error {
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	pod.Annotations["getkonvoy.io/sidecar-injected"] = "true"
	return nil
}
