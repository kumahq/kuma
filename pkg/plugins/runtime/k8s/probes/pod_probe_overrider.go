package probes

import (
	"fmt"
	"github.com/go-logr/logr"
	"strconv"

	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func OverridePodProbes(pod *kube_core.Pod, log logr.Logger) error {
	log.WithValues("name", pod.Name, "namespace", pod.Namespace)
	enabled, _, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaVirtualProbesAnnotation)
	if err != nil {
		return err
	}
	if !enabled {
		log.V(1).Info("skipping adding virtual probes, because it's disabled")
		return err
	}

	virtualProbesPort, _, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return err
	}
	var containersNeedingProbes []kube_core.Container

	var initContainerComesAfterKumaSidecar bool
	for _, c := range pod.Spec.InitContainers {
		if c.Name == util.KumaSidecarContainerName {
			initContainerComesAfterKumaSidecar = true
			continue
		}

		if initContainerComesAfterKumaSidecar && c.RestartPolicy != nil && *c.RestartPolicy == kube_core.ContainerRestartPolicyAlways {
			containersNeedingProbes = append(containersNeedingProbes, c)
		}
	}
	for _, c := range pod.Spec.Containers {
		if c.Name != util.KumaSidecarContainerName {
			// we don't want to create virtual probes for Envoy container, because we generate real listener which is not protected by mTLS
			containersNeedingProbes = append(containersNeedingProbes, c)
		}
	}
	for _, c := range containersNeedingProbes {
		portResolver := namedPortResolver(&c)
		if err := overrideProbe(c.LivenessProbe, virtualProbesPort,
			portResolver, c.Name, "liveness", log); err != nil {
			return err
		}
		if err := overrideProbe(c.ReadinessProbe, virtualProbesPort,
			portResolver, c.Name, "readiness", log); err != nil {
			return err
		}
		if err := overrideProbe(c.StartupProbe, virtualProbesPort,
			portResolver, c.Name, "startup", log); err != nil {
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

func overrideProbe(probe *kube_core.Probe, virtualPort uint32,
	namedPortResolver func(kube_core.ProbeHandler), containerName, probeName string, log logr.Logger) error {
	if probe == nil {
		return nil
	}

	kumaProbe := KumaProbe(*probe)
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
		Port:        intstr.FromInt32(int32(virtual.Port())),
		Path:        virtual.Path(),
		HTTPHeaders: virtual.Headers(),
	}
	return nil
}

func SetVirtualProbesEnabledAnnotation(annotations metadata.Annotations, podAnnotations map[string]string, virtualProbesEnabled bool) error {
	str := func(b bool) string {
		if b {
			return metadata.AnnotationEnabled
		}
		return metadata.AnnotationDisabled
	}

	vpEnabled, vpExist, err := metadata.Annotations(podAnnotations).GetEnabled(metadata.KumaVirtualProbesAnnotation)
	if err != nil {
		return err
	}
	_, gwExists := metadata.Annotations(podAnnotations).GetString(metadata.KumaGatewayAnnotation)
	if gwExists {
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
	annotations[metadata.KumaVirtualProbesAnnotation] = str(virtualProbesEnabled)
	return nil
}

func SetVirtualProbesPortAnnotation(annotations metadata.Annotations, podAnnotations map[string]string, defaultVirtualProbesPort uint32) error {
	port, _, err := metadata.Annotations(podAnnotations).GetUint32WithDefault(defaultVirtualProbesPort, metadata.KumaVirtualProbesPortAnnotation)
	if err != nil {
		return err
	}
	annotations[metadata.KumaVirtualProbesPortAnnotation] = strconv.Itoa(int(port))
	return nil
}
