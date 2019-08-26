package konvoydp_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	konvoy_dp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-dp"
)

var _ = Describe("Config", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := konvoy_dp.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.ControlPlane.BootstrapServer.Address).To(Equal("konvoy-control-plane.internal"))
		Expect(cfg.ControlPlane.BootstrapServer.Port).To(Equal(uint32(1234)))
		Expect(cfg.Dataplane.AdminPort).To(Equal(uint32(2345)))
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
				"KONVOY_CONTROL_PLANE_BOOTSTRAP_SERVER_ADDRESS": "konvoy-control-plane.internal",
				"KONVOY_CONTROL_PLANE_BOOTSTRAP_SERVER_PORT":    "1234",
				"KONVOY_DATAPLANE_ID":                           "example",
				"KONVOY_DATAPLANE_ADMIN_PORT":                   "2345",
				"KONVOY_DATAPLANE_RUNTIME_BINARY_PATH":          "envoy.sh",
				"KONVOY_DATAPLANE_RUNTIME_CONFIG_DIR":           "/var/run/envoy",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := konvoy_dp.Config{}

			// when
			err := config.Load("", &cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.ControlPlane.BootstrapServer.Address).To(Equal("konvoy-control-plane.internal"))
			Expect(cfg.ControlPlane.BootstrapServer.Port).To(Equal(uint32(1234)))
			Expect(cfg.Dataplane.Id).To(Equal("example"))
			Expect(cfg.Dataplane.AdminPort).To(Equal(uint32(2345)))
			Expect(cfg.DataplaneRuntime.BinaryPath).To(Equal("envoy.sh"))
			Expect(cfg.DataplaneRuntime.ConfigDir).To(Equal("/var/run/envoy"))
		})
	})

	It("should have consistent defaults", func() {
		// given
		cfg := konvoy_dp.DefaultConfig()

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
		cfg := konvoy_dp.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err).To(MatchError(`Invalid configuration: .ControlPlane is not valid: .BootstrapServer is not valid: .Address must be non-empty; .Port must be in the range [0, 65535]; .Dataplane is not valid: .Id must be non-empty; .AdminPort must be in the range [0, 65535]; .DataplaneRuntime is not valid: .BinaryPath must be non-empty; .ConfigDir must be non-empty`))
	})
})
