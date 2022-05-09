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

package kubernetes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/kubernetes"
)

var _ = Describe("kubernetes", func() {
	type testCaseKumactl struct {
		pod         *kube_core.Pod
		commandLine []string
	}

	DescribeTable("should generate kumactl command line", func(given testCaseKumactl) {
		podRedirect, err := kubernetes.NewPodRedirectForPod(given.pod)
		Expect(err).ToNot(HaveOccurred())

		commandLine := podRedirect.AsKumactlCommandLine()
		Expect(commandLine).To(Equal(given.commandLine))
	},
		Entry("should generate", testCaseKumactl{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaBuiltinDNS:                                 metadata.AnnotationEnabled,
						metadata.KumaBuiltinDNSPort:                             "25053",
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
					},
				},
			},
			commandLine: []string{
				"--redirect-outbound-port", "25100",
				"--redirect-inbound=" + "true",
				"--redirect-inbound-port", "25204",
				"--redirect-inbound-port-v6", "25206",
				"--kuma-dp-uid", "12345",
				"--exclude-inbound-ports", "12000",
				"--exclude-outbound-ports", "11000",
				"--verbose",
				"--skip-resolv-conf",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "25053",
			},
		}),

		Entry("should generate no builtin DNS", testCaseKumactl{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
					},
				},
			},
			commandLine: []string{
				"--redirect-outbound-port", "25100",
				"--redirect-inbound=" + "true",
				"--redirect-inbound-port", "25204",
				"--redirect-inbound-port-v6", "25206",
				"--kuma-dp-uid", "12345",
				"--exclude-inbound-ports", "12000",
				"--exclude-outbound-ports", "11000",
				"--verbose",
				"--skip-resolv-conf",
			},
		}),
		Entry("should generate experimental engine", testCaseKumactl{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
						metadata.KumaTransparentProxyingExperimentalEngine:      metadata.AnnotationEnabled,
					},
				},
			},
			commandLine: []string{
				"--redirect-outbound-port", "25100",
				"--redirect-inbound=" + "true",
				"--redirect-inbound-port", "25204",
				"--redirect-inbound-port-v6", "25206",
				"--kuma-dp-uid", "12345",
				"--exclude-inbound-ports", "12000",
				"--exclude-outbound-ports", "11000",
				"--verbose",
				"--skip-resolv-conf",
				"--experimental-transparent-proxy-engine",
			},
		}),
		Entry("should generate for Gateway", testCaseKumactl{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaBuiltinDNS:                                 metadata.AnnotationEnabled,
						metadata.KumaBuiltinDNSPort:                             "25053",
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaGatewayAnnotation:                          metadata.AnnotationEnabled,
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
					},
				},
			},
			commandLine: []string{
				"--redirect-outbound-port", "25100",
				"--redirect-inbound=" + "false",
				"--redirect-inbound-port", "25204",
				"--redirect-inbound-port-v6", "25206",
				"--kuma-dp-uid", "12345",
				"--exclude-inbound-ports", "12000",
				"--exclude-outbound-ports", "11000",
				"--verbose",
				"--skip-resolv-conf",
				"--redirect-all-dns-traffic",
				"--redirect-dns-port", "25053",
			},
		}),
	)

	type testCaseTransparentProxyConfig struct {
		pod      *kube_core.Pod
		tpConfig *config.TransparentProxyConfig
	}

	DescribeTable("should generate transparent proxy config", func(given testCaseTransparentProxyConfig) {
		podRedirect, err := kubernetes.NewPodRedirectForPod(given.pod)
		Expect(err).ToNot(HaveOccurred())

		tpConfig := podRedirect.AsTransparentProxyConfig()
		Expect(tpConfig).To(Equal(given.tpConfig))
	},
		Entry("should generate", testCaseTransparentProxyConfig{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaBuiltinDNS:                                 metadata.AnnotationEnabled,
						metadata.KumaBuiltinDNSPort:                             "25053",
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
					},
				},
			},
			tpConfig: &config.TransparentProxyConfig{
				DryRun:                 false,
				Verbose:                true,
				RedirectPortOutBound:   "25100",
				RedirectInBound:        true,
				RedirectPortInBound:    "25204",
				RedirectPortInBoundV6:  "25206",
				ExcludeInboundPorts:    "12000",
				ExcludeOutboundPorts:   "11000",
				UID:                    "12345",
				GID:                    "12345",
				RedirectDNS:            true,
				RedirectAllDNSTraffic:  false,
				AgentDNSListenerPort:   "25053",
				DNSUpstreamTargetChain: "",
			},
		}),
		Entry("should generate experimental engine", testCaseTransparentProxyConfig{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
						metadata.KumaTransparentProxyingExperimentalEngine:      metadata.AnnotationEnabled,
					},
				},
			},
			tpConfig: &config.TransparentProxyConfig{
				DryRun:                 false,
				Verbose:                true,
				RedirectPortOutBound:   "25100",
				RedirectInBound:        true,
				RedirectPortInBound:    "25204",
				RedirectPortInBoundV6:  "25206",
				ExcludeInboundPorts:    "12000",
				ExcludeOutboundPorts:   "11000",
				UID:                    "12345",
				GID:                    "12345",
				RedirectDNS:            false,
				RedirectAllDNSTraffic:  false,
				AgentDNSListenerPort:   "0",
				DNSUpstreamTargetChain: "",
				ExperimentalEngine:     true,
			},
		}),
		Entry("should generate no builtin DNS", testCaseTransparentProxyConfig{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
					},
				},
			},
			tpConfig: &config.TransparentProxyConfig{
				DryRun:                 false,
				Verbose:                true,
				RedirectPortOutBound:   "25100",
				RedirectInBound:        true,
				RedirectPortInBound:    "25204",
				RedirectPortInBoundV6:  "25206",
				ExcludeInboundPorts:    "12000",
				ExcludeOutboundPorts:   "11000",
				UID:                    "12345",
				GID:                    "12345",
				RedirectDNS:            false,
				RedirectAllDNSTraffic:  false,
				AgentDNSListenerPort:   "0",
				DNSUpstreamTargetChain: "",
			},
		}),
		Entry("should generate for Gateway", testCaseTransparentProxyConfig{
			pod: &kube_core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						metadata.KumaBuiltinDNS:                                 metadata.AnnotationEnabled,
						metadata.KumaBuiltinDNSPort:                             "25053",
						metadata.KumaTrafficExcludeOutboundPorts:                "11000",
						metadata.KumaTransparentProxyingOutboundPortAnnotation:  "25100",
						metadata.KumaGatewayAnnotation:                          metadata.AnnotationEnabled,
						metadata.KumaTrafficExcludeInboundPorts:                 "12000",
						metadata.KumaTransparentProxyingInboundPortAnnotation:   "25204",
						metadata.KumaTransparentProxyingInboundPortAnnotationV6: "25206",
						metadata.KumaSidecarUID:                                 "12345",
					},
				},
			},
			tpConfig: &config.TransparentProxyConfig{
				DryRun:                 false,
				Verbose:                true,
				RedirectPortOutBound:   "25100",
				RedirectInBound:        false,
				RedirectPortInBound:    "25204",
				RedirectPortInBoundV6:  "25206",
				ExcludeInboundPorts:    "12000",
				ExcludeOutboundPorts:   "11000",
				UID:                    "12345",
				GID:                    "12345",
				RedirectDNS:            true,
				RedirectAllDNSTraffic:  false,
				AgentDNSListenerPort:   "25053",
				DNSUpstreamTargetChain: "",
			},
		}),
	)
})
