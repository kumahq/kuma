package config_test

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"

	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"

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
			Expect(cfg.XdsServer.DiagnosticsPort).To(Equal(5003))

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
			Expect(cfg.Store.Postgres.MaxOpenConnections).To(Equal(300))

			Expect(cfg.Store.Cache.Enabled).To(BeFalse())
			Expect(cfg.Store.Cache.ExpirationTime).To(Equal(3 * time.Second))

			Expect(cfg.Store.Postgres.TLS.Mode).To(Equal(postgres.VerifyFull))
			Expect(cfg.Store.Postgres.TLS.CertPath).To(Equal("/path/to/cert"))
			Expect(cfg.Store.Postgres.TLS.KeyPath).To(Equal("/path/to/key"))
			Expect(cfg.Store.Postgres.TLS.CAPath).To(Equal("/path/to/rootCert"))

			Expect(cfg.ApiServer.ReadOnly).To(Equal(true))
			Expect(cfg.ApiServer.HTTP.Enabled).To(Equal(false))
			Expect(cfg.ApiServer.HTTP.Interface).To(Equal("192.168.0.1"))
			Expect(cfg.ApiServer.HTTP.Port).To(Equal(uint32(15681)))
			Expect(cfg.ApiServer.HTTPS.Enabled).To(Equal(false))
			Expect(cfg.ApiServer.HTTPS.Interface).To(Equal("192.168.0.2"))
			Expect(cfg.ApiServer.HTTPS.Port).To(Equal(uint32(15682)))
			Expect(cfg.ApiServer.HTTPS.TlsCertFile).To(Equal("/cert"))
			Expect(cfg.ApiServer.HTTPS.TlsKeyFile).To(Equal("/key"))
			Expect(cfg.ApiServer.Auth.AllowFromLocalhost).To(Equal(false))
			Expect(cfg.ApiServer.Auth.ClientCertsDir).To(Equal("/certs"))
			Expect(cfg.ApiServer.CorsAllowedDomains).To(Equal([]string{"https://kuma", "https://someapi"}))

			Expect(cfg.MonitoringAssignmentServer.GrpcPort).To(Equal(uint32(3333)))
			Expect(cfg.MonitoringAssignmentServer.AssignmentRefreshInterval).To(Equal(12 * time.Second))

			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Address).To(Equal("127.0.0.2"))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Port).To(Equal(uint32(9443)))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.CertDir).To(Equal("/var/run/secrets/kuma.io/kuma-admission-server/tls-cert"))

			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeInboundPorts).To(Equal([]uint32{1234, 5678}))
			Expect(cfg.Runtime.Kubernetes.Injector.CaCertFile).To(Equal("/tmp/ca.crt"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeOutboundPorts).To(Equal([]uint32{4321, 8765}))

			Expect(cfg.Reports.Enabled).To(BeFalse())

			Expect(cfg.General.AdvertisedHostname).To(Equal("kuma.internal"))
			Expect(cfg.General.TlsCertFile).To(Equal("/tmp/cert"))
			Expect(cfg.General.TlsKeyFile).To(Equal("/tmp/key"))

			Expect(cfg.GuiServer.ApiServerUrl).To(Equal("http://localhost:1234"))
			Expect(cfg.Mode).To(Equal(config_core.Remote))
			Expect(cfg.Multizone.Remote.Zone).To(Equal("zone-1"))

			Expect(cfg.Multizone.Global.PollTimeout).To(Equal(750 * time.Millisecond))
			Expect(cfg.Multizone.Global.KDS.GrpcPort).To(Equal(uint32(1234)))
			Expect(cfg.Multizone.Global.KDS.RefreshInterval).To(Equal(time.Second * 2))
			Expect(cfg.Multizone.Global.KDS.TlsCertFile).To(Equal("/cert"))
			Expect(cfg.Multizone.Global.KDS.TlsKeyFile).To(Equal("/key"))
			Expect(cfg.Multizone.Remote.GlobalAddress).To(Equal("grpc://1.1.1.1:5685"))
			Expect(cfg.Multizone.Remote.Zone).To(Equal("zone-1"))
			Expect(cfg.Multizone.Remote.KDS.RootCAFile).To(Equal("/rootCa"))

			Expect(cfg.Defaults.SkipMeshCreation).To(BeTrue())

			Expect(cfg.Diagnostics.DebugEndpoints).To(BeTrue())
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
    maxOpenConnections: 300
    tls:
      mode: verifyFull
      certPath: /path/to/cert
      keyPath: /path/to/key
      caPath: /path/to/rootCert
  cache:
    enabled: false
    expirationTime: 3s
xdsServer:
  diagnosticsPort: 5003
bootstrapServer:
  params:
    adminPort: 1234
    xdsHost: kuma-control-plane
    xdsPort: 4321
apiServer:
  http:
    enabled: false # ENV: KUMA_API_SERVER_HTTP_ENABLED
    interface: 192.168.0.1 # ENV: KUMA_API_SERVER_HTTP_INTERFACE
    port: 15681 # ENV: KUMA_API_SERVER_PORT
  https:
    enabled: false # ENV: KUMA_API_SERVER_HTTPS_ENABLED
    interface: 192.168.0.2 # ENV: KUMA_API_SERVER_HTTPS_INTERFACE
    port: 15682 # ENV: KUMA_API_SERVER_HTTPS_PORT
    tlsCertFile: "/cert" # ENV: KUMA_API_SERVER_HTTPS_TLS_CERT_FILE
    tlsKeyFile: "/key" # ENV: KUMA_API_SERVER_HTTPS_TLS_KEY_FILE
  auth:
    clientCertsDir: "/certs" # ENV: KUMA_API_SERVER_AUTH_CLIENT_CERTS_DIR
    allowFromLocalhost: false # ENV: KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST
  readOnly: true
  corsAllowedDomains:
    - https://kuma
    - https://someapi
monitoringAssignmentServer:
  grpcPort: 3333
  assignmentRefreshInterval: 12s
runtime:
  kubernetes:
    admissionServer:
      address: 127.0.0.2
      port: 9443
      certDir: /var/run/secrets/kuma.io/kuma-admission-server/tls-cert
    injector:
      caCertFile: /tmp/ca.crt
      sidecarTraffic:
        excludeInboundPorts:
        - 1234
        - 5678
        excludeOutboundPorts:
        - 4321
        - 8765
reports:
  enabled: false
general:
  advertisedHostname: kuma.internal
  tlsKeyFile: /tmp/key
  tlsCertFile: /tmp/cert
guiServer:
  apiServerUrl: http://localhost:1234
mode: remote
multizone:
  global:
    pollTimeout: 750ms
    kds:
      grpcPort: 1234
      refreshInterval: 2s
      tlsCertFile: /cert
      tlsKeyFile: /key
  remote:
    globalAddress: "grpc://1.1.1.1:5685"
    zone: "zone-1"
    kds:
      rootCaFile: /rootCa
dnsServer:
  port: 15653
  CIDR: 127.1.0.0/16
defaults:
  skipMeshCreation: true
diagnostics:
  debugEndpoints: true
`,
		}),
		Entry("from env variables", testCase{
			envVars: map[string]string{
				"KUMA_XDS_SERVER_DIAGNOSTICS_PORT":                               "5003",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT":                        "1234",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST":                          "kuma-control-plane",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_PORT":                          "4321",
				"KUMA_ENVIRONMENT":                                               "kubernetes",
				"KUMA_STORE_TYPE":                                                "postgres",
				"KUMA_STORE_POSTGRES_HOST":                                       "postgres.host",
				"KUMA_STORE_POSTGRES_PORT":                                       "5432",
				"KUMA_STORE_POSTGRES_USER":                                       "kuma",
				"KUMA_STORE_POSTGRES_PASSWORD":                                   "kuma",
				"KUMA_STORE_POSTGRES_DB_NAME":                                    "kuma",
				"KUMA_STORE_POSTGRES_CONNECTION_TIMEOUT":                         "10",
				"KUMA_STORE_POSTGRES_MAX_OPEN_CONNECTIONS":                       "300",
				"KUMA_STORE_POSTGRES_TLS_MODE":                                   "verifyFull",
				"KUMA_STORE_POSTGRES_TLS_CERT_PATH":                              "/path/to/cert",
				"KUMA_STORE_POSTGRES_TLS_KEY_PATH":                               "/path/to/key",
				"KUMA_STORE_POSTGRES_TLS_CA_PATH":                                "/path/to/rootCert",
				"KUMA_STORE_CACHE_ENABLED":                                       "false",
				"KUMA_STORE_CACHE_EXPIRATION_TIME":                               "3s",
				"KUMA_API_SERVER_READ_ONLY":                                      "true",
				"KUMA_API_SERVER_HTTP_PORT":                                      "15681",
				"KUMA_API_SERVER_HTTP_INTERFACE":                                 "192.168.0.1",
				"KUMA_API_SERVER_HTTP_ENABLED":                                   "false",
				"KUMA_API_SERVER_HTTPS_ENABLED":                                  "false",
				"KUMA_API_SERVER_HTTPS_PORT":                                     "15682",
				"KUMA_API_SERVER_HTTPS_INTERFACE":                                "192.168.0.2",
				"KUMA_API_SERVER_HTTPS_TLS_CERT_FILE":                            "/cert",
				"KUMA_API_SERVER_HTTPS_TLS_KEY_FILE":                             "/key",
				"KUMA_API_SERVER_AUTH_CLIENT_CERTS_DIR":                          "/certs",
				"KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST":                      "false",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_GRPC_PORT":                    "3333",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_ASSIGNMENT_REFRESH_INTERVAL":  "12s",
				"KUMA_REPORTS_ENABLED":                                           "false",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_ADDRESS":               "127.0.0.2",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_PORT":                  "9443",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_CERT_DIR":              "/var/run/secrets/kuma.io/kuma-admission-server/tls-cert",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_INBOUND_PORTS":  "1234,5678",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_OUTBOUND_PORTS": "4321,8765",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_CA_CERT_FILE":                  "/tmp/ca.crt",
				"KUMA_GENERAL_ADVERTISED_HOSTNAME":                               "kuma.internal",
				"KUMA_GENERAL_TLS_CERT_FILE":                                     "/tmp/cert",
				"KUMA_GENERAL_TLS_KEY_FILE":                                      "/tmp/key",
				"KUMA_API_SERVER_CORS_ALLOWED_DOMAINS":                           "https://kuma,https://someapi",
				"KUMA_GUI_SERVER_PORT":                                           "8888",
				"KUMA_GUI_SERVER_API_SERVER_URL":                                 "http://localhost:1234",
				"KUMA_DNS_SERVER_PORT":                                           "15653",
				"KUMA_DNS_CIDR":                                                  "127.1.0.0/16",
				"KUMA_MODE":                                                      "remote",
				"KUMA_MULTIZONE_GLOBAL_POLL_TIMEOUT":                             "750ms",
				"KUMA_MULTIZONE_GLOBAL_KDS_GRPC_PORT":                            "1234",
				"KUMA_MULTIZONE_GLOBAL_KDS_REFRESH_INTERVAL":                     "2s",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_CERT_FILE":                        "/cert",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_KEY_FILE":                         "/key",
				"KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS":                           "grpc://1.1.1.1:5685",
				"KUMA_MULTIZONE_REMOTE_ZONE":                                     "zone-1",
				"KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE":                         "/rootCa",
				"KUMA_DEFAULTS_SKIP_MESH_CREATION":                               "true",
				"KUMA_DIAGNOSTICS_DEBUG_ENDPOINTS":                               "true",
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
