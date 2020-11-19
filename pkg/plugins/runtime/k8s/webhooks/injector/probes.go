package injector

import (
	"fmt"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
)

func (i *KumaInjector) overrideHTTPProbes(pod *kube_core.Pod) error {
	log.WithValues("name", pod.Name, "namespace", pod.Namespace)
	enabled, _, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaVirtualProbesAnnotation)
	if err != nil {
		return err
	}
	if !enabled {
		log.V(1).Info("skip adding virtual probes")
		return err
	}

	port, _, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return err
	}

	for _, c := range pod.Spec.Containers {
		if c.LivenessProbe != nil && c.LivenessProbe.HTTPGet != nil {
			log.V(1).Info("overriding liveness probe", "container", c.Name)
			if err := overrideHTTPProbe(c.LivenessProbe, port); err != nil {
				return err
			}
		}
		if c.ReadinessProbe != nil && c.ReadinessProbe.HTTPGet != nil {
			log.V(1).Info("overriding readiness probe", "container", c.Name)
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


func virtualProbesEnabled(annotations metadata.Annotations, pod *kube_core.Pod, cfg runtime_k8s.Injector) error {
	str := func(b bool) string {
		if b {
			return metadata.AnnotationEnabled
		}
		return metadata.AnnotationDisabled
	}

	vpEnabled, vpExist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaVirtualProbesAnnotation)
	if err != nil {
		return err
	}
	gwEnabled, _, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaGatewayAnnotation)
	if err != nil {
		return err
	}

	if gwEnabled {
		if vpEnabled {
			return errors.New("virtual probes can't be enabled in gateway mode")
		}
		annotations[metadata.KumaVirtualProbesAnnotation] = metadata.AnnotationDisabled
		return nil
	}

	if vpExist {
		annotations[metadata.KumaVirtualProbesAnnotation] = str(vpEnabled)
		return nil
	}
	annotations[metadata.KumaVirtualProbesAnnotation] = str(cfg.VirtualProbesEnabled)
	return nil
}

func virtualProbesPort(annotations metadata.Annotations, pod *kube_core.Pod, cfg runtime_k8s.Injector) error {
	port, exist, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return err
	}
	if exist {
		annotations[metadata.KumaVirtualProbesPortAnnotation] = fmt.Sprintf("%d", port)
		return nil
	}
	annotations[metadata.KumaVirtualProbesPortAnnotation] = fmt.Sprintf("%d", cfg.VirtualProbesPort)
	return nil
}
