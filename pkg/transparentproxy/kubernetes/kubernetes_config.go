package kubernetes

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	core_config "github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/config/plugins/runtime/k8s"
	k8s_metadata "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	k8s_probes "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
	tproxy_config "github.com/kumahq/kuma/pkg/transparentproxy/config"
)

const (
	warningTProxyConfigMismatch              = "[WARNING]: provided value in the transparent proxy configuration ConfigMap cannot be modified independently and must align with the value specified in the runtime configuration"
	warningAnnotationValueMismatch           = "[WARNING]: annotation cannot be modified directly and must be consistent with the value specified in the runtime configuration"
	warningInvalidAnnotationAndValueMismatch = warningAnnotationValueMismatch + "additionally, provided value is invalid"
	warningDisablingTProxyAnnotation         = "[WARNING]: disabling transparent proxying is not supported in Kubernetes environments; the annotation will be ignored"
)

type trafficKind string

const (
	trafficKindOutbound trafficKind = "outbound"
	trafficKindInbound  trafficKind = "inbound"
	trafficKindDNS      trafficKind = "dns"
)

func (c trafficKind) capitalize() string {
	if c == "" {
		return ""
	}

	return strings.ToUpper(string(c)[:1]) + string(c)[1:]
}

type configurer struct {
	logger      logr.Logger
	config      tproxy_config.Config
	runtime     k8s.Injector
	annotations k8s_metadata.Annotations
}

func (c configurer) kumaDPUser() string {
	annotation := k8s_metadata.KumaSidecarUID

	valueDefault := tproxy_config.DefaultConfig().KumaDPUser
	valueCurrent := c.config.KumaDPUser
	valueRuntime := fmt.Sprintf("%d", c.runtime.SidecarContainer.UID)

	pathConfigMap := c.pathConfigMapf("kumaDPUser")
	pathRuntime := c.pathRuntimef("sidecarContainer.uid")

	l := c.logger.WithValues(
		"default", valueDefault,
		pathConfigMap, valueCurrent,
		pathRuntime, valueRuntime,
	)

	if valueCurrent != valueRuntime && valueCurrent != valueDefault {
		l.Info(warningTProxyConfigMismatch)
	}

	if v, exists := c.annotations.GetString(annotation); exists && v != valueRuntime {
		l.Info(warningAnnotationValueMismatch)
	}

	return valueRuntime
}

func (c configurer) redirectPort(kind trafficKind) tproxy_config.Port {
	var annotation string

	var valueDefault tproxy_config.Port
	var valueCurrent tproxy_config.Port
	var valueRuntime tproxy_config.Port

	pathRuntime := c.pathRuntimef("sidecarContainer.redirectPort%s", kind.capitalize())

	switch kind {
	case trafficKindOutbound:
		annotation = k8s_metadata.KumaTransparentProxyingOutboundPortAnnotation
		valueDefault = tproxy_config.DefaultConfig().Redirect.Outbound.Port
		valueCurrent = c.config.Redirect.Outbound.Port
		valueRuntime = tproxy_config.Port(c.runtime.SidecarContainer.RedirectPortOutbound)
	case trafficKindInbound:
		annotation = k8s_metadata.KumaTransparentProxyingInboundPortAnnotation
		valueDefault = tproxy_config.DefaultConfig().Redirect.Inbound.Port
		valueCurrent = c.config.Redirect.Inbound.Port
		valueRuntime = tproxy_config.Port(c.runtime.SidecarContainer.RedirectPortInbound)
	case trafficKindDNS:
		annotation = k8s_metadata.KumaBuiltinDNSPort
		valueDefault = tproxy_config.DefaultConfig().Redirect.DNS.Port
		valueCurrent = c.config.Redirect.DNS.Port
		valueRuntime = tproxy_config.Port(c.runtime.BuiltinDNS.Port)
		pathRuntime = c.pathRuntimef("builtinDNS.port")
	default:
		return 0
	}

	pathConfigMap := c.pathConfigMapf("redirect.%s.port", kind)

	l := c.logger.WithValues(
		"default", valueDefault,
		pathConfigMap, valueCurrent,
		pathRuntime, valueRuntime,
	)

	if valueCurrent != valueRuntime && valueCurrent != valueDefault {
		l.Info(warningTProxyConfigMismatch)
	}

	if v, exists, err := c.annotations.GetUint32(annotation); err != nil {
		l.Info(warningInvalidAnnotationAndValueMismatch, annotation, err)
	} else if exists && tproxy_config.Port(v) != valueRuntime {
		l.Info(warningAnnotationValueMismatch, annotation, v)
	}

	return valueRuntime
}

func (c configurer) redirectExcludePorts(kind trafficKind) (tproxy_config.Ports, error) {
	var annotation string

	var valueDefault tproxy_config.Ports
	var valueRuntime tproxy_config.Ports
	var valueCurrent tproxy_config.Ports
	var valueAnnotation tproxy_config.Ports

	var err error

	switch kind {
	case trafficKindOutbound:
		annotation = k8s_metadata.KumaTrafficExcludeOutboundPorts
		valueDefault = tproxy_config.DefaultConfig().Redirect.Outbound.ExcludePorts
		runtimeExcludePorts := c.runtime.SidecarTraffic.ExcludeOutboundPorts
		if valueRuntime, err = convertToPorts(runtimeExcludePorts); err != nil {
			return tproxy_config.Ports{}, err
		}
		valueCurrent = c.config.Redirect.Outbound.ExcludePorts
	case trafficKindInbound:
		annotation = k8s_metadata.KumaTrafficExcludeInboundPorts
		valueDefault = tproxy_config.DefaultConfig().Redirect.Inbound.ExcludePorts
		runtimeExcludePorts := c.runtime.SidecarTraffic.ExcludeInboundPorts
		if valueRuntime, err = convertToPorts(runtimeExcludePorts); err != nil {
			return tproxy_config.Ports{}, err
		}
		valueCurrent = c.config.Redirect.Inbound.ExcludePorts
	default:
		return tproxy_config.Ports{}, nil
	}

	if v, exists := c.annotations.GetString(annotation); exists {
		if valueAnnotation, err = convertToPorts(v); err != nil {
			return tproxy_config.Ports{}, err
		}
	}

	pathRuntime := c.pathRuntimef("sidecarTraffic.exclude%sPorts", kind.capitalize())
	pathConfigMap := c.pathConfigMapf("redirect.%s.excludePorts", kind)

	l := c.logger.WithValues(
		"default", valueDefault,
		pathRuntime, valueRuntime,
		pathConfigMap, valueCurrent,
		annotation, valueAnnotation,
	)

	logDebugUsageAndReturn := func(
		source string,
		value tproxy_config.Ports,
	) (tproxy_config.Ports, error) {
		l.V(1).Info(fmt.Sprintf("using exclude %s ports from the %s", kind, source))
		return value, nil
	}

	switch {
	case len(valueAnnotation) > 0:
		return logDebugUsageAndReturn("annotation", valueAnnotation)
	case len(valueCurrent) > 0:
		return logDebugUsageAndReturn("transparent proxy ConfigMap", valueCurrent)
	case len(valueRuntime) > 0:
		return logDebugUsageAndReturn("runtime configuration", valueRuntime)
	default:
		return logDebugUsageAndReturn("default configuration", valueDefault)
	}
}

func (c configurer) redirectExcludePortsForIPs(kind trafficKind) []string {
	var annotation string

	var valueDefault []string
	var valueRuntime []string
	var valueCurrent []string
	var valueAnnotation []string

	switch kind {
	case trafficKindOutbound:
		annotation = k8s_metadata.KumaTrafficExcludeOutboundIPs
		valueDefault = tproxy_config.DefaultConfig().Redirect.Outbound.ExcludePortsForIPs
		valueRuntime = c.runtime.SidecarTraffic.ExcludeOutboundIPs
		valueCurrent = c.config.Redirect.Outbound.ExcludePortsForIPs
	case trafficKindInbound:
		annotation = k8s_metadata.KumaTrafficExcludeInboundIPs
		valueDefault = tproxy_config.DefaultConfig().Redirect.Inbound.ExcludePortsForIPs
		valueRuntime = c.runtime.SidecarTraffic.ExcludeInboundIPs
		valueCurrent = c.config.Redirect.Inbound.ExcludePortsForIPs
	default:
		return nil
	}

	if v, exists := c.annotations.GetList(annotation); exists {
		valueAnnotation = v
	}

	pathRuntime := c.pathRuntimef("sidecarTraffic.exclude%sIPs", kind.capitalize())
	pathConfigMap := c.pathConfigMapf("redirect.%s.excludePortsForIPs", kind)

	l := c.logger.WithValues(
		"default", valueDefault,
		pathRuntime, valueRuntime,
		pathConfigMap, valueCurrent,
		annotation, valueAnnotation,
	)

	logDebugUsageAndReturn := func(source string, value []string) []string {
		l.V(1).Info(fmt.Sprintf("using exclude %s IPs from the %s", kind, source))
		return value
	}

	switch {
	case len(valueAnnotation) > 0:
		return logDebugUsageAndReturn("annotation", valueAnnotation)
	case len(valueCurrent) > 0:
		return logDebugUsageAndReturn("transparent proxy ConfigMap", valueCurrent)
	case len(valueRuntime) > 0:
		return logDebugUsageAndReturn("runtime configuration", valueRuntime)
	default:
		return logDebugUsageAndReturn("default configuration", valueDefault)
	}
}

func (c configurer) ipFamilyMode() (tproxy_config.IPFamilyMode, error) {
	annotation := k8s_metadata.KumaTransparentProxyingIPFamilyMode

	valueDefault := tproxy_config.DefaultConfig().IPFamilyMode
	valueCurrent := c.config.IPFamilyMode
	valueRuntimeRaw := c.runtime.SidecarContainer.IpFamilyMode

	var valueRuntime tproxy_config.IPFamilyMode
	var valueAnnotation tproxy_config.IPFamilyMode

	if valueRuntimeRaw != "" && valueDefault.String() != valueRuntimeRaw {
		if err := valueRuntime.Set(valueRuntimeRaw); err != nil {
			return "", errors.Wrapf(
				err,
				"invalid IP Family Mode in runtime configuration: %s",
				valueRuntimeRaw,
			)
		}
	}

	if v, exists := c.annotations.GetString(annotation); exists {
		if err := valueAnnotation.Set(v); err != nil {
			return "", errors.Wrapf(
				err,
				"invalid IP Family Mode in annotation '%s': %s",
				annotation,
				v,
			)
		}
	}

	pathConfigMap := c.pathConfigMapf("ipFamilyMode")
	pathRuntime := c.pathRuntimef("ipFamilyMode")

	l := c.logger.WithValues(
		"default", valueDefault,
		pathRuntime, valueRuntime,
		pathConfigMap, valueCurrent,
		annotation, valueAnnotation,
	)

	logDebugUsageAndReturn := func(
		source string,
		value tproxy_config.IPFamilyMode,
	) (tproxy_config.IPFamilyMode, error) {
		l.V(1).Info(fmt.Sprintf("using IP Family Mode from the %s", source))
		return value, nil
	}

	switch {
	case len(valueAnnotation) > 0:
		return logDebugUsageAndReturn("annotation", valueAnnotation)
	case len(valueCurrent) > 0:
		return logDebugUsageAndReturn("transparent proxy ConfigMap", valueCurrent)
	case len(valueRuntime) > 0:
		return logDebugUsageAndReturn("runtime configuration", valueRuntime)
	default:
		return logDebugUsageAndReturn("default configuration", valueDefault)
	}
}

func (c configurer) redirectDNSEnabled() bool {
	annotation := k8s_metadata.KumaBuiltinDNS

	valueCurrent := c.config.Redirect.DNS.Enabled
	valueDefault := tproxy_config.DefaultConfig().Redirect.DNS.Enabled
	valueRuntime := c.runtime.BuiltinDNS.Enabled

	pathConfigMap := c.pathConfigMapf("redirect.dns.enabled")
	pathRuntime := c.pathRuntimef("builtinDNS.enabled")

	l := c.logger.WithValues(
		"default", valueDefault,
		pathConfigMap, valueCurrent,
		pathRuntime, valueRuntime,
	)

	if valueCurrent != valueRuntime && valueCurrent != valueDefault {
		l.Info(warningTProxyConfigMismatch)
	}

	if v, exists, err := c.annotations.GetEnabled(annotation); err != nil {
		l.Info(warningInvalidAnnotationAndValueMismatch, annotation, err)
	} else if exists && v != valueRuntime {
		l.Info(warningAnnotationValueMismatch, annotation, v)
	}

	return valueCurrent
}

func (c configurer) pathConfigMapf(format string, a ...any) string {
	return fmt.Sprintf(
		"configmap/%s/%s",
		c.runtime.TransparentProxyConfigMapName,
		fmt.Sprintf(format, a...),
	)
}

func (c configurer) pathRuntimef(format string, a ...any) string {
	return "runtime.injector." + fmt.Sprintf(format, a...)
}

func ConfigForKubernetes(
	cfg tproxy_config.Config,
	runtimeCfg k8s.Injector,
	annotations k8s_metadata.Annotations,
	logger logr.Logger,
) (tproxy_config.Config, error) {
	l := logger.WithName("transparentproxy.config")

	k8sConfigurer := configurer{
		config:      cfg,
		runtime:     runtimeCfg,
		annotations: annotations,
		logger:      l,
	}

	if v, exists, _ := annotations.GetEnabled(k8s_metadata.KumaTransparentProxyingAnnotation); exists && !v {
		l.Info(
			warningDisablingTProxyAnnotation,
			"annotation", k8s_metadata.KumaTransparentProxyingAnnotation,
		)
	}

	cfg.KumaDPUser = k8sConfigurer.kumaDPUser()
	cfg.CNIMode = runtimeCfg.CNIEnabled

	if v, err := k8sConfigurer.ipFamilyMode(); err != nil {
		return tproxy_config.Config{}, err
	} else {
		cfg.IPFamilyMode = v
	}

	if v, exists, err := annotations.GetBoolean(k8s_metadata.KumaTrafficDropInvalidPackets); err != nil {
		return cfg, err
	} else if exists {
		cfg.DropInvalidPackets = v
	}

	if v, exists, err := annotations.GetBoolean(k8s_metadata.KumaTrafficIptablesLogs); err != nil {
		return cfg, err
	} else if exists {
		cfg.Log.Enabled = v
	}

	cfg.Redirect.DNS.Enabled = k8sConfigurer.redirectDNSEnabled()
	cfg.Redirect.DNS.Port = k8sConfigurer.redirectPort(trafficKindDNS)

	cfg.Redirect.Outbound.Port = k8sConfigurer.redirectPort(trafficKindOutbound)

	if v, err := k8sConfigurer.redirectExcludePorts(trafficKindOutbound); err != nil {
		return cfg, err
	} else {
		cfg.Redirect.Outbound.ExcludePorts = v
	}

	if v, exists := annotations.GetString(k8s_metadata.KumaTrafficExcludeOutboundPortsForUIDs); exists {
		cfg.Redirect.Outbound.ExcludePortsForUIDs = strings.Split(v, ";")
	}

	cfg.Redirect.Outbound.ExcludePortsForIPs = k8sConfigurer.redirectExcludePortsForIPs(trafficKindOutbound)

	if v, exists, err := annotations.GetEnabled(k8s_metadata.KumaGatewayAnnotation); err != nil {
		return cfg, err
	} else if exists && v {
		cfg.Redirect.Inbound.Enabled = false
	}

	if cfg.Redirect.Inbound.Enabled {
		cfg.Redirect.Inbound.Port = k8sConfigurer.redirectPort(trafficKindInbound)
		cfg.Redirect.Inbound.ExcludePortsForIPs = k8sConfigurer.redirectExcludePortsForIPs(trafficKindInbound)

		if v, err := k8sConfigurer.redirectExcludePorts(trafficKindInbound); err != nil {
			return cfg, err
		} else {
			cfg.Redirect.Inbound.ExcludePorts = v
		}

		if v, err := k8s_probes.GetApplicationProbeProxyPort(
			annotations,
			runtimeCfg.ApplicationProbeProxyPort,
		); err != nil {
			return cfg, err
		} else if v != 0 {
			if err := cfg.Redirect.Inbound.ExcludePorts.Append(fmt.Sprintf("%d", v)); err != nil {
				return cfg, err
			}
		}
	}

	if v, exists, err := annotations.GetEnabled(k8s_metadata.KumaTransparentProxyingEbpf); err != nil {
		return cfg, err
	} else if exists {
		cfg.Ebpf.Enabled = v
	}

	if cfg.Ebpf.Enabled {
		if v, _ := annotations.GetStringWithDefault(
			runtimeCfg.EBPF.BPFFSPath,
			k8s_metadata.KumaTransparentProxyingEbpfBPFFSPath,
		); v != "" {
			cfg.Ebpf.BPFFSPath = v
		}

		if v, _ := annotations.GetStringWithDefault(
			runtimeCfg.EBPF.CgroupPath,
			k8s_metadata.KumaTransparentProxyingEbpfCgroupPath,
		); v != "" {
			cfg.Ebpf.CgroupPath = v
		}

		if v, _ := annotations.GetStringWithDefault(
			runtimeCfg.EBPF.TCAttachIface,
			k8s_metadata.KumaTransparentProxyingEbpfTCAttachIface,
		); v != "" {
			cfg.Ebpf.TCAttachIface = v
		}

		if v, _ := annotations.GetStringWithDefault(
			runtimeCfg.EBPF.InstanceIPEnvVarName,
			k8s_metadata.KumaTransparentProxyingEbpfInstanceIPEnvVarName,
		); v != "" {
			cfg.Ebpf.InstanceIPEnvVarName = v
		}

		if v, _ := annotations.GetStringWithDefault(
			runtimeCfg.EBPF.ProgramsSourcePath,
			k8s_metadata.KumaTransparentProxyingEbpfProgramsSourcePath,
		); v != "" {
			cfg.Ebpf.ProgramsSourcePath = v
		}
	}

	return cfg, nil
}

func ConfigToAnnotations(
	cfg tproxy_config.Config,
	runtimeCfg k8s.Injector,
	annotations map[string]string,
	mesh string,
	defaultAdminPort uint32,
) (map[string]string, error) {
	result := map[string]string{
		k8s_metadata.KumaSidecarInjectedAnnotation:                 k8s_metadata.AnnotationTrue,
		k8s_metadata.KumaTransparentProxyingAnnotation:             k8s_metadata.AnnotationEnabled,
		k8s_metadata.KumaMeshAnnotation:                            mesh, // either user-defined value or default
		k8s_metadata.KumaSidecarUID:                                cfg.KumaDPUser,
		k8s_metadata.KumaTransparentProxyingOutboundPortAnnotation: cfg.Redirect.Outbound.Port.String(),
		k8s_metadata.KumaTransparentProxyingInboundPortAnnotation:  cfg.Redirect.Inbound.Port.String(),
		k8s_metadata.KumaTransparentProxyingIPFamilyMode:           cfg.IPFamilyMode.String(),
	}

	if cfg.CNIMode {
		result[k8s_metadata.CNCFNetworkAnnotation] = k8s_metadata.KumaCNI
	}

	if cfg.DropInvalidPackets {
		result[k8s_metadata.KumaTrafficDropInvalidPackets] = k8s_metadata.AnnotationTrue
	}

	if cfg.Log.Enabled {
		result[k8s_metadata.KumaTrafficIptablesLogs] = k8s_metadata.AnnotationTrue
	}

	if cfg.Ebpf.Enabled {
		result[k8s_metadata.KumaTransparentProxyingEbpf] = k8s_metadata.AnnotationEnabled
		result[k8s_metadata.KumaTransparentProxyingEbpfBPFFSPath] = cfg.Ebpf.BPFFSPath
		result[k8s_metadata.KumaTransparentProxyingEbpfCgroupPath] = cfg.Ebpf.CgroupPath
		result[k8s_metadata.KumaTransparentProxyingEbpfTCAttachIface] = cfg.Ebpf.TCAttachIface
		result[k8s_metadata.KumaTransparentProxyingEbpfProgramsSourcePath] = cfg.Ebpf.ProgramsSourcePath
		result[k8s_metadata.KumaTransparentProxyingEbpfInstanceIPEnvVarName] = cfg.Ebpf.InstanceIPEnvVarName
	}

	if len(cfg.Redirect.Outbound.ExcludePorts) > 0 {
		result[k8s_metadata.KumaTrafficExcludeOutboundPorts] = cfg.Redirect.Outbound.ExcludePorts.String()
	}

	if len(cfg.Redirect.Outbound.ExcludePortsForIPs) > 0 {
		result[k8s_metadata.KumaTrafficExcludeOutboundIPs] = strings.Join(cfg.Redirect.Outbound.ExcludePortsForIPs, ",")
	}

	if len(cfg.Redirect.Inbound.ExcludePorts) > 0 {
		result[k8s_metadata.KumaTrafficExcludeInboundPorts] = cfg.Redirect.Inbound.ExcludePorts.String()
	}

	if len(cfg.Redirect.Inbound.ExcludePortsForIPs) > 0 {
		result[k8s_metadata.KumaTrafficExcludeInboundIPs] = strings.Join(cfg.Redirect.Inbound.ExcludePortsForIPs, ",")
	}

	if cfg.Redirect.DNS.Enabled {
		result[k8s_metadata.KumaBuiltinDNS] = k8s_metadata.AnnotationEnabled
		result[k8s_metadata.KumaBuiltinDNSPort] = cfg.Redirect.DNS.Port.String()

		if v, _, err := k8s_metadata.Annotations(result).GetEnabledWithDefault(
			runtimeCfg.BuiltinDNS.Logging,
			k8s_metadata.KumaBuiltinDNSLogging,
		); err != nil {
			return nil, err
		} else {
			result[k8s_metadata.KumaBuiltinDNSLogging] = strconv.FormatBool(v)
		}
	}

	if err := k8s_probes.SetVirtualProbesEnabledAnnotation(
		result,
		annotations,
		runtimeCfg.VirtualProbesEnabled,
	); err != nil {
		return nil, errors.Wrapf(err, "unable to set %s", k8s_metadata.KumaVirtualProbesAnnotation)
	}

	if v, _, err := k8s_metadata.Annotations(result).GetUint32WithDefault(
		runtimeCfg.VirtualProbesPort,
		k8s_metadata.KumaVirtualProbesPortAnnotation,
	); err != nil {
		return nil, errors.Wrapf(err, "unable to set %s", k8s_metadata.KumaVirtualProbesPortAnnotation)
	} else {
		result[k8s_metadata.KumaVirtualProbesPortAnnotation] = fmt.Sprintf("%d", v)
	}

	if v, err := k8s_probes.GetApplicationProbeProxyPort(
		annotations,
		runtimeCfg.ApplicationProbeProxyPort,
	); err != nil {
		return nil, errors.Wrapf(err, "unable to set %s", k8s_metadata.KumaApplicationProbeProxyPortAnnotation)
	} else {
		result[k8s_metadata.KumaApplicationProbeProxyPortAnnotation] = fmt.Sprintf("%d", v)
	}

	if v, _, err := k8s_metadata.Annotations(result).GetUint32WithDefault(
		defaultAdminPort,
		k8s_metadata.KumaEnvoyAdminPort,
	); err != nil {
		return nil, err
	} else {
		result[k8s_metadata.KumaEnvoyAdminPort] = fmt.Sprintf("%d", v)
	}

	return result, nil
}

func ConfigFromAnnotations(annotations k8s_metadata.Annotations) (tproxy_config.Config, error) {
	annotation := k8s_metadata.KumaTrafficTransparentProxyConfig

	if cfgYAML, ok := annotations.GetString(annotation); ok && cfgYAML != "" {
		cfg := tproxy_config.DefaultConfig()

		if err := core_config.NewLoader(&cfg).LoadBytes([]byte(cfgYAML)); err != nil {
			return tproxy_config.Config{}, errors.Wrapf(
				err,
				"failed to load transparent proxy configuration from '%s' annotation",
				annotation,
			)
		}

		return cfg, nil
	}

	return tproxy_config.Config{}, errors.Errorf(
		"annotation '%s' is either missing or does not contain a valid transparent proxy configuration",
		annotation,
	)
}

func convertUint32SliceToPorts(values []uint32) (tproxy_config.Ports, error) {
	result := tproxy_config.Ports{}
	portStrings := make([]string, len(values))

	for i, port := range values {
		portStrings[i] = fmt.Sprintf("%d", port)
	}

	if err := result.Set(strings.Join(portStrings, ",")); err != nil {
		return tproxy_config.Ports{}, err
	}

	return result, nil
}

func convertStringToPorts(values string) (tproxy_config.Ports, error) {
	result := tproxy_config.Ports{}

	if err := result.Set(values); err != nil {
		return tproxy_config.Ports{}, err
	}

	return result, nil
}

func convertToPorts[T []uint32 | string](v T) (tproxy_config.Ports, error) {
	switch v := any(v).(type) {
	case []uint32:
		return convertUint32SliceToPorts(v)
	case string:
		return convertStringToPorts(v)
	default:
		return tproxy_config.Ports{}, errors.Errorf("unsupported type %T", v)
	}
}
