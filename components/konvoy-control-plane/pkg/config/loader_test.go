package config

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
)

var _ = Describe("Config loader", func() {

	BeforeEach(func() {
		os.Clearenv()
	})

	sampleConfigYaml := `
grpcPort: 5000
httpPort: 5001
diagnosticsPort: 5003
environment: kubernetes
store:
  type: postgres
  postgres:
    host: postgres.host
    port: 5432
    user: konvoy
    password: konvoy
    dbName: konvoy
apiServer:
  readOnly: true
  bindAddress: "0.0.0.0:9090"
  apiDocsPath: "/apidocs.json"
`

	It("should load config from file", func() {
		// given file with sample config
		file, err := ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
		_, err = file.WriteString(sampleConfigYaml)
		Expect(err).ToNot(HaveOccurred())

		// when
		config, err := Load(file.Name())
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(config.GrpcPort).To(Equal(5000))
		Expect(config.HttpPort).To(Equal(5001))
		Expect(config.DiagnosticsPort).To(Equal(5003))

		Expect(config.Environment).To(Equal(KubernetesEnvironmentType))

		Expect(config.Store.Type).To(Equal(PostgresStoreType))

		Expect(config.Store.Postgres.Host).To(Equal("postgres.host"))
		Expect(int(config.Store.Postgres.Port)).To(Equal(5432))
		Expect(config.Store.Postgres.User).To(Equal("konvoy"))
		Expect(config.Store.Postgres.Password).To(Equal("konvoy"))
		Expect(config.Store.Postgres.DbName).To(Equal("konvoy"))

		Expect(config.ApiServer.BindAddress).To(Equal("0.0.0.0:9090"))
		Expect(config.ApiServer.ReadOnly).To(Equal(true))
		Expect(config.ApiServer.ApiDocsPath).To(Equal("/apidocs.json"))
	})

	setEnv := func(key, value string) {
		err := os.Setenv(key, value)
		Expect(err).ToNot(HaveOccurred())
	}

	It("should load config from env vars", func() {
		// given
		setEnv("KONVOY_GRPC_PORT", "5000")
		setEnv("KONVOY_HTTP_PORT", "5001")
		setEnv("KONVOY_DIAGNOSTICS_PORT", "5003")
		setEnv("KONVOY_ENVIRONMENT", "kubernetes")
		setEnv("KONVOY_STORE_TYPE", "postgres")
		setEnv("KONVOY_STORE_POSTGRES_HOST", "postgres.host")
		setEnv("KONVOY_STORE_POSTGRES_PORT", "5432")
		setEnv("KONVOY_STORE_POSTGRES_USER", "konvoy")
		setEnv("KONVOY_STORE_POSTGRES_PASSWORD", "konvoy")
		setEnv("KONVOY_STORE_POSTGRES_DB_NAME", "konvoy")
		setEnv("KONVOY_API_SERVER_READ_ONLY", "true")
		setEnv("KONVOY_API_SERVER_BIND_ADDRESS", "0.0.0.0:9090")
		setEnv("KONVOY_API_SERVER_API_DOCS_PATH", "/apidocs.json")

		// when
		config, err := Load("")
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(config.GrpcPort).To(Equal(5000))
		Expect(config.HttpPort).To(Equal(5001))
		Expect(config.DiagnosticsPort).To(Equal(5003))

		Expect(config.Environment).To(Equal(KubernetesEnvironmentType))

		Expect(config.Store.Type).To(Equal(PostgresStoreType))
		Expect(config.Store.Postgres.Host).To(Equal("postgres.host"))
		Expect(int(config.Store.Postgres.Port)).To(Equal(5432))
		Expect(config.Store.Postgres.User).To(Equal("konvoy"))
		Expect(config.Store.Postgres.Password).To(Equal("konvoy"))
		Expect(config.Store.Postgres.DbName).To(Equal("konvoy"))

		Expect(config.ApiServer.BindAddress).To(Equal("0.0.0.0:9090"))
		Expect(config.ApiServer.ReadOnly).To(Equal(true))
		Expect(config.ApiServer.ApiDocsPath).To(Equal("/apidocs.json"))
	})

	It("should override via env var", func() {
		// given file with sample config
		file, err := ioutil.TempFile("", "*")
		Expect(err).ToNot(HaveOccurred())
		_, err = file.WriteString(sampleConfigYaml)
		Expect(err).ToNot(HaveOccurred())

		// and overriden config
		setEnv("KONVOY_STORE_POSTGRES_HOST", "overriden.host")

		// when
		config, err := Load(file.Name())
		Expect(err).ToNot(HaveOccurred())
		err = loadFromEnv(config)
		Expect(err).ToNot(HaveOccurred())

		// then
		Expect(config.Store.Postgres.Host).To(Equal("overriden.host"))

	})

})
