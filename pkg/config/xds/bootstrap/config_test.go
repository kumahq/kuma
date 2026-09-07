package bootstrap_test

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config"
	. "github.com/kumahq/kuma/v3/pkg/config/xds/bootstrap"
)

var _ = Describe("BootstrappServerConfig", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := DefaultBootstrapServerConfig()

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.Params.AdminAddress).To(Equal("192.168.0.1"))
		Expect(cfg.Params.AdminPort).To(Equal(uint32(4321)))
		Expect(cfg.Params.AdminAccessLogPath).To(Equal("/var/log"))
		Expect(cfg.Params.XdsHost).To(Equal("kuma-control-plane.internal"))
		Expect(cfg.Params.XdsPort).To(Equal(uint32(10101)))
		Expect(cfg.Params.XdsConnectTimeout.Duration).To(Equal(2 * time.Second))
		Expect(cfg.Params.XdsGrpcMaxReceiveMessageBytes).To(Equal(uint32(33554432)))
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
				"KUMA_BOOTSTRAP_SERVER_API_VERSION":                               "v3",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ADDRESS":                      "192.168.0.1",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT":                         "4321",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ACCESS_LOG_PATH":              "/var/log",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST":                           "kuma-control-plane.internal",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_PORT":                           "10101",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_CONNECT_TIMEOUT":                "2s",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_GRPC_MAX_RECEIVE_MESSAGE_BYTES": "33554432",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := DefaultBootstrapServerConfig()

			// when
			err := config.Load("", cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.Params.AdminAddress).To(Equal("192.168.0.1"))
			Expect(cfg.Params.AdminPort).To(Equal(uint32(4321)))
			Expect(cfg.Params.AdminAccessLogPath).To(Equal("/var/log"))
			Expect(cfg.Params.XdsHost).To(Equal("kuma-control-plane.internal"))
			Expect(cfg.Params.XdsPort).To(Equal(uint32(10101)))
			Expect(cfg.Params.XdsConnectTimeout.Duration).To(Equal(2 * time.Second))
			Expect(cfg.Params.XdsGrpcMaxReceiveMessageBytes).To(Equal(uint32(33554432)))
		})
	})

	DescribeTable("should validate the readiness port",
		func(port uint32, valid bool) {
			cfg := DefaultBootstrapServerConfig()
			cfg.Params.ReadinessPort = port

			err := cfg.Validate()

			if valid {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(MatchError(ContainSubstring("ReadinessPort must be in the range (0, 65535]")))
			}
		},
		Entry("zero is rejected, the injector cannot build a probe for it", uint32(0), false),
		Entry("above the port range is rejected", uint32(65536), false),
		Entry("the default is accepted", uint32(9902), true),
		Entry("a custom port is accepted", uint32(19902), true),
	)

	It("should have consistent defaults", func() {
		// given
		cfg := DefaultBootstrapServerConfig()

		// when
		actual, err := config.ToYAML(cfg)
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := os.ReadFile(filepath.Join("testdata", "default-config.golden.yaml"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(actual).To(MatchYAML(expected))
	})
})
