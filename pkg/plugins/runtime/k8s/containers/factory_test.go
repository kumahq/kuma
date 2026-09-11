package containers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"

	runtime_k8s "github.com/kumahq/kuma/v3/pkg/config/plugins/runtime/k8s"
	config_types "github.com/kumahq/kuma/v3/pkg/config/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/runtime/k8s/metadata"
)

var _ = Describe("DataplaneProxyFactory", func() {
	Describe("sidecarEnvVars", func() {
		findEnvVar := func(envVars []kube_core.EnvVar, name string) (kube_core.EnvVar, bool) {
			for _, ev := range envVars {
				if ev.Name == name {
					return ev, true
				}
			}
			return kube_core.EnvVar{}, false
		}

		It("does not inject OTEL env feature flags", func() {
			factory := &DataplaneProxyFactory{
				ContainerConfig: runtime_k8s.DataplaneContainer{
					DrainTime: config_types.Duration{Duration: 30 * time.Second},
					EnvVars:   map[string]string{},
				},
				BuiltinDNS: runtime_k8s.BuiltinDNS{},
			}
			envVars, err := factory.sidecarEnvVars("default", nil)
			Expect(err).ToNot(HaveOccurred())

			_, ok := findEnvVar(envVars, "KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED")
			Expect(ok).To(BeFalse(), "expected KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED to stay unset")
		})

		It("passes the component log level annotation to the sidecar", func() {
			factory := &DataplaneProxyFactory{
				ContainerConfig: runtime_k8s.DataplaneContainer{
					DrainTime: config_types.Duration{Duration: 30 * time.Second},
					EnvVars:   map[string]string{},
				},
				BuiltinDNS: runtime_k8s.BuiltinDNS{},
			}
			envVars, err := factory.sidecarEnvVars("default", map[string]string{
				metadata.KumaComponentLogLevel: "dnsproxy:debug",
			})
			Expect(err).ToNot(HaveOccurred())

			envVar, ok := findEnvVar(envVars, "KUMA_DATAPLANE_RUNTIME_COMPONENT_LOG_LEVEL")
			Expect(ok).To(BeTrue(), "expected KUMA_DATAPLANE_RUNTIME_COMPONENT_LOG_LEVEL to be set")
			Expect(envVar.Value).To(Equal("dnsproxy:debug"))
		})

		It("leaves the component log level unset without the annotation", func() {
			factory := &DataplaneProxyFactory{
				ContainerConfig: runtime_k8s.DataplaneContainer{
					DrainTime: config_types.Duration{Duration: 30 * time.Second},
					EnvVars:   map[string]string{},
				},
				BuiltinDNS: runtime_k8s.BuiltinDNS{},
			}
			envVars, err := factory.sidecarEnvVars("default", nil)
			Expect(err).ToNot(HaveOccurred())

			_, ok := findEnvVar(envVars, "KUMA_DATAPLANE_RUNTIME_COMPONENT_LOG_LEVEL")
			Expect(ok).To(BeFalse())
		})
	})

	DescribeTable("proxyConcurrencyFor",
		func(cpuLimit string, annotations map[string]string, expected int64) {
			factory := &DataplaneProxyFactory{
				ContainerConfig: runtime_k8s.DataplaneContainer{
					Resources: runtime_k8s.SidecarResources{
						Limits: runtime_k8s.SidecarResourceLimits{CPU: cpuLimit},
					},
				},
			}
			concurrency, err := factory.proxyConcurrencyFor(annotations)
			Expect(err).ToNot(HaveOccurred())
			Expect(concurrency).To(Equal(expected))
		},
		Entry("defaults to 2 without a CPU limit", "0", nil, int64(2)),
		Entry("floors at 2 with a small CPU limit", "500m", nil, int64(2)),
		Entry("follows a larger CPU limit", "3000m", nil, int64(3)),
		Entry("annotation overrides the CPU limit", "3000m", map[string]string{metadata.KumaSidecarConcurrencyAnnotation: "8"}, int64(8)),
		Entry("annotation 0 keeps Envoy's own sizing", "0", map[string]string{metadata.KumaSidecarConcurrencyAnnotation: "0"}, int64(0)),
	)
})
