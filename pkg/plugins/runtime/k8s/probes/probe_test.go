package probes_test

import (
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo/v2"
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

			inbound, err := probes.KumaProbe(probe).ToReal(9000)
			Expect(err).ToNot(HaveOccurred())
			Expect(inbound.Path()).To(Equal("/c1/health/liveness"))
			Expect(inbound.Port()).To(Equal(uint32(8080)))
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

			virtual, err := probes.KumaProbe(probe).ToVirtual(9000)
			Expect(err).ToNot(HaveOccurred())
			Expect(virtual.Path()).To(Equal("/8080/c1/health/liveness"))
			Expect(virtual.Port()).To(Equal(uint32(9000)))
		})

		It("should return an error if virtual port is equal to real", func() {
			podProbeYaml := `
                httpGet:
                  path: /c1/health/liveness
                  port: 9000
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			_, err = probes.KumaProbe(probe).ToVirtual(9000)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot override Pod's probes. Port for probe cannot be set " +
				"to 9000. It is reserved for the dataplane that will serve pods without mTLS."))
		})
	})

	Context("Prepend /", func() {
		It("should convert to path with prepended /", func() {
			podProbeYaml := `
                httpGet:
                  path: c1/hc
                  port: 8080
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			virtual, err := probes.KumaProbe(probe).ToVirtual(9000)
			Expect(err).ToNot(HaveOccurred())
			Expect(virtual.Path()).To(Equal("/8080/c1/hc"))
			Expect(virtual.Port()).To(Equal(uint32(9000)))
		})
	})
})
