package config_test

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config"
	konvoy_cp "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/app/konvoy-cp"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/config/core/resources/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
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
    user: konvoy
    password: konvoy
    dbName: konvoy
xdsServer:
  grpcPort: 5000
  httpPort: 5001
  diagnosticsPort: 5003
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
		cfg := konvoy_cp.DefaultConfig()
		err = config.Load(file.Name(), &cfg)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfg.XdsServerConfig.GrpcPort).To(Equal(5000))
		Expect(cfg.XdsServerConfig.HttpPort).To(Equal(5001))
		Expect(cfg.XdsServerConfig.DiagnosticsPort).To(Equal(5003))

		Expect(cfg.Environment).To(Equal(konvoy_cp.KubernetesEnvironmentType))

		Expect(cfg.Store.Type).To(Equal(store.PostgresStoreType))

		Expect(cfg.Store.Postgres.Host).To(Equal("postgres.host"))
		Expect(int(cfg.Store.Postgres.Port)).To(Equal(5432))
		Expect(cfg.Store.Postgres.User).To(Equal("konvoy"))
		Expect(cfg.Store.Postgres.Password).To(Equal("konvoy"))
		Expect(cfg.Store.Postgres.DbName).To(Equal("konvoy"))

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
		setEnv("KONVOY_XDS_SERVER_GRPC_PORT", "5000")
		setEnv("KONVOY_XDS_SERVER_HTTP_PORT", "5001")
		setEnv("KONVOY_XDS_SERVER_DIAGNOSTICS_PORT", "5003")
		setEnv("KONVOY_ENVIRONMENT", "kubernetes")
		setEnv("KONVOY_STORE_TYPE", "postgres")
		setEnv("KONVOY_STORE_POSTGRES_HOST", "postgres.host")
		setEnv("KONVOY_STORE_POSTGRES_PORT", "5432")
		setEnv("KONVOY_STORE_POSTGRES_USER", "konvoy")
		setEnv("KONVOY_STORE_POSTGRES_PASSWORD", "konvoy")
		setEnv("KONVOY_STORE_POSTGRES_DB_NAME", "konvoy")
		setEnv("KONVOY_API_SERVER_READ_ONLY", "true")
		setEnv("KONVOY_API_SERVER_PORT", "9090")
		setEnv("KONVOY_API_SERVER_API_DOCS_PATH", "/apidocs.json")

		// when
		cfg := konvoy_cp.DefaultConfig()
		err := config.Load("", &cfg)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfg.XdsServerConfig.GrpcPort).To(Equal(5000))
		Expect(cfg.XdsServerConfig.HttpPort).To(Equal(5001))
		Expect(cfg.XdsServerConfig.DiagnosticsPort).To(Equal(5003))

		Expect(cfg.Environment).To(Equal(konvoy_cp.KubernetesEnvironmentType))

		Expect(cfg.Store.Type).To(Equal(store.PostgresStoreType))
		Expect(cfg.Store.Postgres.Host).To(Equal("postgres.host"))
		Expect(int(cfg.Store.Postgres.Port)).To(Equal(5432))
		Expect(cfg.Store.Postgres.User).To(Equal("konvoy"))
		Expect(cfg.Store.Postgres.Password).To(Equal("konvoy"))
		Expect(cfg.Store.Postgres.DbName).To(Equal("konvoy"))

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
		setEnv("KONVOY_STORE_POSTGRES_HOST", "overriden.host")

		// when
		cfg := konvoy_cp.DefaultConfig()
		err = config.Load(file.Name(), &cfg)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(cfg.Store.Postgres.Host).To(Equal("overriden.host"))

	})

})
