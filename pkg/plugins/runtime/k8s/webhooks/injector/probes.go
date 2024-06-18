package injector

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	runtime_k8s "github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func (i *KumaInjector) overrideProbes(pod *kube_core.Pod) error {
	log.WithValues("name", pod.Name, "namespace", pod.Namespace)
	enabled, _, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaVirtualProbesAnnotation)
	if err != nil {
		return err
	}
	if !enabled {
		log.V(1).Info("skip adding virtual probes")
		return err
	}

	virtualProbesPort, _, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return err
	}

	for _, c := range pod.Spec.Containers {
		if c.Name == util.KumaSidecarContainerName {
			// we don't want to create virtual probes for Envoy container, because we generate real listener which is not protected by mTLS
			continue
		}

		portResolver := namedPortResolver(&c)
		if err := tryOverrideProbe(c.LivenessProbe, virtualProbesPort,
			portResolver, c.Name, "liveness"); err != nil {
			return err
		}
		if err := tryOverrideProbe(c.ReadinessProbe, virtualProbesPort,
			portResolver, c.Name, "readiness"); err != nil {
			return err
		}
		if err := tryOverrideProbe(c.StartupProbe, virtualProbesPort,
			portResolver, c.Name, "startup"); err != nil {
			return err
		}
	}
	return nil
}

func namedPortResolver(container *kube_core.Container) func(kube_core.ProbeHandler) {
	return func(probe kube_core.ProbeHandler) {
		var portStr intstr.IntOrString
		if probe.HTTPGet != nil {
			portStr = probe.HTTPGet.Port
		} else if probe.TCPSocket != nil {
			portStr = probe.TCPSocket.Port
		} else {
			return
		}

		if portStr.IntValue() != 0 {
			return
		}

		for _, containerPort := range container.Ports {
			if containerPort.Name != "" && containerPort.Name == portStr.String() {
				if probe.HTTPGet != nil {
					probe.HTTPGet.Port = intstr.FromInt32(containerPort.ContainerPort)
				} else if probe.TCPSocket != nil {
					probe.TCPSocket.Port = intstr.FromInt32(containerPort.ContainerPort)
				}

				break
			}
		}
	}
}

func tryOverrideProbe(probe *kube_core.Probe, virtualPort uint32, namedPortResolver func(kube_core.ProbeHandler), containerName, probeName string) error {
	if probe == nil {
		return nil
	}

	kumaProbe := probes.KumaProbe(*probe)
	if !kumaProbe.OverridingSupported() {
		return nil
	}

	log.V(1).Info(fmt.Sprintf("overriding %s probe", probeName), "container", containerName)

	namedPortResolver(probe.ProbeHandler)

	virtual, err := kumaProbe.ToVirtual(virtualPort)
	if err != nil {
		return err
	}

	probe.GRPC = nil
	probe.TCPSocket = nil
	probe.HTTPGet = &kube_core.HTTPGetAction{
		Scheme:      kube_core.URISchemeHTTP,
		Port:        intstr.FromInt32(int32(virtual.Port())),
		Path:        virtual.Path(),
		HTTPHeaders: virtual.Headers(),
	}
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
