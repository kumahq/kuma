package kumadp_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	kuma_dp "github.com/kumahq/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/kumahq/kuma/pkg/config/types"
)

var _ = Describe("Config", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := kuma_dp.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.ControlPlane.URL).To(Equal("https://kuma-control-plane.internal:5682"))
		Expect(cfg.Dataplane.AdminPort).To(Equal(config_types.MustExactPort(2345)))
		Expect(cfg.Dataplane.DrainTime).To(Equal(60 * time.Second))
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
				"KUMA_CONTROL_PLANE_URL":                                 "https://kuma-control-plane.internal:5682",
				"KUMA_CONTROL_PLANE_RETRY_BACKOFF":                       "1s",
				"KUMA_CONTROL_PLANE_RETRY_MAX_DURATION":                  "10s",
				"KUMA_CONTROL_PLANE_BOOTSTRAP_SERVER_RETRY_BACKOFF":      "2s",
				"KUMA_CONTROL_PLANE_BOOTSTRAP_SERVER_RETRY_MAX_DURATION": "11s",
				"KUMA_DATAPLANE_MESH":                                    "demo",
				"KUMA_DATAPLANE_NAME":                                    "example",
				"KUMA_DATAPLANE_ADMIN_PORT":                              "2345",
				"KUMA_DATAPLANE_DRAIN_TIME":                              "60s",
				"KUMA_DATAPLANE_RUNTIME_BINARY_PATH":                     "envoy.sh",
				"KUMA_DATAPLANE_RUNTIME_CONFIG_DIR":                      "/var/run/envoy",
				"KUMA_DATAPLANE_RUNTIME_TOKEN_PATH":                      "/tmp/token",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := kuma_dp.Config{}

			// when
			err := config.Load("", &cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.ControlPlane.URL).To(Equal("https://kuma-control-plane.internal:5682"))
			Expect(cfg.ControlPlane.Retry.Backoff).To(Equal(1 * time.Second))
			Expect(cfg.ControlPlane.Retry.MaxDuration).To(Equal(10 * time.Second))
			Expect(cfg.Dataplane.Mesh).To(Equal("demo"))
			Expect(cfg.Dataplane.Name).To(Equal("example"))
			Expect(cfg.Dataplane.AdminPort).To(Equal(config_types.MustExactPort(2345)))
			Expect(cfg.Dataplane.DrainTime).To(Equal(60 * time.Second))
			Expect(cfg.DataplaneRuntime.BinaryPath).To(Equal("envoy.sh"))
			Expect(cfg.DataplaneRuntime.ConfigDir).To(Equal("/var/run/envoy"))
			Expect(cfg.DataplaneRuntime.TokenPath).To(Equal("/tmp/token"))
		})
	})

	It("should have consistent defaults", func() {
		// given
		cfg := kuma_dp.DefaultConfig()

		// when
		actual, err := config.ToYAML(&cfg)
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
		cfg := kuma_dp.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err.Error()).To(Equal(`Invalid configuration: .ControlPlane is not valid: .Retry is not valid: .Backoff must be a positive duration; .Dataplane is not valid: .Mesh must be non-empty; .Name must be non-empty; .DrainTime must be positive; .DataplaneRuntime is not valid: .BinaryPath must be non-empty`))
	})
})
