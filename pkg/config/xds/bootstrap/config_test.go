package bootstrap_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config"
	. "github.com/Kong/kuma/pkg/config/xds/bootstrap"
)

var _ = Describe("BootstrappServerConfig", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := BootstrapServerConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.Port).To(Equal(uint32(1234)))
		Expect(cfg.Params.AdminAddress).To(Equal("192.168.0.1"))
		Expect(cfg.Params.AdminPort).To(Equal(uint32(4321)))
		Expect(cfg.Params.AdminAccessLogPath).To(Equal("/var/log"))
		Expect(cfg.Params.XdsHost).To(Equal("kuma-control-plane.internal"))
		Expect(cfg.Params.XdsPort).To(Equal(uint32(10101)))
		Expect(cfg.Params.XdsConnectTimeout).To(Equal(2 * time.Second))
		Expect(cfg.Params.XdsClientTlsCertFile).To(Equal("/tmp/clientcert.pem"))
		Expect(cfg.Params.XdsClientTlsKeyFile).To(Equal("/tmp/clientkey.pem"))
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
				"KUMA_BOOTSTRAP_SERVER_PORT":                         "1234",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ADDRESS":         "192.168.0.1",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT":            "4321",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ACCESS_LOG_PATH": "/var/log",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST":              "kuma-control-plane.internal",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_PORT":              "10101",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_CONNECT_TIMEOUT":   "2s",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_CLIENT_TLS_CERT_FILE": "/tmp/clientcertenv.pem",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_CLIENT_TLS_KEY_FILE":  "/tmp/clientkeyenv.pem",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := BootstrapServerConfig{}

			// when
			err := config.Load("", &cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.Port).To(Equal(uint32(1234)))
			Expect(cfg.Params.AdminAddress).To(Equal("192.168.0.1"))
			Expect(cfg.Params.AdminPort).To(Equal(uint32(4321)))
			Expect(cfg.Params.AdminAccessLogPath).To(Equal("/var/log"))
			Expect(cfg.Params.XdsHost).To(Equal("kuma-control-plane.internal"))
			Expect(cfg.Params.XdsPort).To(Equal(uint32(10101)))
			Expect(cfg.Params.XdsConnectTimeout).To(Equal(2 * time.Second))
			Expect(cfg.Params.XdsClientTlsCertFile).To(Equal("/tmp/clientcertenv.pem"))
			Expect(cfg.Params.XdsClientTlsKeyFile).To(Equal("/tmp/clientkeyenv.pem"))
		})
	})

	It("should have consistent defaults", func() {
		// given
		cfg := DefaultBootstrapServerConfig()

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
		cfg := BootstrapServerConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err).To(MatchError(`Invalid configuration: Port must be in the range [0, 65535]`))
	})
})
