package diagnostics_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/diagnostics"

	"github.com/kumahq/kuma/pkg/config"
)

var _ = Describe("XdsServerConfig", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := diagnostics.DiagnosticsConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.ServerPort).To(Equal(3456))
	})

	Context("with modified environment variables", func() {

		var backupEnvVars []string

		BeforeEach(func() {
			backupEnvVars = os.Environ()
		})

		AfterEach(func() {
			os.Clearenv()
			for _, envVar := range backupEnvVars {
				parts := strings.SplitN(envVar, "=", 2)
				os.Setenv(parts[0], parts[1])
			}
		})

		It("should be loadable from environment variables", func() {
			// setup
			env := map[string]string{
				"KUMA_DIAGNOSTICS_SERVER_PORT": "3456",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := diagnostics.DiagnosticsConfig{}

			// when
			err := config.Load("", &cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.ServerPort).To(Equal(3456))
		})
	})

	It("should have consistent defaults", func() {
		// given
		cfg := diagnostics.DefaultDiagnosticsConfig()

		// when
		actual, err := config.ToYAML(cfg)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := ioutil.ReadFile(filepath.Join("testdata", "default-config.golden.yaml"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(expected))
	})

	It("should have validators", func() {
		// given
		cfg := diagnostics.DiagnosticsConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err).To(MatchError(`Invalid configuration: DataplaneConfigurationRefreshInterval must be positive`))
	})
})
