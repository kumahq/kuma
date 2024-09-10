/*
Copyright 2021 Kuma authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"fmt"
	"slices"
	"strings"

	kube_core "k8s.io/api/core/v1"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

type PodRedirect struct {
	// while https://github.com/kumahq/kuma/issues/8324 is not implemented, when changing the config,
	// keep in mind to update all other places listed in the issue

	BuiltinDNSEnabled                        bool
	BuiltinDNSPort                           uint32
	ExcludeOutboundPorts                     string
	RedirectPortOutbound                     uint32
	RedirectInbound                          bool
	ExcludeInboundPorts                      string
	RedirectPortInbound                      uint32
	IpFamilyMode                             string
	UID                                      string
	TransparentProxyEnableEbpf               bool
	TransparentProxyEbpfBPFFSPath            string
	TransparentProxyEbpfCgroupPath           string
	TransparentProxyEbpfTCAttachIface        string
	TransparentProxyEbpfInstanceIPEnvVarName string
	TransparentProxyEbpfProgramsSourcePath   string
	ExcludeOutboundPortsForUIDs              []string
	DropInvalidPackets                       bool
	IptablesLogs                             bool
	ExcludeInboundIPs                        string
	ExcludeOutboundIPs                       string
}

func NewPodRedirectForPod(pod *kube_core.Pod) (*PodRedirect, error) {
	var err error
	podRedirect := &PodRedirect{}

	podRedirect.BuiltinDNSEnabled, _, err = metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaBuiltinDNS)
	if err != nil {
		return nil, err
	}

	podRedirect.BuiltinDNSPort, _, err = metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaBuiltinDNSPort)
	if err != nil {
		return nil, err
	}

	podRedirect.ExcludeOutboundPorts, _ = metadata.Annotations(pod.Annotations).GetString(metadata.KumaTrafficExcludeOutboundPorts)
	excludeOutboundPortsForUIDs, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTrafficExcludeOutboundPortsForUIDs)
	if exists {
		podRedirect.ExcludeOutboundPortsForUIDs = strings.Split(excludeOutboundPortsForUIDs, ";")
	}

	podRedirect.RedirectPortOutbound, _, err = metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaTransparentProxyingOutboundPortAnnotation)
	if err != nil {
		return nil, err
	}

	podRedirect.RedirectInbound = true
	enabled, exist, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaGatewayAnnotation)
	if err != nil {
		return nil, err
	}
	if exist && enabled {
		podRedirect.RedirectInbound = false
	}

	podRedirect.ExcludeInboundPorts = excludeApplicationProbeProxyPort(pod.Annotations)
	podRedirect.RedirectPortInbound, _, err = metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotation)
	if err != nil {
		return nil, err
	}

	podRedirect.IpFamilyMode, _ = metadata.Annotations(pod.Annotations).GetStringWithDefault(metadata.IpFamilyModeDualStack, metadata.KumaTransparentProxyingIPFamilyMode)

	podRedirect.DropInvalidPackets, _, _ = metadata.Annotations(pod.Annotations).GetBoolean(metadata.KumaTrafficDropInvalidPackets)

	podRedirect.IptablesLogs, _, _ = metadata.Annotations(pod.Annotations).GetBoolean(metadata.KumaTrafficIptablesLogs)

	podRedirect.UID, _ = metadata.Annotations(pod.Annotations).GetString(metadata.KumaSidecarUID)

	if value, exists, err := metadata.Annotations(pod.Annotations).GetEnabled(metadata.KumaTransparentProxyingEbpf); err != nil {
		return nil, err
	} else if exists {
		podRedirect.TransparentProxyEnableEbpf = value
	}

	if value, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTransparentProxyingEbpfBPFFSPath); exists {
		podRedirect.TransparentProxyEbpfBPFFSPath = value
	}

	if value, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTransparentProxyingEbpfCgroupPath); exists {
		podRedirect.TransparentProxyEbpfCgroupPath = value
	}

	if value, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTransparentProxyingEbpfTCAttachIface); exists {
		podRedirect.TransparentProxyEbpfTCAttachIface = value
	}

	if value, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTransparentProxyingEbpfInstanceIPEnvVarName); exists {
		podRedirect.TransparentProxyEbpfInstanceIPEnvVarName = value
	}

	if value, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTransparentProxyingEbpfProgramsSourcePath); exists {
		podRedirect.TransparentProxyEbpfProgramsSourcePath = value
	}

	if value, exists := metadata.Annotations(pod.Annotations).GetString(
		metadata.KumaTrafficExcludeInboundIPs,
	); exists {
		var addresses []string

		for _, address := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(address); trimmed != "" {
				addresses = append(addresses, trimmed)
			}
		}

		podRedirect.ExcludeInboundIPs = strings.Join(addresses, ",")
	}

	if value, exists := metadata.Annotations(pod.Annotations).GetString(
		metadata.KumaTrafficExcludeOutboundIPs,
	); exists {
		var addresses []string

		for _, address := range strings.Split(value, ",") {
			if trimmed := strings.TrimSpace(address); trimmed != "" {
				addresses = append(addresses, trimmed)
			}
		}

		podRedirect.ExcludeOutboundIPs = strings.Join(addresses, ",")
	}

	return podRedirect, nil
}

func excludeApplicationProbeProxyPort(annotations map[string]string) string {
	// the annotations are validated/defaulted in a previous step in injector.NewAnnotations, so we can safely ignore the errors here
	inboundPortsToExclude, _ := metadata.Annotations(annotations).GetString(metadata.KumaTrafficExcludeInboundPorts)
	appProbeProxyPort, _ := metadata.Annotations(annotations).GetString(metadata.KumaApplicationProbeProxyPortAnnotation)
	if appProbeProxyPort == "0" || appProbeProxyPort == "" {
		return inboundPortsToExclude
	}

	if inboundPortsToExclude == "" {
		return appProbeProxyPort
	}

	return fmt.Sprintf("%s,%s", inboundPortsToExclude, appProbeProxyPort)
}

func flag[T string | bool | uint32](name string, values ...T) []string {
	var result []string

	for _, value := range values {
		boolValue, isBool := any(value).(bool)
		switch {
		case isBool && boolValue:
			result = append(result, fmt.Sprintf("--%s", name))
		case value != *new(T):
			result = append(result, fmt.Sprintf("--%s=%v", name, value))
		}
	}

	return result
}

func flagsIf[T string | bool](condition T, flags ...[]string) []string {
	if condition == *new(T) {
		return nil
	}

	return slices.Concat(flags...)
}

func (pr *PodRedirect) AsKumactlCommandLine() []string {
	return slices.Concat(
		flag("kuma-dp-user", pr.UID),
		flag("ip-family-mode", pr.IpFamilyMode),
		// outbound
		flag("redirect-outbound-port", pr.RedirectPortOutbound),
		flag("exclude-outbound-ports", pr.ExcludeOutboundPorts),
		flag("exclude-outbound-ips", pr.ExcludeOutboundIPs),
		flag("exclude-outbound-ports-for-uids", pr.ExcludeOutboundPortsForUIDs...),
		// inbound
		flagsIf(!pr.RedirectInbound,
			flag("redirect-inbound", "false"),
		),
		flagsIf(pr.RedirectInbound,
			flag("redirect-inbound-port", pr.RedirectPortInbound),
			flag("exclude-inbound-ports", pr.ExcludeInboundPorts),
			flag("exclude-inbound-ips", pr.ExcludeInboundIPs),
		),
		// dns
		flagsIf(pr.BuiltinDNSEnabled,
			flag("redirect-all-dns-traffic", true),
			flag("redirect-dns-port", pr.BuiltinDNSPort),
		),
		// ebpf
		flagsIf(pr.TransparentProxyEnableEbpf,
			flag("ebpf-enabled", true),
			flag("ebpf-bpffs-path", pr.TransparentProxyEbpfBPFFSPath),
			flag("ebpf-cgroup-path", pr.TransparentProxyEbpfCgroupPath),
			flag("ebpf-tc-attach-iface", pr.TransparentProxyEbpfTCAttachIface),
			flag("ebpf-programs-source-path", pr.TransparentProxyEbpfProgramsSourcePath),
			flagsIf(pr.TransparentProxyEbpfInstanceIPEnvVarName,
				flag("ebpf-instance-ip", fmt.Sprintf("$(%s)", pr.TransparentProxyEbpfInstanceIPEnvVarName)),
			),
		),
		// other
		flag("drop-invalid-packets", pr.DropInvalidPackets),
		flag("iptables-logs", pr.IptablesLogs),
		flag("verbose", true),
	)
}
