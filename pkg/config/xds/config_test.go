package xds_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	kuma_xds "github.com/kumahq/kuma/pkg/config/xds"
	. "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("XdsServerConfig", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := kuma_xds.XdsServerConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.DataplaneConfigurationRefreshInterval.Duration).To(Equal(3 * time.Second))
		Expect(cfg.DataplaneStatusFlushInterval.Duration).To(Equal(5 * time.Second))
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
				"KUMA_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL": "3s",
				"KUMA_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL":          "5s",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := kuma_xds.XdsServerConfig{}

			// when
			err := config.Load("", &cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.DataplaneConfigurationRefreshInterval.Duration).To(Equal(3 * time.Second))
			Expect(cfg.DataplaneStatusFlushInterval.Duration).To(Equal(5 * time.Second))
		})
	})

	It("should have consistent defaults", func() {
		// given
		cfg := kuma_xds.DefaultXdsServerConfig()

		// when
		actual, err := config.ToYAML(cfg)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "default-config.golden.yaml")))
	})

	It("should have validators", func() {
		// given
		cfg := kuma_xds.XdsServerConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err).To(MatchError("parsing configuration from file 'testdata/invalid-config.input.yaml' failed: configuration validation failed: DataplaneConfigurationRefreshInterval must be positive"))
	})
})
