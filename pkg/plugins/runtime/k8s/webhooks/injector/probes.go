package injector

import (
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
)

func (i *KumaInjector) overrideHTTPProbes(pod *kube_core.Pod) error {
	if ok, err := needVirtualProbes(pod, i.cfg); err != nil || !ok {
		return err
	}

	port, err := virtualProbesPort(pod, i.cfg)
	if err != nil {
		return err
	}

	for _, c := range pod.Spec.Containers {
		if c.LivenessProbe != nil && c.LivenessProbe.HTTPGet != nil {
			if err := overrideHTTPProbe(c.LivenessProbe, port); err != nil {
				return err
			}
		}
		if c.ReadinessProbe != nil && c.ReadinessProbe.HTTPGet != nil {
			if err := overrideHTTPProbe(c.ReadinessProbe, port); err != nil {
				return err
			}
		}
	}
	return nil
}

func overrideHTTPProbe(probe *kube_core.Probe, virtualPort uint32) error {
	virtual, err := probes.KumaProbe(*probe).ToVirtual(virtualPort)
	if err != nil {
		return err
	}
	probe.HTTPGet.Port = intstr.FromInt(int(virtual.Port()))
	probe.HTTPGet.Path = virtual.Path()
	return nil
}

func needVirtualProbes(pod *kube_core.Pod, cfg runtime_k8s.Injector) (bool, error) {
	enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaVirtualProbesAnnotation)
	if err != nil {
		return false, err
	}
	if exist {
		return enabled, nil
	}
	return cfg.VirtualProbesEnabled, nil
}

func virtualProbesPort(pod *kube_core.Pod, cfg runtime_k8s.Injector) (uint32, error) {
	port, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return 0, err
	}
	if exist {
		return port, nil
	}
	return cfg.VirtualProbesPort, nil
}
