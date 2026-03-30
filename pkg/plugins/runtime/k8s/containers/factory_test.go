package containers

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"

	runtime_k8s "github.com/kumahq/kuma/v2/pkg/config/plugins/runtime/k8s"
	config_types "github.com/kumahq/kuma/v2/pkg/config/types"
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
				BuiltinDNS:      runtime_k8s.BuiltinDNS{},
				otelPipeEnabled: true,
			}
			envVars, err := factory.sidecarEnvVars("default", nil)
			Expect(err).ToNot(HaveOccurred())

			_, ok := findEnvVar(envVars, "KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED")
			Expect(ok).To(BeFalse(), "expected KUMA_DATAPLANE_RUNTIME_OTEL_ENV_ENABLED to stay unset")
		})
	})
})
