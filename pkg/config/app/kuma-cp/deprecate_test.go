package kuma_cp_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
)

var _ = Describe("Deprecate", func() {

	It("should print deprecation warnings if value is set in config", func() {
		// setup
		cfg := kuma_cp.DefaultConfig()
		cfg.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort = 1234

		// when
		var stringBuilder strings.Builder
		kuma_cp.PrintDeprecations(&cfg, &stringBuilder)

		// then
		expected := `Deprecated: Runtime.Kubernetes.Injector.SidecarContainer.AdminPort. Use BootstrapServer.Params.AdminPort instead.
`
		Expect(stringBuilder.String()).To(Equal(expected))
	})

	It("should print deprecation warnings if env is set", func() {
		// setup
		cfg := kuma_cp.DefaultConfig()
		cfg.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort = 0
		Expect(os.Setenv("KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT", "1234")).To(Succeed())
		defer os.Unsetenv("KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT")

		// when
		var stringBuilder strings.Builder
		kuma_cp.PrintDeprecations(&cfg, &stringBuilder)

		// then
		expected := `Deprecated: KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT. Use KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT instead.
`
		Expect(stringBuilder.String()).To(Equal(expected))
	})
})
