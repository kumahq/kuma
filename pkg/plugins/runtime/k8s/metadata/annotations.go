package metadata

import (
	"strconv"
	"strings"

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
	// inbound listeners won't be generated in that case.
	// It can be used to mark a pod as providing a builtin gateway.
	KumaGatewayAnnotation = "kuma.io/gateway"

	// KumaIngressAnnotation allows to mark pod with Kuma Ingress
	// which is crucial for Multizone communication
	KumaIngressAnnotation = "kuma.io/ingress"

	// KumaTagsAnnotation holds a JSON representation of desired tags
	KumaTagsAnnotation = "kuma.io/tags"

	// KumaIngressPublicAddressAnnotation allows to pick public address for Ingress
	// If not defined, Kuma will try to pick this address from the Ingress Service
	KumaIngressPublicAddressAnnotation = "kuma.io/ingress-public-address"

	// KumaIngressPublicPortAnnotation allows to pick public port for Ingress
	// If not defined, Kuma will try to pick this address from the Ingress Service
	KumaIngressPublicPortAnnotation = "kuma.io/ingress-public-port"

	// KumaDirectAccess defines a comma-separated list of Services that will be accessed directly
	KumaDirectAccess = "kuma.io/direct-access-services"

	// KumaVirtualProbesAnnotation enables automatic converting HttpGet probes to virtual. Virtual probe
	// serves on sub-path of insecure port defined in 'KumaVirtualProbesPortAnnotation',
	// i.e :8080/health/readiness -> :9000/8080/health/readiness where 9000 is a value of'KumaVirtualProbesPortAnnotation'
	KumaVirtualProbesAnnotation = "kuma.io/virtual-probes"

	// KumaVirtualProbesPortAnnotation is an insecure port for listening virtual probes
	KumaVirtualProbesPortAnnotation = "kuma.io/virtual-probes-port"

	// KumaSidecarEnvVarsAnnotation is a ; separated list of env vars that will be applied on Kuma Sidecar
	// Example value: TEST1=1;TEST2=2
	KumaSidecarEnvVarsAnnotation = "kuma.io/sidecar-env-vars"

	// KumaSidecarConcurrencyAnnotation is an integer value that explicitly sets the Envoy proxy concurrency
	// in the Kuma sidecar. Setting this annotation overrides the default injection behavior of deriving the
	// concurrency from the sidecar container resource limits. A value of 0 tells Envoy to try to use all the
	// visible CPUs.
	KumaSidecarConcurrencyAnnotation = "kuma.io/sidecar-proxy-concurrency"

	// KumaMetricsPrometheusPort allows to override `Mesh`-wide default port
	KumaMetricsPrometheusPort = "prometheus.metrics.kuma.io/port"

	// KumaMetricsPrometheusPath to override `Mesh`-wide default path
	KumaMetricsPrometheusPath = "prometheus.metrics.kuma.io/path"

	// KumaBuiltinDNS the sidecar will use its builtin DNS
	KumaBuiltinDNS     = "kuma.io/builtindns"
	KumaBuiltinDNSPort = "kuma.io/builtindnsport"

	KumaTrafficExcludeInboundPorts  = "traffic.kuma.io/exclude-inbound-ports"
	KumaTrafficExcludeOutboundPorts = "traffic.kuma.io/exclude-outbound-ports"
)

// Annotations that are being automatically set by the Kuma Sidecar Injector.
const (
	KumaSidecarInjectedAnnotation                  = "kuma.io/sidecar-injected"
	KumaIgnoreAnnotation                           = "kuma.io/ignore"
	KumaSidecarUID                                 = "kuma.io/sidecar-uid"
	KumaTransparentProxyingAnnotation              = "kuma.io/transparent-proxying"
	KumaTransparentProxyingInboundPortAnnotation   = "kuma.io/transparent-proxying-inbound-port"
	KumaTransparentProxyingInboundPortAnnotationV6 = "kuma.io/transparent-proxying-inbound-v6-port"
	KumaTransparentProxyingOutboundPortAnnotation  = "kuma.io/transparent-proxying-outbound-port"
	CNCFNetworkAnnotation                          = "k8s.v1.cni.cncf.io/networks"
	KumaCNI                                        = "kuma-cni"
)

// Annotations related to the gateway
const (
	IngressServiceUpstream = "ingress.kubernetes.io/service-upstream"
)

const (
	// Used with the KumaGatewayAnnotation to mark a pod as providing a builtin
	// gateway.
	AnnotationBuiltin = "builtin"
)

const (
	AnnotationEnabled  = "enabled"
	AnnotationDisabled = "disabled"
	AnnotationTrue     = "true"
	AnnotationFalse    = "false"
)

type Annotations map[string]string

func (a Annotations) GetEnabled(key string) (bool, bool, error) {
	value, ok := a[key]
	if !ok {
		return false, false, nil
	}
	switch value {
	case AnnotationEnabled, AnnotationTrue:
		return true, true, nil
	case AnnotationDisabled, AnnotationFalse:
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
		return 0, true, errors.Errorf("failed to parse annotation %q: %s", key, err.Error())
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
		return false, false, errors.Errorf("failed to parse annotation %q: %s", key, err.Error())
	}
	return b, true, nil
}

// GetMap returns map from annotation. Example: "kuma.io/sidecar-env-vars: TEST1=1;TEST2=2"
func (a Annotations) GetMap(key string) (map[string]string, error) {
	value, ok := a[key]
	if !ok {
		return nil, nil
	}
	result := map[string]string{}

	pairs := strings.Split(value, ";")
	for _, pair := range pairs {
		kvSplit := strings.Split(pair, "=")
		if len(kvSplit) != 2 {
			return nil, errors.Errorf("invalid format. Map in %q has to be provided in the following format: key1=value1;key2=value2", key)
		}
		result[kvSplit[0]] = kvSplit[1]
	}
	return result, nil
}
