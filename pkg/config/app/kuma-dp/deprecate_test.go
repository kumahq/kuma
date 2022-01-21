package kumadp_test

import (
	"os"
	"strings"

	kumadp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	"github.com/kumahq/kuma/pkg/config/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deprecate", func() {

	It("should print deprecation warnings if value is set in config", func() {
		// setup
		cfg := kumadp.DefaultConfig()
		cfg.Dataplane.AdminPort = types.MustExactPort(1234)

		// when
		var stringBuilder strings.Builder
		kumadp.PrintDeprecations(&cfg, &stringBuilder)

		// then
		expected := `Deprecated: Dataplane.AdminPort. Please set adminPort directly in Data Plane Proxy resource, in the field 'networking.admin.port'.
`
		Expect(stringBuilder.String()).To(Equal(expected))
	})

	It("should print deprecation warnings if env is set", func() {
		// setup
		cfg := kumadp.DefaultConfig()
		cfg.Dataplane.AdminPort = types.PortRange{}
		Expect(os.Setenv("KUMA_DATAPLANE_ADMIN_PORT", "1234")).To(Succeed())
		defer os.Unsetenv("KUMA_DATAPLANE_ADMIN_PORT")

		// when
		var stringBuilder strings.Builder
		kumadp.PrintDeprecations(&cfg, &stringBuilder)

		// then
		expected := `Deprecated: KUMA_DATAPLANE_ADMIN_PORT. Please set adminPort directly in Data Plane Proxy resource, in the field 'networking.admin.port'.
`
		Expect(stringBuilder.String()).To(Equal(expected))
	})
})
