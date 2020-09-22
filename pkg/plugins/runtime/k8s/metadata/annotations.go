package metadata

import (
	"strconv"

	"github.com/pkg/errors"
)

// Annotations that can be used by the end users.
const (
	// KumaMeshAnnotation defines a Pod annotation that
	// associates a given Pod with a particular Mesh.
	// Annotation value must be the name of a Mesh resource.
	KumaMeshAnnotation = "kuma.io/mesh"

	// KumaSidecarInjectionAnnotation defines a Pod/Namespace annotation that
	// gives users an ability to enable or disable sidecar-injection
	KumaSidecarInjectionAnnotation = "kuma.io/sidecar-injection"

	// KumaGatewayAnnotation allows to mark Gateway pod,
	// inbound listeners won't be generated in that case
	KumaGatewayAnnotation = "kuma.io/gateway"

	// KumaIngressAnnotation allows to mark pod with Kuma Ingress
	// which is crucial for Multicluster communication
	KumaIngressAnnotation = "kuma.io/ingress"

	// KumaDirectAccess defines a comma-separated list of Services that will be accessed directly
	KumaDirectAccess = "kuma.io/direct-access-services"

	// KumaVirtualProbesAnnotation enables automatic converting HttpGet probes to virtual. Virtual probe
	// serves on sub-path of insecure port defined in 'KumaVirtualProbesPortAnnotation',
	// i.e :8080/health/readiness -> :9000/8080/health/readiness where 9000 is a value of'KumaVirtualProbesPortAnnotation'
	KumaVirtualProbesAnnotation = "kuma.io/virtual-probes"

	// KumaVirtualProbesPortAnnotation is an insecure port for listening virtual probes
	KumaVirtualProbesPortAnnotation = "kuma.io/virtual-probes-port"

	// KumaMetricsPrometheusPort allows to override `Mesh`-wide default port
	KumaMetricsPrometheusPort = "prometheus.metrics.kuma.io/port"

	// KumaMetricsPrometheusPath to override `Mesh`-wide default path
	KumaMetricsPrometheusPath = "prometheus.metrics.kuma.io/path"
)

// Annotations that are being automatically set by the Kuma Sidecar Injector.
const (
	KumaSidecarInjectedAnnotation                 = "kuma.io/sidecar-injected"
	KumaTransparentProxyingAnnotation             = "kuma.io/transparent-proxying"
	KumaTransparentProxyingInboundPortAnnotation  = "kuma.io/transparent-proxying-inbound-port"
	KumaTransparentProxyingOutboundPortAnnotation = "kuma.io/transparent-proxying-outbound-port"
	CNCFNetworkAnnotation                         = "k8s.v1.cni.cncf.io/networks"
	KumaCNI                                       = "kuma-cni"
)

const (
	AnnotationEnabled  = "enabled"
	AnnotationDisabled = "disabled"
)

type Annotations map[string]string

func (a Annotations) GetEnabled(key string) (bool, bool, error) {
	value, ok := a[key]
	if !ok {
		return false, false, nil
	}
	switch value {
	case AnnotationEnabled:
		return true, true, nil
	case AnnotationDisabled:
		return false, true, nil
	default:
		return false, true, errors.Errorf("annotation \"%s\" has wrong value \"%s\", available values are: \"enabled\", \"disabled\"", key, value)
	}
}

func (a Annotations) GetUint32(key string) (uint32, bool, error) {
	value, ok := a[key]
	if !ok {
		return 0, false, nil
	}
	u, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, true, err
	}
	return uint32(u), true, nil
}

func (a Annotations) GetString(key string) (string, bool) {
	value, ok := a[key]
	if !ok {
		return "", false
	}
	return value, true
}

func (a Annotations) GetBool(key string) (bool, bool, error) {
	value, ok := a[key]
	if !ok {
		return false, false, nil
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false, false, err
	}
	return b, true, nil
}
