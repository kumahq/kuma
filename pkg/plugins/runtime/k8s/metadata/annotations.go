package metadata

import (
	"fmt"
	"maps"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Annotations that can be used by the end users.
const (
	// KumaSidecarInjectionAnnotation defines the label that enables or disables
	// sidecar injection on Pods and Namespaces.
	KumaSidecarInjectionAnnotation = "kuma.io/sidecar-injection"

	// KumaGatewayAnnotation allows to mark Gateway pod,
	// inbound listeners won't be generated in that case.
	KumaGatewayAnnotation = "kuma.io/gateway"

	// KumaTagsAnnotation holds a JSON representation of desired tags
	KumaTagsAnnotation = "kuma.io/tags"

	// KumaDirectAccess defines a comma-separated list of Services that will be accessed directly
	KumaDirectAccess = "kuma.io/direct-access-services"

	// KumaApplicationProbeProxyPortAnnotation is a port for proxying application probes
	KumaApplicationProbeProxyPortAnnotation = "kuma.io/application-probe-proxy-port"

	// KumaSidecarEnvVarsAnnotation is a ; separated list of env vars that will be applied on Kuma Sidecar
	// Example value: TEST1=1;TEST2=2
	KumaSidecarEnvVarsAnnotation = "kuma.io/sidecar-env-vars"

	// KumaSidecarConcurrencyAnnotation is an integer value that explicitly sets the Envoy proxy concurrency
	// in the Kuma sidecar. Setting this annotation overrides the default injection behavior of deriving the
	// concurrency from the sidecar container resource limits. A value of 0 tells Envoy to try to use all the
	// visible CPUs.
	KumaSidecarConcurrencyAnnotation = "kuma.io/sidecar-proxy-concurrency"

	// KumaTrafficTransparentProxyConfig is an annotation used to pass a YAML with the transparent proxy
	// configuration in CNI mode, allowing the new logic to retrieve the config from the annotation
	// instead of processing the ConfigMap explicitly
	KumaTrafficTransparentProxyConfig = "traffic.kuma.io/transparent-proxy-config"
	// KumaTrafficTransparentProxyConfigMapName is an annotation used to specify the name of the
	// ConfigMap containing the transparent proxy configuration. This allows the configuration to be
	// retrieved by referencing the ConfigMap's name, enabling flexible and dynamic assignment of
	// proxy settings
	KumaTrafficTransparentProxyConfigMapName = "traffic.kuma.io/transparent-proxy-configmap-name"
	KumaTrafficExcludeInboundPorts           = "traffic.kuma.io/exclude-inbound-ports"
	KumaTrafficExcludeOutboundPorts          = "traffic.kuma.io/exclude-outbound-ports"
	KumaTrafficExcludeOutboundPortsForUIDs   = "traffic.kuma.io/exclude-outbound-ports-for-uids"
	KumaTrafficDropInvalidPackets            = "traffic.kuma.io/drop-invalid-packets"
	KumaTrafficIptablesLogs                  = "traffic.kuma.io/iptables-logs"
	KumaTrafficExcludeInboundIPs             = "traffic.kuma.io/exclude-inbound-ips"
	KumaTrafficExcludeOutboundIPs            = "traffic.kuma.io/exclude-outbound-ips"

	// KumaSidecarTokenVolumeAnnotation allows to specify which volume contains the service account token
	KumaSidecarTokenVolumeAnnotation = "kuma.io/service-account-token-volume"

	// KumaSidecarDrainTime allows to specify drain time of Kuma DP sidecar.
	KumaSidecarDrainTime = "kuma.io/sidecar-drain-time"

	// KumaContainerPatches is a comma-separated list of ContainerPatch names to be applied to injected containers on a given workload
	KumaContainerPatches = "kuma.io/container-patches"

	// KumaEnvoyLogLevel allows to control Envoy log level.
	// Available values are: [trace][debug][info][warning|warn][error][critical][off]
	KumaEnvoyLogLevel          = "kuma.io/envoy-log-level"
	KumaEnvoyComponentLogLevel = "kuma.io/envoy-component-log-level"

	// KumaInitFirst allows to specify whether the init container should be prepended or appended to the existing
	// list of init containers
	KumaInitFirst = "kuma.io/init-first"
	// KumaWaitForDataplaneReady allows to specify if the application sidecar should be hold until Envoy is ready
	KumaWaitForDataplaneReady = "kuma.io/wait-for-dataplane-ready"

	// KumaServiceName points to the Service that a MeshService is derived from
	KumaServiceName = "k8s.kuma.io/service-name"

	// HeadlessService is "true" when the Service had ClusterIP: None, otherwise "false"
	HeadlessService = "k8s.kuma.io/is-headless-service"

	// KumaServiceAccount specifies the ServiceAccount associated with the Pod.
	KumaServiceAccount = "k8s.kuma.io/service-account"

	// KumaWorkload specifies the workload identifier associated with the Pod.
	KumaWorkload = "kuma.io/workload"

	// KumaSpireSupport allows injecting Spire-related volumes into a single Pod.
	KumaSpireSupport = "k8s.kuma.io/spire-support"
)

var PodAnnotationDeprecations = []Deprecation{
	{
		Key:     KumaSidecarInjectionAnnotation,
		Message: "WARNING: you are using kuma.io/sidecar-injection as annotation. This is not supported you should use it as a label instead",
	},
}

type Deprecation struct {
	Key     string
	Message string
}

func NewReplaceByDeprecation(old, n string, removed bool) Deprecation {
	msg := fmt.Sprintf("'%s' is being replaced by: '%s'", old, n)
	if removed {
		msg = fmt.Sprintf("'%s' is no longer supported and it will be ignored, use '%s' instead", old, n)
	}
	return Deprecation{
		Key:     old,
		Message: msg,
	}
}

func NewDeprecation(old string, removed bool) Deprecation {
	msg := fmt.Sprintf("'%s' will be removed in a future release", old)
	if removed {
		msg = fmt.Sprintf("'%s' is no longer supported and it will be ignored, please see documentation on how to migrate", old)
	}
	return Deprecation{
		Key:     old,
		Message: msg,
	}
}

// Annotations that are being automatically set by the Kuma Sidecar Injector.
const (
	KumaSidecarInjectedAnnotation       = "kuma.io/sidecar-injected"
	KumaIgnoreAnnotation                = "kuma.io/ignore"
	KumaSidecarUID                      = "kuma.io/sidecar-uid"
	KumaEnvoyAdminPort                  = "kuma.io/envoy-admin-port"
	KumaTransparentProxyingAnnotation   = "kuma.io/transparent-proxying"
	KumaTransparentProxyingIPFamilyMode = "kuma.io/transparent-proxying-ip-family-mode"
	KumaReachableBackends               = "kuma.io/reachable-backends"
	CNCFNetworkAnnotation               = "k8s.v1.cni.cncf.io/networks"
	KumaCNI                             = "kuma-cni"
)

// Annotations related to the gateway
const (
	IngressServiceUpstream      = "ingress.kubernetes.io/service-upstream"
	NginxIngressServiceUpstream = "nginx.ingress.kubernetes.io/service-upstream"
)

const (
	AnnotationEnabled  = "enabled"
	AnnotationDisabled = "disabled"
	AnnotationTrue     = "true"
	AnnotationFalse    = "false"
	AnnotationYes      = "yes"
	AnnotationNo       = "no"
)

// these values are defined for users to specify in configuration:
// values comes from mesh_proto.Dataplane_Networking_TransparentProxying_IpFamilyMode_name
const (
	IpFamilyModeDualStack = "dualstack"
	IpFamilyModeIPv4      = "ipv4"
	IpFamilyModeIPv6      = "ipv6"
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

func (a Annotations) GetBoolean(keys ...string) (bool, bool, error) {
	return a.GetBooleanWithDefault(false, false, keys...)
}

func (a Annotations) GetEnabledWithDefault(def bool, keys ...string) (bool, bool, error) {
	return a.GetBooleanWithDefault(def, true, keys...)
}

func (a Annotations) GetBooleanWithDefault(def bool, supportEnabled bool, keys ...string) (bool, bool, error) {
	v, exists, err := a.getWithDefault(def, func(key, value string) (any, error) {
		switch value {
		case AnnotationTrue, AnnotationYes:
			return true, nil
		case AnnotationFalse, AnnotationNo:
			return false, nil
		default:
			if supportEnabled {
				switch value {
				case AnnotationEnabled:
					return true, nil
				case AnnotationDisabled:
					return false, nil
				}
			}
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
	v, exists, err := a.getWithDefault(def, func(key string, value string) (any, error) {
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
	v, exists, _ := a.getWithDefault(def, func(key string, value string) (any, error) {
		return value, nil
	}, keys...)
	return v.(string), exists
}

func (a Annotations) GetDurationWithDefault(def time.Duration, keys ...string) (time.Duration, bool, error) {
	v, exists, err := a.getWithDefault(def, func(key string, value string) (any, error) {
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
	v, exists, _ := a.getWithDefault(defCopy, func(key string, value string) (any, error) {
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
	maps.Copy(defCopy, def)
	v, exists, err := a.getWithDefault(defCopy, func(key string, value string) (any, error) {
		result := map[string]string{}

		pairs := strings.SplitSeq(value, ";")
		for pair := range pairs {
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

func (a Annotations) getWithDefault(def any, fn func(string, string) (any, error), keys ...string) (any, bool, error) {
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
