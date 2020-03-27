package kumadp_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config"
	kuma_dp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	config_types "github.com/Kong/kuma/pkg/config/types"
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
		Expect(cfg.ControlPlane.ApiServer.URL).To(Equal("https://kuma-control-plane.internal:5682"))
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
				"KUMA_CONTROL_PLANE_API_SERVER_URL":  "https://kuma-control-plane.internal:5682",
				"KUMA_DATAPLANE_MESH":                "demo",
				"KUMA_DATAPLANE_NAME":                "example",
				"KUMA_DATAPLANE_ADMIN_PORT":          "2345",
				"KUMA_DATAPLANE_DRAIN_TIME":          "60s",
				"KUMA_DATAPLANE_RUNTIME_BINARY_PATH": "envoy.sh",
				"KUMA_DATAPLANE_RUNTIME_CONFIG_DIR":  "/var/run/envoy",
				"KUMA_DATAPLANE_RUNTIME_TOKEN_PATH":  "/tmp/token",
				"KUMA_SDS_TYPE":  "dpVault",
				"KUMA_SDS_VAULT_ADDRESS": "https://vault.local",
				"KUMA_SDS_VAULT_AGENT_ADDRESS": "https://vault-agent.local",
				"KUMA_SDS_VAULT_TOKEN": "testToken",
				"KUMA_SDS_VAULT_NAMESPACE": "testNamespace",
				"KUMA_SDS_VAULT_TLS_CA_CERT_PATH": "/cert.pem",
				"KUMA_SDS_VAULT_TLS_CA_CERT_DIR": "/certs",
				"KUMA_SDS_VAULT_TLS_CLIENT_CERT_PATH": "/client/cert.pem",
				"KUMA_SDS_VAULT_TLS_CLIENT_KEY_PATH": "/client/key.pem",
				"KUMA_SDS_VAULT_TLS_SKIP_VERIFY": "true",
				"KUMA_SDS_VAULT_TLS_SERVER_NAME": "name",
				// todo
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
			Expect(cfg.ControlPlane.ApiServer.URL).To(Equal("https://kuma-control-plane.internal:5682"))
			Expect(cfg.Dataplane.Mesh).To(Equal("demo"))
			Expect(cfg.Dataplane.Name).To(Equal("example"))
			Expect(cfg.Dataplane.AdminPort).To(Equal(config_types.MustExactPort(2345)))
			Expect(cfg.Dataplane.DrainTime).To(Equal(60 * time.Second))
			Expect(cfg.DataplaneRuntime.BinaryPath).To(Equal("envoy.sh"))
			Expect(cfg.DataplaneRuntime.ConfigDir).To(Equal("/var/run/envoy"))
			Expect(cfg.DataplaneRuntime.TokenPath).To(Equal("/tmp/token"))
			Expect(cfg.SDS.Type).To(Equal(kuma_dp.SdsDpVault))
			Expect(cfg.SDS.Vault.Token).To(Equal("testToken"))
			Expect(cfg.SDS.Vault.Address).To(Equal("https://vault.local"))
			Expect(cfg.SDS.Vault.AgentAddress).To(Equal("https://vault-agent.local"))
			Expect(cfg.SDS.Vault.Namespace).To(Equal("testNamespace"))
			Expect(cfg.SDS.Vault.TLS.CaCertPath).To(Equal("/cert.pem"))
			Expect(cfg.SDS.Vault.TLS.CaCertDir).To(Equal("/certs"))
			Expect(cfg.SDS.Vault.TLS.ClientCertPath).To(Equal("/client/cert.pem"))
			Expect(cfg.SDS.Vault.TLS.ClientKeyPath).To(Equal("/client/key.pem"))
			Expect(cfg.SDS.Vault.TLS.SkipVerify).To(BeTrue())
			Expect(cfg.SDS.Vault.TLS.ServerName).To(Equal("name"))
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
		Expect(err).To(MatchError(`Invalid configuration: .ControlPlane is not valid: .ApiServer is not valid: .URL must be a valid absolute URI; .Dataplane is not valid: .Mesh must be non-empty; .Name must be non-empty; .DrainTime must be positive; .DataplaneRuntime is not valid: .BinaryPath must be non-empty`))
	})
})
