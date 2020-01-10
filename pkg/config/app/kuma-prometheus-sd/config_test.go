package kuma_prometheus_sd_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config"
	kuma_promsd "github.com/Kong/kuma/pkg/config/app/kuma-prometheus-sd"
)

var _ = Describe("Config", func() {

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
				"KUMA_CONTROL_PLANE_API_SERVER_URL":      "https://kuma-control-plane.internal:5682",
				"KUMA_MONITORING_ASSIGNMENT_CLIENT_NAME": "custom",
				"KUMA_PROMETHEUS_OUTPUT_FILE":            "/path/to/file",
			}
			for key, value := range env {
				os.Setenv(key, value)
			}

			// given
			cfg := kuma_promsd.Config{}

			// when
			err := config.Load("", &cfg)

			// then
			Expect(err).ToNot(HaveOccurred())

			// and
			Expect(cfg.ControlPlane.ApiServer.URL).To(Equal("https://kuma-control-plane.internal:5682"))
			Expect(cfg.MonitoringAssignment.Client.Name).To(Equal("custom"))
			Expect(cfg.Prometheus.OutputFile).To(Equal("/path/to/file"))
		})
	})

	It("should be loadable from configuration file", func() {
		// given
		cfg := kuma_promsd.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "valid-config.input.yaml"), &cfg)

		// then
		Expect(err).ToNot(HaveOccurred())

		// and
		Expect(cfg.ControlPlane.ApiServer.URL).To(Equal("https://kuma-control-plane.internal:5682"))
		Expect(cfg.MonitoringAssignment.Client.Name).To(Equal("custom"))
		Expect(cfg.Prometheus.OutputFile).To(Equal("/path/to/file"))
	})

	It("should have consistent defaults", func() {
		// given
		cfg := kuma_promsd.DefaultConfig()

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
		cfg := kuma_promsd.Config{}

		// when
		err := config.Load(filepath.Join("testdata", "invalid-config.input.yaml"), &cfg)

		// then
		Expect(err.Error()).To(Equal(`Invalid configuration: .ControlPlane is not valid: .ApiServer is not valid: .URL must be a valid absolute URI; .MonitoringAssignment is not valid: .Client is not valid: .Name must be non-empty; .Prometheus is not valid: .OutputFile must be non-empty`))
	})
})
