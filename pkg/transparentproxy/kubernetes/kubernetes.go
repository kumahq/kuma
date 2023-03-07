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
	"strconv"
	"strings"

	kube_core "k8s.io/api/core/v1"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
)

type PodRedirect struct {
	BuiltinDNSEnabled                        bool
	BuiltinDNSPort                           uint32
	ExcludeOutboundPorts                     string
	RedirectPortOutbound                     uint32
	RedirectInbound                          bool
	ExcludeInboundPorts                      string
	RedirectPortInbound                      uint32
	RedirectPortInboundV6                    uint32
	UID                                      string
	UseTransparentProxyEngineV1              bool
	TransparentProxyEnableEbpf               bool
	TransparentProxyEbpfBPFFSPath            string
	TransparentProxyEbpfCgroupPath           string
	TransparentProxyEbpfTCAttachIface        string
	TransparentProxyEbpfInstanceIPEnvVarName string
	TransparentProxyEbpfProgramsSourcePath   string
	ExcludeOutboundTCPPortsForUIDs           []string
	ExcludeOutboundUDPPortsForUIDs           []string
}

func NewPodRedirectForPod(transparentProxyV1 bool, pod *kube_core.Pod) (*PodRedirect, error) {
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

	excludeOutboundTCPPortsForUIDs, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTrafficExcludeOutboundTCPPortsForUIDs)
	if exists {
		podRedirect.ExcludeOutboundTCPPortsForUIDs = strings.Split(excludeOutboundTCPPortsForUIDs, ";")
	}

	excludeOutboundUDPPortsForUIDs, exists := metadata.Annotations(pod.Annotations).GetString(metadata.KumaTrafficExcludeOutboundUDPPortsForUIDs)
	if exists {
		podRedirect.ExcludeOutboundUDPPortsForUIDs = strings.Split(excludeOutboundUDPPortsForUIDs, ";")
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

	podRedirect.ExcludeInboundPorts, _ = metadata.Annotations(pod.Annotations).GetString(metadata.KumaTrafficExcludeInboundPorts)

	podRedirect.RedirectPortInbound, _, err = metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotation)
	if err != nil {
		return nil, err
	}

	podRedirect.RedirectPortInboundV6, _, err = metadata.Annotations(pod.Annotations).GetUint32(metadata.KumaTransparentProxyingInboundPortAnnotationV6)
	if err != nil {
		return nil, err
	}

	podRedirect.UID, _ = metadata.Annotations(pod.Annotations).GetString(metadata.KumaSidecarUID)

	podRedirect.UseTransparentProxyEngineV1, _, err = metadata.Annotations(pod.Annotations).GetEnabledWithDefault(
		transparentProxyV1,
		metadata.KumaTransparentProxyingEngineV1,
	)
	if err != nil {
		return nil, err
	}

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

	return podRedirect, nil
}

func (pr *PodRedirect) AsKumactlCommandLine() []string {
	result := []string{
		"--redirect-outbound-port",
		fmt.Sprintf("%d", pr.RedirectPortOutbound),
		"--redirect-inbound=" + fmt.Sprintf("%t", pr.RedirectInbound),
		"--redirect-inbound-port",
		fmt.Sprintf("%d", pr.RedirectPortInbound),
		"--redirect-inbound-port-v6",
		fmt.Sprintf("%d", pr.RedirectPortInboundV6),
		"--kuma-dp-uid",
		pr.UID,
		"--exclude-inbound-ports",
		pr.ExcludeInboundPorts,
		"--exclude-outbound-ports",
		pr.ExcludeOutboundPorts,
		"--verbose",
		// Remove with https://github.com/kumahq/kuma/issues/4759
		"--skip-resolv-conf",
	}

	for _, exclusion := range pr.ExcludeOutboundTCPPortsForUIDs {
		result = append(result,
			"--exclude-outbound-tcp-ports-for-uids",
			exclusion,
		)
	}

	for _, exclusion := range pr.ExcludeOutboundUDPPortsForUIDs {
		result = append(result,
			"--exclude-outbound-udp-ports-for-uids",
			exclusion,
		)
	}

	if pr.BuiltinDNSEnabled {
		result = append(result,
			"--redirect-all-dns-traffic",
			"--redirect-dns-port", strconv.FormatInt(int64(pr.BuiltinDNSPort), 10),
		)
	}

	if pr.UseTransparentProxyEngineV1 {
		result = append(result, "--use-transparent-proxy-engine-v1")
	}

	if pr.TransparentProxyEnableEbpf {
		result = append(result, "--ebpf-enabled")

		instanceIPEnvVarName := "INSTANCE_IP"
		if pr.TransparentProxyEbpfInstanceIPEnvVarName != "" {
			instanceIPEnvVarName = pr.TransparentProxyEbpfInstanceIPEnvVarName
		}
		result = append(result, "--ebpf-instance-ip", fmt.Sprintf("$(%s)", instanceIPEnvVarName))

		if pr.TransparentProxyEbpfBPFFSPath != "" {
			result = append(result, "--ebpf-bpffs-path", pr.TransparentProxyEbpfBPFFSPath)
		}

		if pr.TransparentProxyEbpfCgroupPath != "" {
			result = append(result, "--ebpf-cgroup-path", pr.TransparentProxyEbpfCgroupPath)
		}

		if pr.TransparentProxyEbpfTCAttachIface != "" {
			result = append(result, "--ebpf-tc-attach-iface", pr.TransparentProxyEbpfTCAttachIface)
		}

		if pr.TransparentProxyEbpfProgramsSourcePath != "" {
			result = append(result, "--ebpf-programs-source-path", pr.TransparentProxyEbpfProgramsSourcePath)
		}
	}

	return result
}
