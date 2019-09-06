package xds_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	konvoy_xds "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/xds"
)

var _ = Describe("XdsServerConfig", func() {
	It("should be loadable from configuration file", func() {
		// given
		cfg := konvoy_xds.XdsServerConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.GrpcPort).To(Equal(1234))
		Expect(cfg.HttpPort).To(Equal(2345))
		Expect(cfg.DiagnosticsPort).To(Equal(3456))
		Expect(cfg.DataplaneConfigurationRefreshInterval).To(Equal(3 * time.Second))
		Expect(cfg.DataplaneStatusFlushInterval).To(Equal(5 * time.Second))
		Expect(cfg.Snapshot.SdsLocation).To(Equal("konvoy-control-plane:1234"))
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
				"KONVOY_XDS_SERVER_GRPC_PORT":                                "1234",
				"KONVOY_XDS_SERVER_HTTP_PORT":                                "2345",
				"KONVOY_XDS_SERVER_DIAGNOSTICS_PORT":                         "3456",
				"KONVOY_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL": "3s",
				"KONVOY_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL":          "5s",
				"KONVOY_XDS_SERVER_SNAPSHOT_SDS_LOCATION":                    "konvoy-control-plane:1234",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := konvoy_xds.XdsServerConfig{}

			// when
			err := config.Load("", &cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.GrpcPort).To(Equal(1234))
			Expect(cfg.HttpPort).To(Equal(2345))
			Expect(cfg.DiagnosticsPort).To(Equal(3456))
			Expect(cfg.DataplaneConfigurationRefreshInterval).To(Equal(3 * time.Second))
			Expect(cfg.DataplaneStatusFlushInterval).To(Equal(5 * time.Second))
			Expect(cfg.Snapshot.SdsLocation).To(Equal("konvoy-control-plane:1234"))
		})
	})

	It("should have consistent defaults", func() {
		// given
		cfg := konvoy_xds.DefaultXdsServerConfig()

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
		cfg := konvoy_xds.XdsServerConfig{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err).To(MatchError(`Invalid configuration: DataplaneConfigurationRefreshInterval must be positive`))
	})
})
