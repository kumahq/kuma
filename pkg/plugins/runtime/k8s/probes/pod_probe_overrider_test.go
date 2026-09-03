package probes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/probes"
)

var _ = Describe("ApplicationProbeProxyPort", func() {
	DescribeTable("GetApplicationProbeProxyPort",
		func(annotations map[string]string, defaultPort int, expected int, expectedErr string) {
			port, err := probes.GetApplicationProbeProxyPort(annotations, uint32(defaultPort))

			if expectedErr != "" {
				Expect(err).To(MatchError(expectedErr))
			} else {
				Expect(err).ToNot(HaveOccurred())
				Expect(port).To(Equal(uint32(expected)))
			}
		},
		Entry("gateway mode with proxy port set", map[string]string{
			"kuma.io/application-probe-proxy-port": "9000",
			"kuma.io/gateway":                      "enabled",
		}, 10001, 0, "application probe proxies probes can't be enabled in gateway mode"),

		Entry("gateway mode without proxy", map[string]string{
			"kuma.io/gateway": "enabled",
		}, 10001, 0, ""),

		Entry("proxy port set", map[string]string{
			"kuma.io/application-probe-proxy-port": "9001",
		}, 10001, 9001, ""),

		Entry("default port fallback", map[string]string{}, 10001, 10001, ""),
	)
})

var _ = Describe("RealProbePorts", func() {
	It("returns the sorted, deduplicated ports probes target, resolving named ports and skipping the sidecar", func() {
		pod := &kube_core.Pod{
			Spec: kube_core.PodSpec{
				Containers: []kube_core.Container{
					{
						Name: "app",
						Ports: []kube_core.ContainerPort{
							{Name: "http", ContainerPort: 8080},
						},
						LivenessProbe: &kube_core.Probe{
							HTTPGet: &kube_core.HTTPGetAction{Port: intstr.FromString("http")},
						},
						ReadinessProbe: &kube_core.Probe{
							TCPSocket: &kube_core.TCPSocketAction{Port: intstr.FromInt32(6379)},
						},
						StartupProbe: &kube_core.Probe{
							GRPC: &kube_core.GRPCAction{Port: 8080},
						},
					},
					{
						Name: "kuma-sidecar",
						LivenessProbe: &kube_core.Probe{
							HTTPGet: &kube_core.HTTPGetAction{Port: intstr.FromInt32(9901)},
						},
					},
				},
			},
		}

		Expect(probes.RealProbePorts(pod)).To(Equal([]uint32{6379, 8080}))
	})

	It("doesn't mutate the pod's probes", func() {
		pod := &kube_core.Pod{
			Spec: kube_core.PodSpec{
				Containers: []kube_core.Container{
					{
						Name: "app",
						Ports: []kube_core.ContainerPort{
							{Name: "http", ContainerPort: 8080},
						},
						ReadinessProbe: &kube_core.Probe{
							HTTPGet: &kube_core.HTTPGetAction{Port: intstr.FromString("http")},
						},
					},
				},
			},
		}

		Expect(probes.RealProbePorts(pod)).To(Equal([]uint32{8080}))
		Expect(pod.Spec.Containers[0].ReadinessProbe.HTTPGet.Port).To(Equal(intstr.FromString("http")))
	})
})
