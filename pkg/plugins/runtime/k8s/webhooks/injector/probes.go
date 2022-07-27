package injector

import (
	"strconv"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
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
		if c.Name == util.KumaSidecarContainerName {
			// we don't want to create virtual probes for Envoy container, because we generate real listener which is not protected by mTLS
			continue
		}
		if c.LivenessProbe != nil && c.LivenessProbe.HTTPGet != nil {
			log.V(1).Info("overriding liveness probe", "container", c.Name)
			resolveNamedPort(c, c.LivenessProbe)
			if err := overrideHTTPProbe(c.LivenessProbe, port); err != nil {
				return err
			}
		}
		if c.ReadinessProbe != nil && c.ReadinessProbe.HTTPGet != nil {
			log.V(1).Info("overriding readiness probe", "container", c.Name)
			resolveNamedPort(c, c.ReadinessProbe)
			if err := overrideHTTPProbe(c.ReadinessProbe, port); err != nil {
				return err
			}
		}
		if c.StartupProbe != nil && c.StartupProbe.HTTPGet != nil {
			log.V(1).Info("overriding startup probe", "container", c.Name)
			resolveNamedPort(c, c.StartupProbe)
			if err := overrideHTTPProbe(c.StartupProbe, port); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolveNamedPort(container kube_core.Container, probe *kube_core.Probe) {
	port := probe.HTTPGet.Port
	if port.IntValue() != 0 {
		return
	}
	for _, containerPort := range container.Ports {
		if containerPort.Name != "" && containerPort.Name == port.String() {
			probe.HTTPGet.Port = intstr.FromInt(int(containerPort.ContainerPort))
		}
	}
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

func setVirtualProbesEnabledAnnotation(annotations metadata.Annotations, pod *kube_core.Pod, cfg runtime_k8s.Injector) error {
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

func setVirtualProbesPortAnnotation(annotations metadata.Annotations, pod *kube_core.Pod, cfg runtime_k8s.Injector) error {
	port, _, err := metadata.Annotations(pod.Annotations).GetUint32WithDefault(cfg.VirtualProbesPort, metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return err
	}
	annotations[metadata.KumaVirtualProbesPortAnnotation] = strconv.Itoa(int(port))
	return nil
}
