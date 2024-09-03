package probes

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
)

func ApplicationProbeProxyDisabled(pod *kube_core.Pod) (bool, error) {
	appProbeProxyPort, _, err := metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaApplicationProbeProxyPortAnnotation)
	if err != nil {
		return false, err
	}
	return appProbeProxyPort == 0, nil
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

	log.V(1).Info(fmt.Sprintf("overriding %s probe", probeName), "container", containerName)

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

	// scenarios of switching between virtual probes (vp) and application probe proxy (proxy):
	// vp   proxy    	result
	// Y    Y      	     proxy
	// Y    N            vp
	// N    N      	     N
	// N    Y - config   N
	// N    Y - pod      proxy

	// if disabled by "kuma.io/virtual-probes", we honor it when there is no "kuma.io/application-probe-proxy-port" annotation
	// this is treated as deprecated though
	proxyPortAnno, proxyPortAnnoExists, err := metadata.Annotations(podAnnotations).GetUint32(metadata.KumaApplicationProbeProxyPortAnnotation)
	if err != nil {
		return err
	}
	if vpEnabled, _, _ := annotations.GetEnabled(metadata.KumaVirtualProbesAnnotation); !vpEnabled && !proxyPortAnnoExists {
		annotations[metadata.KumaApplicationProbeProxyPortAnnotation] = "0"
		return nil
	}
	appProbeProxyPort := defaultAppProbeProxyPort
	if proxyPortAnnoExists {
		appProbeProxyPort = proxyPortAnno
	}
	_, gwExists := metadata.Annotations(podAnnotations).GetString(metadata.KumaGatewayAnnotation)
	if gwExists {
		if proxyPortAnnoExists && proxyPortAnno > 0 {
			return errors.New("application probe proxies probes can't be enabled in gateway mode")
		}
		annotations[metadata.KumaApplicationProbeProxyPortAnnotation] = "0"
		return nil
	}

	annotations[metadata.KumaApplicationProbeProxyPortAnnotation] = str(appProbeProxyPort)
	return nil
}
