package metadata

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	// KumaEgressAnnotation allows marking pod with Kuma Egress
	// which is crucial for Multizone communication
	KumaEgressAnnotation = "kuma.io/egress"

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

	// Remove with: https://github.com/kumahq/kuma/issues/4675
	KumaBuiltinDNSDeprecated     = "kuma.io/builtindns"
	KumaBuiltinDNSPortDeprecated = "kuma.io/builtindnsport"
	// KumaBuiltinDNS the sidecar will use its builtin DNS
	KumaBuiltinDNS     = "kuma.io/builtin-dns"
	KumaBuiltinDNSPort = "kuma.io/builtin-dns-port"

	KumaTrafficExcludeInboundPorts            = "traffic.kuma.io/exclude-inbound-ports"
	KumaTrafficExcludeOutboundPorts           = "traffic.kuma.io/exclude-outbound-ports"
	KumaTrafficExcludeOutboundTCPPortsForUIDs = "traffic.kuma.io/exclude-outbound-tcp-ports-for-uids"
	KumaTrafficExcludeOutboundUDPPortsForUIDs = "traffic.kuma.io/exclude-outbound-udp-ports-for-uids"

	// KumaSidecarTokenVolumeAnnotation allows to specify which volume contains the service account token
	KumaSidecarTokenVolumeAnnotation = "kuma.io/service-account-token-volume"

	// KumaTransparentProxyingExperimentalEngine enables experimental transparent proxy engine.
	KumaTransparentProxyingExperimentalEngine = "kuma.io/transparent-proxying-experimental-engine"

	// KumaSidecarDrainTime allows to specify drain time of Kuma DP sidecar.
	KumaSidecarDrainTime = "kuma.io/sidecar-drain-time"

	// KumaContainerPatches is a comma-separated list of ContainerPatch names to be applied to injected containers on a given workload
	KumaContainerPatches = "kuma.io/container-patches"

	// KumaEnvoyLogLevel allows to control Envoy log level.
	// Available values are: [trace][debug][info][warning|warn][error][critical][off]
	KumaEnvoyLogLevel = "kuma.io/envoy-log-level"

	// KumaMetricsPrometheusAggregatePath allows to specify which path for specific app should request for metrics
	KumaMetricsPrometheusAggregatePath = "prometheus.metrics.kuma.io/aggregate-%s-path"
	// KumaMetricsPrometheusAggregatePort allows to specify which port for specific app should request for metrics
	KumaMetricsPrometheusAggregatePort = "prometheus.metrics.kuma.io/aggregate-%s-port"
	// KumaMetricsPrometheusAggregateEnabled allows to specify if we want to enable specific scraping, default: true
	KumaMetricsPrometheusAggregateEnabled = "prometheus.metrics.kuma.io/aggregate-%s-enabled"
	// KumaMetricsPrometheusAggregatePattern allows to retrieve all the apps for which need to get port/path configuration
	KumaMetricsPrometheusAggregatePattern = "^prometheus\\.metrics\\.kuma\\.io/aggregate-([a-zA-Z0-9-]+)-(port|path|enabled)$"
)

var PodAnnotationDeprecations = []Deprecation{
	NewReplaceByDeprecation(KumaBuiltinDNSDeprecated, KumaBuiltinDNS),
	NewReplaceByDeprecation(KumaBuiltinDNSPortDeprecated, KumaBuiltinDNSPort),
	{
		Key:     KumaSidecarInjectionAnnotation,
		Message: "WARNING: you are using kuma.io/sidecar-injection as annotation. Please migrate it to label to have strong guarantee that application can only start with sidecar",
	},
}

type Deprecation struct {
	Key     string
	Message string
}

func NewReplaceByDeprecation(old, new string) Deprecation {
	return Deprecation{
		Key:     old,
		Message: fmt.Sprintf("'%s' is being replaced by: '%s'", old, new),
	}
}

// Annotations that are being automatically set by the Kuma Sidecar Injector.
const (
	KumaSidecarInjectedAnnotation                      = "kuma.io/sidecar-injected"
	KumaIgnoreAnnotation                               = "kuma.io/ignore"
	KumaSidecarUID                                     = "kuma.io/sidecar-uid"
	KumaEnvoyAdminPort                                 = "kuma.io/envoy-admin-port"
	KumaTransparentProxyingAnnotation                  = "kuma.io/transparent-proxying"
	KumaTransparentProxyingInboundPortAnnotation       = "kuma.io/transparent-proxying-inbound-port"
	KumaTransparentProxyingInboundPortAnnotationV6     = "kuma.io/transparent-proxying-inbound-v6-port"
	KumaTransparentProxyingOutboundPortAnnotation      = "kuma.io/transparent-proxying-outbound-port"
	KumaTransparentProxyingReachableServicesAnnotation = "kuma.io/transparent-proxying-reachable-services"
	CNCFNetworkAnnotation                              = "k8s.v1.cni.cncf.io/networks"
	KumaCNI                                            = "kuma-cni"
	KumaTransparentProxyingEbpf                        = "kuma.io/transparent-proxying-ebpf"
	KumaTransparentProxyingEbpfBPFFSPath               = "kuma.io/transparent-proxying-ebpf-bpf-fs-path"
	KumaTransparentProxyingEbpfCgroupPath              = "kuma.io/transparent-proxying-ebpf-cgroup-path"
	KumaTransparentProxyingEbpfTCAttachIface           = "kuma.io/transparent-proxying-ebpf-tc-attach-iface"
	KumaTransparentProxyingEbpfInstanceIPEnvVarName    = "kuma.io/transparent-proxying-ebpf-instance-ip-env-var-name"
	KumaTransparentProxyingEbpfProgramsSourcePath      = "kuma.io/transparent-proxying-ebpf-programs-source-path"
)

// Annotations related to the gateway
const (
	IngressServiceUpstream      = "ingress.kubernetes.io/service-upstream"
	NginxIngressServiceUpstream = "nginx.ingress.kubernetes.io/service-upstream"
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

func BoolToEnabled(b bool) string {
	if b {
		return AnnotationEnabled
	}

	return AnnotationDisabled
}

type Annotations map[string]string

func (a Annotations) GetEnabled(keys ...string) (bool, bool, error) {
	return a.GetEnabledWithDefault(false, keys...)
}
func (a Annotations) GetEnabledWithDefault(def bool, keys ...string) (bool, bool, error) {
	v, exists, err := a.getWithDefault(def, func(key, value string) (interface{}, error) {
		switch value {
		case AnnotationEnabled, AnnotationTrue:
			return true, nil
		case AnnotationDisabled, AnnotationFalse:
			return false, nil
		default:
			return false, errors.Errorf("annotation \"%s\" has wrong value \"%s\"", key, value)
		}
	}, keys...)
	if err != nil {
		return def, exists, err
	}
	return v.(bool), exists, nil
}

func (a Annotations) GetUint32(keys ...string) (uint32, bool, error) {
	return a.GetUint32WithDefault(0, keys...)
}
func (a Annotations) GetUint32WithDefault(def uint32, keys ...string) (uint32, bool, error) {
	v, exists, err := a.getWithDefault(def, func(key string, value string) (interface{}, error) {
		u, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return 0, errors.Errorf("failed to parse annotation %q: %s", key, err.Error())
		}
		return uint32(u), nil
	}, keys...)
	if err != nil {
		return def, exists, err
	}
	return v.(uint32), exists, nil
}

func (a Annotations) GetString(keys ...string) (string, bool) {
	return a.GetStringWithDefault("", keys...)
}
func (a Annotations) GetStringWithDefault(def string, keys ...string) (string, bool) {
	v, exists, _ := a.getWithDefault(def, func(key string, value string) (interface{}, error) {
		return value, nil
	}, keys...)
	return v.(string), exists
}

func (a Annotations) GetDurationWithDefault(def time.Duration, keys ...string) (time.Duration, bool, error) {
	v, exists, err := a.getWithDefault(def, func(key string, value string) (interface{}, error) {
		return time.ParseDuration(value)
	}, keys...)
	if err != nil {
		return def, exists, err
	}
	return v.(time.Duration), exists, err
}

func (a Annotations) GetList(keys ...string) ([]string, bool) {
	return a.GetListWithDefault(nil, keys...)
}
func (a Annotations) GetListWithDefault(def []string, keys ...string) ([]string, bool) {
	defCopy := []string{}
	defCopy = append(defCopy, def...)
	v, exists, _ := a.getWithDefault(defCopy, func(key string, value string) (interface{}, error) {
		r := strings.Split(value, ",")
		var res []string
		for _, v := range r {
			if v != "" {
				res = append(res, v)
			}
		}
		return res, nil
	}, keys...)
	return v.([]string), exists
}

// GetMap returns map from annotation. Example: "kuma.io/sidecar-env-vars: TEST1=1;TEST2=2"
func (a Annotations) GetMap(keys ...string) (map[string]string, bool, error) {
	return a.GetMapWithDefault(map[string]string{}, keys...)
}
func (a Annotations) GetMapWithDefault(def map[string]string, keys ...string) (map[string]string, bool, error) {
	defCopy := make(map[string]string, len(def))
	for k, v := range def {
		defCopy[k] = v
	}
	v, exists, err := a.getWithDefault(defCopy, func(key string, value string) (interface{}, error) {
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
	}, keys...)
	if err != nil {
		return def, exists, err
	}
	return v.(map[string]string), exists, nil
}

func (a Annotations) getWithDefault(def interface{}, fn func(string, string) (interface{}, error), keys ...string) (interface{}, bool, error) {
	res := def
	exists := false
	for _, k := range keys {
		v, ok := a[k]
		if ok {
			exists = true
			r, err := fn(k, v)
			if err != nil {
				return nil, exists, err
			}
			res = r
		}
	}
	return res, exists, nil
}
