package config_test

import (
	"io/ioutil"
	"os"

	"github.com/Kong/kuma/pkg/config"
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/config/core/resources/store"
	. "github.com/onsi/ginkgo"
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

	sampleConfigYaml := `
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
  httpPort: 5001
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
  apiDocsPath: "/apidocs.json"
`

	It("should load config from file", func() {
		// given file with sample config
		file, err := ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
		_, err = file.WriteString(sampleConfigYaml)
		Expect(err).ToNot(HaveOccurred())

		// when
		cfg := kuma_cp.DefaultConfig()
		err = config.Load(file.Name(), &cfg)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfg.XdsServer.GrpcPort).To(Equal(5000))
		Expect(cfg.XdsServer.HttpPort).To(Equal(5001))
		Expect(cfg.XdsServer.DiagnosticsPort).To(Equal(5003))

		Expect(cfg.BootstrapServer.Port).To(Equal(5004))
		Expect(cfg.BootstrapServer.Params.AdminPort).To(Equal(uint32(1234)))
		Expect(cfg.BootstrapServer.Params.XdsHost).To(Equal("kuma-control-plane"))
		Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(4321)))

		Expect(cfg.Environment).To(Equal(kuma_cp.KubernetesEnvironment))

		Expect(cfg.Store.Type).To(Equal(store.PostgresStore))

		Expect(cfg.Store.Postgres.Host).To(Equal("postgres.host"))
		Expect(int(cfg.Store.Postgres.Port)).To(Equal(5432))
		Expect(cfg.Store.Postgres.User).To(Equal("kuma"))
		Expect(cfg.Store.Postgres.Password).To(Equal("kuma"))
		Expect(cfg.Store.Postgres.DbName).To(Equal("kuma"))
		Expect(cfg.Store.Postgres.ConnectionTimeout).To(Equal(10))

		Expect(cfg.ApiServer.Port).To(Equal(9090))
		Expect(cfg.ApiServer.ReadOnly).To(Equal(true))
		Expect(cfg.ApiServer.ApiDocsPath).To(Equal("/apidocs.json"))
	})

	setEnv := func(key, value string) {
		err := os.Setenv(key, value)
		Expect(err).ToNot(HaveOccurred())
	}

	It("should load config from env vars", func() {
		// given
		setEnv("KUMA_XDS_SERVER_GRPC_PORT", "5000")
		setEnv("KUMA_XDS_SERVER_HTTP_PORT", "5001")
		setEnv("KUMA_XDS_SERVER_DIAGNOSTICS_PORT", "5003")
		setEnv("KUMA_BOOTSTRAP_SERVER_PORT", "5004")
		setEnv("KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT", "1234")
		setEnv("KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST", "kuma-control-plane")
		setEnv("KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_PORT", "4321")
		setEnv("KUMA_ENVIRONMENT", "kubernetes")
		setEnv("KUMA_STORE_TYPE", "postgres")
		setEnv("KUMA_STORE_POSTGRES_HOST", "postgres.host")
		setEnv("KUMA_STORE_POSTGRES_PORT", "5432")
		setEnv("KUMA_STORE_POSTGRES_USER", "kuma")
		setEnv("KUMA_STORE_POSTGRES_PASSWORD", "kuma")
		setEnv("KUMA_STORE_POSTGRES_DB_NAME", "kuma")
		setEnv("KUMA_STORE_POSTGRES_CONNECTION_TIMEOUT", "10")
		setEnv("KUMA_API_SERVER_READ_ONLY", "true")
		setEnv("KUMA_API_SERVER_PORT", "9090")
		setEnv("KUMA_API_SERVER_API_DOCS_PATH", "/apidocs.json")

		// when
		cfg := kuma_cp.DefaultConfig()
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfg.XdsServer.GrpcPort).To(Equal(5000))
		Expect(cfg.XdsServer.HttpPort).To(Equal(5001))
		Expect(cfg.XdsServer.DiagnosticsPort).To(Equal(5003))

		Expect(cfg.BootstrapServer.Port).To(Equal(5004))
		Expect(cfg.BootstrapServer.Params.AdminPort).To(Equal(uint32(1234)))
		Expect(cfg.BootstrapServer.Params.XdsHost).To(Equal("kuma-control-plane"))
		Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(4321)))

		Expect(cfg.Environment).To(Equal(kuma_cp.KubernetesEnvironment))

		Expect(cfg.Store.Type).To(Equal(store.PostgresStore))
		Expect(cfg.Store.Postgres.Host).To(Equal("postgres.host"))
		Expect(int(cfg.Store.Postgres.Port)).To(Equal(5432))
		Expect(cfg.Store.Postgres.User).To(Equal("kuma"))
		Expect(cfg.Store.Postgres.Password).To(Equal("kuma"))
		Expect(cfg.Store.Postgres.DbName).To(Equal("kuma"))
		Expect(cfg.Store.Postgres.ConnectionTimeout).To(Equal(10))

		Expect(cfg.ApiServer.Port).To(Equal(9090))
		Expect(cfg.ApiServer.ReadOnly).To(Equal(true))
		Expect(cfg.ApiServer.ApiDocsPath).To(Equal("/apidocs.json"))
	})

	It("should override via env var", func() {
		// given file with sample cfg
		file, err := ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
		_, err = file.WriteString(sampleConfigYaml)
		Expect(err).ToNot(HaveOccurred())

		// and overriden config
		setEnv("KUMA_STORE_POSTGRES_HOST", "overriden.host")

		// when
		cfg := kuma_cp.DefaultConfig()
		err = config.Load(file.Name(), &cfg)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfg.Store.Postgres.Host).To(Equal("overriden.host"))

	})

})
