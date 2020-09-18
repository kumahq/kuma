package probes_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
)

var _ = Describe("KumaProbe", func() {

	Context("ToInbound", func() {
		It("should convert virtual probe to inbound probe", func() {
			podProbeYaml := `
                httpGet:
                  path: /8080/c1/health/liveness
                  port: 9000
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			inbound, ok := probes.KumaProbe(probe).ToInbound()
			Expect(ok).To(BeTrue())
			Expect(inbound.Path()).To(Equal("/c1/health/liveness"))
			Expect(inbound.Port()).To(Equal(8080))
		})
	})

	Context("ToVirtual", func() {
		It("should convert inbound probe to virtual probe", func() {
			podProbeYaml := `
                httpGet:
                  path: /c1/health/liveness
                  port: 8080
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			virtual := probes.KumaProbe(probe).ToVirtual()
			Expect(virtual.Path()).To(Equal("/8080/c1/health/liveness"))
			Expect(virtual.Port()).To(Equal(probes.ProbePort))
		})
	})
})
