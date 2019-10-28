package config_test

import (
	"io/ioutil"
	"os"

	"github.com/Kong/kuma/pkg/config"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/Kong/kuma/pkg/config/core"
	"github.com/Kong/kuma/pkg/config/core/resources/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config loader", func() {

	var configFile *os.File

	BeforeEach(func() {
		os.Clearenv()
		file, err := ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
		configFile = file
	})

	AfterEach(func() {
		if configFile != nil {
			err := os.Remove(configFile.Name())
			Expect(err).ToNot(HaveOccurred())
		}
	})

	setEnv := func(key, value string) {
		err := os.Setenv(key, value)
		Expect(err).ToNot(HaveOccurred())
	}

	type testCase struct {
		yamlFileConfig string
		envVars        map[string]string
	}
	DescribeTable("should load config",
		func(given testCase) {
			// given file with sample config
			file, err := ioutil.TempFile("", "*")
			Expect(err).ToNot(HaveOccurred())
			_, err = file.WriteString(given.yamlFileConfig)
			Expect(err).ToNot(HaveOccurred())

			// and config from environment variables
			for key, value := range given.envVars {
				setEnv(key, value)
			}

			// when
			cfg := kuma_cp.DefaultConfig()
			err = config.Load(file.Name(), &cfg)
			Expect(err).ToNot(HaveOccurred())

			// then
			Expect(cfg.XdsServer.GrpcPort).To(Equal(5000))
			Expect(cfg.XdsServer.DiagnosticsPort).To(Equal(5003))

			Expect(cfg.BootstrapServer.Port).To(Equal(uint32(5004)))
			Expect(cfg.BootstrapServer.Params.AdminPort).To(Equal(uint32(1234)))
			Expect(cfg.BootstrapServer.Params.XdsHost).To(Equal("kuma-control-plane"))
			Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(4321)))

			Expect(cfg.Environment).To(Equal(config_core.KubernetesEnvironment))

			Expect(cfg.Store.Type).To(Equal(store.PostgresStore))
			Expect(cfg.Store.Postgres.Host).To(Equal("postgres.host"))
			Expect(cfg.Store.Postgres.Port).To(Equal(5432))
			Expect(cfg.Store.Postgres.User).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.Password).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.DbName).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.ConnectionTimeout).To(Equal(10))

			Expect(cfg.ApiServer.Port).To(Equal(9090))
			Expect(cfg.ApiServer.ReadOnly).To(Equal(true))

			Expect(cfg.DataplaneTokenServer.Local.Port).To(Equal(uint32(1111)))
			Expect(cfg.DataplaneTokenServer.Public.Port).To(Equal(uint32(2222)))
			Expect(cfg.DataplaneTokenServer.Public.Interface).To(Equal("192.168.0.1"))
			Expect(cfg.DataplaneTokenServer.Public.TlsKeyFile).To(Equal("/tmp/key"))
			Expect(cfg.DataplaneTokenServer.Public.TlsCertFile).To(Equal("/tmp/cert"))
			Expect(cfg.DataplaneTokenServer.Public.ClientCertFiles).To(Equal([]string{"/tmp/cert1", "/tmp/cert2"}))

			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Address).To(Equal("127.0.0.2"))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Port).To(Equal(uint32(9443)))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.CertDir).To(Equal("/var/run/secrets/kuma.io/kuma-admission-server/tls-cert"))

			Expect(cfg.Reports.Enabled).To(BeFalse())
		},
		Entry("from config file", testCase{
			envVars: map[string]string{},
			yamlFileConfig: `
environment: kubernetes
store:
  type: postgres
  postgres:
    host: postgres.host
    port: 5432
    user: kuma
    password: kuma
    dbName: kuma
    connectionTimeout: 10
xdsServer:
  grpcPort: 5000
  diagnosticsPort: 5003
bootstrapServer:
  port: 5004
  params:
    adminPort: 1234
    xdsHost: kuma-control-plane
    xdsPort: 4321
apiServer:
  port: 9090
  readOnly: true
dataplaneTokenServer:
  local:
    port: 1111
  public:
    interface: 192.168.0.1
    port: 2222
    tlsCertFile: /tmp/cert
    tlsKeyFile: /tmp/key
    clientCertFiles:
    - /tmp/cert1
    - /tmp/cert2
runtime:
  kubernetes:
    admissionServer:
      address: 127.0.0.2
      port: 9443
      certDir: /var/run/secrets/kuma.io/kuma-admission-server/tls-cert
reports:
  enabled: false
`,
		}),
		Entry("from env variables", testCase{
			envVars: map[string]string{
				"KUMA_XDS_SERVER_GRPC_PORT":                            "5000",
				"KUMA_XDS_SERVER_DIAGNOSTICS_PORT":                     "5003",
				"KUMA_BOOTSTRAP_SERVER_PORT":                           "5004",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT":              "1234",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST":                "kuma-control-plane",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_PORT":                "4321",
				"KUMA_ENVIRONMENT":                                     "kubernetes",
				"KUMA_STORE_TYPE":                                      "postgres",
				"KUMA_STORE_POSTGRES_HOST":                             "postgres.host",
				"KUMA_STORE_POSTGRES_PORT":                             "5432",
				"KUMA_STORE_POSTGRES_USER":                             "kuma",
				"KUMA_STORE_POSTGRES_PASSWORD":                         "kuma",
				"KUMA_STORE_POSTGRES_DB_NAME":                          "kuma",
				"KUMA_STORE_POSTGRES_CONNECTION_TIMEOUT":               "10",
				"KUMA_API_SERVER_READ_ONLY":                            "true",
				"KUMA_API_SERVER_PORT":                                 "9090",
				"KUMA_DATAPLANE_TOKEN_SERVER_LOCAL_PORT":               "1111",
				"KUMA_DATAPLANE_TOKEN_SERVER_PUBLIC_INTERFACE":         "192.168.0.1",
				"KUMA_DATAPLANE_TOKEN_SERVER_PUBLIC_PORT":              "2222",
				"KUMA_DATAPLANE_TOKEN_SERVER_PUBLIC_TLS_KEY_FILE":      "/tmp/key",
				"KUMA_DATAPLANE_TOKEN_SERVER_PUBLIC_TLS_CERT_FILE":     "/tmp/cert",
				"KUMA_DATAPLANE_TOKEN_SERVER_PUBLIC_CLIENT_CERT_FILES": "/tmp/cert1,/tmp/cert2",
				"KUMA_REPORTS_ENABLED":                                 "false",
				"KUMA_KUBERNETES_ADMISSION_SERVER_ADDRESS":             "127.0.0.2",
				"KUMA_KUBERNETES_ADMISSION_SERVER_PORT":                "9443",
				"KUMA_KUBERNETES_ADMISSION_SERVER_CERT_DIR":            "/var/run/secrets/kuma.io/kuma-admission-server/tls-cert",
			},
			yamlFileConfig: "",
		}),
	)

	It("should override via env var", func() {
		// given file with sample cfg
		file, err := ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
		_, err = file.WriteString("environment: kubernetes")
		Expect(err).ToNot(HaveOccurred())

		// and overridden config
		setEnv("KUMA_ENVIRONMENT", "universal")

		// when
		cfg := kuma_cp.DefaultConfig()
		err = config.Load(file.Name(), &cfg)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfg.Environment).To(Equal(config_core.UniversalEnvironment))

	})

})
