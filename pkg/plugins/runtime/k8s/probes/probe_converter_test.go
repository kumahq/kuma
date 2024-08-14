package probes_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/probes"
)

var _ = Describe("ProxiedApplicationProbe", func() {
	Describe("OverridingSupported", func() {
		type testCase struct {
			input             kube_core.Probe
			expectedSupported bool
		}

		DescribeTable("should check if probe is supported to be overridden",
			func(given testCase) {
				virtual := probes.ProxiedApplicationProbe(given.input)

				Expect(virtual.OverridingSupported()).To(Equal(given.expectedSupported))
			},
			Entry("HTTP", testCase{
				input: kube_core.Probe{
					ProbeHandler: kube_core.ProbeHandler{
						HTTPGet: &kube_core.HTTPGetAction{},
					},
				},
				expectedSupported: true,
			}),
			Entry("TCPSocket", testCase{
				input: kube_core.Probe{
					ProbeHandler: kube_core.ProbeHandler{
						TCPSocket: &kube_core.TCPSocketAction{},
					},
				},
				expectedSupported: true,
			}),
			Entry("gRPC", testCase{
				input: kube_core.Probe{
					ProbeHandler: kube_core.ProbeHandler{
						GRPC: &kube_core.GRPCAction{},
					},
				},
				expectedSupported: true,
			}),
			Entry("exec", testCase{
				input: kube_core.Probe{
					ProbeHandler: kube_core.ProbeHandler{
						Exec: &kube_core.ExecAction{},
					},
				},
				expectedSupported: false,
			}))
	})

	Context("ToVirtual - HTTP", func() {
		It("should convert pod probe to virtual probe", func() {
			podProbeYaml := `
                httpGet:
                  path: /c1/health/liveness
                  port: 8080
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			virtual, err := probes.ProxiedApplicationProbe(probe).ToVirtual(9000)
			Expect(err).ToNot(HaveOccurred())
			Expect(virtual.Path()).To(Equal("/8080/c1/health/liveness"))
			Expect(virtual.Port()).To(Equal(uint32(9000)))
		})

		It("should convert probe with custom headers and timeout", func() {
			podProbeYaml := `
                timeoutSeconds: 15
                httpGet:
                  scheme: HTTPS
                  path: /c1/healthz
                  port: 8080
                  httpHeaders:
                  - name: Host
                    value: example.com
                  - name: X-Custom-Header
                    value: custom-value
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			virtual, err := probes.ProxiedApplicationProbe(probe).ToVirtual(9000)
			Expect(err).ToNot(HaveOccurred())

			Expect(virtual.Path()).To(Equal("/8080/c1/healthz"))
			Expect(virtual.Port()).To(Equal(uint32(9000)))

			Expect(getHeader(virtual.Headers(), probes.HeaderNameHost)).To(Equal("example.com"))
			Expect(getHeader(virtual.Headers(), probes.HeaderNameScheme)).To(Equal("HTTPS"))
			Expect(getHeader(virtual.Headers(), probes.HeaderNameTimeout)).To(Equal("15"))
			Expect(getHeader(virtual.Headers(), "X-Custom-Header")).To(Equal("custom-value"))
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

			_, err = probes.ProxiedApplicationProbe(probe).ToVirtual(9000)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot override Pod's probes. Port for probe cannot be set " +
				"to 9000. It is reserved for the dataplane that will serve pods without mTLS."))
		})
	})

	Context("ToVirtual - TCP Socket & gRPC", func() {
		It("should convert TCP socket probe", func() {
			podProbeYaml := `
                timeoutSeconds: 10
                tcpSocket:
                  port: 6379
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			virtual, err := probes.ProxiedApplicationProbe(probe).ToVirtual(9000)
			Expect(err).ToNot(HaveOccurred())

			Expect(virtual.TCPSocket).To(BeNil())
			Expect(virtual.Path()).To(Equal("/tcp/6379"))
			Expect(virtual.Port()).To(Equal(uint32(9000)))

			Expect(getHeader(virtual.Headers(), probes.HeaderNameTimeout)).To(Equal("10"))
		})

		It("should convert gRPC probe", func() {
			podProbeYaml := `
                timeoutSeconds: 10
                grpc:
                  port: 6379
                  service: liveness
`
			probe := kube_core.Probe{}
			err := yaml.Unmarshal([]byte(podProbeYaml), &probe)
			Expect(err).ToNot(HaveOccurred())

			virtual, err := probes.ProxiedApplicationProbe(probe).ToVirtual(9000)
			Expect(err).ToNot(HaveOccurred())

			Expect(virtual.GRPC).To(BeNil())
			Expect(virtual.Path()).To(Equal("/grpc/6379"))
			Expect(virtual.Port()).To(Equal(uint32(9000)))

			Expect(getHeader(virtual.Headers(), probes.HeaderNameTimeout)).To(Equal("10"))
			Expect(getHeader(virtual.Headers(), probes.HeaderNameGRPCService)).To(Equal("liveness"))
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

			virtual, err := probes.ProxiedApplicationProbe(probe).ToVirtual(9000)
			Expect(err).ToNot(HaveOccurred())
			Expect(virtual.Path()).To(Equal("/8080/c1/hc"))
			Expect(virtual.Port()).To(Equal(uint32(9000)))
		})
	})
})

func getHeader(headers []kube_core.HTTPHeader, name string) string {
	for _, header := range headers {
		if header.Name == name {
			return header.Value
		}
	}
	return ""
}
