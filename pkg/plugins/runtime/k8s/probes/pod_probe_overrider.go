package probes

import (
	"fmt"
	"slices"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/util"
)

func containersNeedingProbes(pod *kube_core.Pod) []kube_core.Container {
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
			// we don't want to proxy probes for the Envoy container itself
			containersNeedingProbes = append(containersNeedingProbes, c)
		}
	}
	return containersNeedingProbes
}

// RealProbePorts returns the ports the pod's own containers use for their
// Liveness/Readiness/Startup probes, resolving named container ports. It is
// used to exclude those ports from mTLS interception when application probe
// proxying is disabled, since in that case K8s probes hit them directly.
// Unlike SetupAppProbeProxies, it doesn't mutate the pod's probes.
func RealProbePorts(pod *kube_core.Pod) []uint32 {
	var ports []uint32
	seen := map[uint32]struct{}{}
	for _, c := range containersNeedingProbes(pod) {
		for _, probe := range []*kube_core.Probe{c.LivenessProbe, c.ReadinessProbe, c.StartupProbe} {
			if probe == nil {
				continue
			}
			port, ok := resolveProbePort(probe.ProbeHandler, c.Ports)
			if !ok {
				continue
			}
			if _, ok := seen[port]; ok {
				continue
			}
			seen[port] = struct{}{}
			ports = append(ports, port)
		}
	}
	slices.Sort(ports)
	return ports
}

// resolveProbePort returns the port a probe handler targets, resolving named
// container ports, without mutating the probe.
func resolveProbePort(handler kube_core.ProbeHandler, containerPorts []kube_core.ContainerPort) (uint32, bool) {
	var portStr intstr.IntOrString
	switch {
	case handler.HTTPGet != nil:
		portStr = handler.HTTPGet.Port
	case handler.TCPSocket != nil:
		portStr = handler.TCPSocket.Port
	case handler.GRPC != nil:
		return uint32(handler.GRPC.Port), true
	default:
		return 0, false
	}

	if portStr.IntValue() != 0 {
		return uint32(portStr.IntValue()), true
	}
	for _, containerPort := range containerPorts {
		if containerPort.Name != "" && containerPort.Name == portStr.String() {
			return uint32(containerPort.ContainerPort), true
		}
	}
	return 0, false
}

func SetupAppProbeProxies(pod *kube_core.Pod, log logr.Logger) error {
	log.WithValues("name", pod.Name, "namespace", pod.Namespace)
	appProbeProxyPort, _, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaApplicationProbeProxyPortAnnotation)
	if err != nil {
		return err
	}
	if appProbeProxyPort == 0 {
		log.V(1).Info("skipping adding application probe proxies, because it's disabled")
		return err
	}

	for _, c := range containersNeedingProbes(pod) {
		portResolver := namedPortResolver(c.Ports)
		if err := overrideProbe(c.LivenessProbe, appProbeProxyPort,
			portResolver, c.Name, "liveness", log); err != nil {
			return err
		}
		if err := overrideProbe(c.ReadinessProbe, appProbeProxyPort,
			portResolver, c.Name, "readiness", log); err != nil {
			return err
		}
		if err := overrideProbe(c.StartupProbe, appProbeProxyPort,
			portResolver, c.Name, "startup", log); err != nil {
			return err
		}
	}
	return nil
}

func namedPortResolver(containerPorts []kube_core.ContainerPort) func(kube_core.ProbeHandler) {
	return func(probe kube_core.ProbeHandler) {
		var portStr intstr.IntOrString
		switch {
		case probe.HTTPGet != nil:
			portStr = probe.HTTPGet.Port
		case probe.TCPSocket != nil:
			portStr = probe.TCPSocket.Port
		default:
			return
		}

		if portStr.IntValue() != 0 {
			return
		}

		for _, containerPort := range containerPorts {
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
	namedPortResolver func(kube_core.ProbeHandler), containerName, probeName string, log logr.Logger,
) error {
	if probe == nil {
		return nil
	}

	proxiedProbe := ProxiedApplicationProbe(*probe)
	if !proxiedProbe.OverridingSupported() {
		return nil
	}

	log.V(1).Info("overriding probe", "probe", probeName, "container", containerName)

	namedPortResolver(probe.ProbeHandler)

	virtual, err := proxiedProbe.ToVirtual(virtualPort)
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

func SetApplicationProbeProxyPortAnnotation(annotations metadata.Annotations, podAnnotations map[string]string, defaultAppProbeProxyPort uint32) error {
	str := func(port uint32) string {
		return fmt.Sprintf("%d", port)
	}

	proxyPortAnno, proxyPortAnnoExists, err := metadata.Annotations(podAnnotations).GetUint32(metadata.KumaApplicationProbeProxyPortAnnotation)
	if err != nil {
		return err
	}
	appProbeProxyPort := defaultAppProbeProxyPort
	if proxyPortAnnoExists {
		appProbeProxyPort = proxyPortAnno
	}
	gwEnabled, _, err := metadata.Annotations(podAnnotations).GetEnabled(metadata.KumaGatewayAnnotation)
	if err != nil {
		return err
	}
	if gwEnabled {
		if proxyPortAnnoExists && proxyPortAnno > 0 {
			return errors.New("application probe proxies probes can't be enabled in gateway mode")
		}
		annotations[metadata.KumaApplicationProbeProxyPortAnnotation] = "0"
		return nil
	}

	annotations[metadata.KumaApplicationProbeProxyPortAnnotation] = str(appProbeProxyPort)
	return nil
}

func GetApplicationProbeProxyPort(
	annotations metadata.Annotations,
	defaultAppProbeProxyPort uint32,
) (uint32, error) {
	proxyPort, proxyPortExist, err := annotations.GetUint32(metadata.KumaApplicationProbeProxyPortAnnotation)
	if err != nil {
		return 0, err
	}

	gwEnabled, _, _ := annotations.GetEnabled(metadata.KumaGatewayAnnotation)

	switch {
	case gwEnabled && proxyPort > 0:
		return 0, errors.New("application probe proxies probes can't be enabled in gateway mode")
	case gwEnabled:
		return 0, nil
	case proxyPortExist:
		return proxyPort, nil
	default:
		return defaultAppProbeProxyPort, nil
	}
}
