package config_test

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	"github.com/kumahq/kuma/test/testenvconfig"
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

			if len(given.envVars) != 0 {
				infos, err := testenvconfig.GatherInfo("", &cfg)
				Expect(err).ToNot(HaveOccurred())

				configEnvs := map[string]bool{}
				for _, info := range infos {
					if info.Alt != "" {
						configEnvs[info.Alt] = true
					}
				}

				testEnvs := map[string]bool{}
				for key := range given.envVars {
					testEnvs[key] = true
				}

				Expect(testEnvs).To(Equal(configEnvs), "config values are not overridden in the test. Add overrides for them with a value that is different than default.")
			}

			// then
			Expect(cfg.BootstrapServer.APIVersion).To(Equal(envoy.APIV3))
			Expect(cfg.BootstrapServer.Params.AdminPort).To(Equal(uint32(1234)))
			Expect(cfg.BootstrapServer.Params.XdsHost).To(Equal("kuma-control-plane"))
			Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(4321)))
			Expect(cfg.BootstrapServer.Params.XdsConnectTimeout).To(Equal(13 * time.Second))
			Expect(cfg.BootstrapServer.Params.AdminAccessLogPath).To(Equal("/access/log/test"))
			Expect(cfg.BootstrapServer.Params.AdminAddress).To(Equal("1.1.1.1"))

			Expect(cfg.Environment).To(Equal(config_core.KubernetesEnvironment))

			Expect(cfg.Store.Type).To(Equal(store.PostgresStore))
			Expect(cfg.Store.Postgres.Host).To(Equal("postgres.host"))
			Expect(cfg.Store.Postgres.Port).To(Equal(5432))
			Expect(cfg.Store.Postgres.User).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.Password).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.DbName).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.ConnectionTimeout).To(Equal(10))
			Expect(cfg.Store.Postgres.MaxOpenConnections).To(Equal(300))
			Expect(cfg.Store.Postgres.MinReconnectInterval).To(Equal(44 * time.Second))
			Expect(cfg.Store.Postgres.MaxReconnectInterval).To(Equal(55 * time.Second))

			Expect(cfg.Store.Kubernetes.SystemNamespace).To(Equal("test-namespace"))

			Expect(cfg.Store.Cache.Enabled).To(BeFalse())
			Expect(cfg.Store.Cache.ExpirationTime).To(Equal(3 * time.Second))

			Expect(cfg.Store.Upsert.ConflictRetryBaseBackoff).To(Equal(4 * time.Second))
			Expect(cfg.Store.Upsert.ConflictRetryMaxTimes).To(Equal(uint(10)))

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

			Expect(cfg.Runtime.Kubernetes.ControlPlaneServiceName).To(Equal("custom-control-plane"))

			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Address).To(Equal("127.0.0.2"))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Port).To(Equal(uint32(9443)))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.CertDir).To(Equal("/var/run/secrets/kuma.io/kuma-admission-server/tls-cert"))
			Expect(cfg.Runtime.Kubernetes.MarshalingCacheExpirationTime).To(Equal(28 * time.Second))

			Expect(cfg.Runtime.Kubernetes.Injector.Exceptions.Labels).To(Equal(map[string]string{"openshift.io/build.name": "value1", "openshift.io/deployer-pod-for.name": "value2"}))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeInboundPorts).To(Equal([]uint32{1234, 5678}))
			Expect(cfg.Runtime.Kubernetes.Injector.CaCertFile).To(Equal("/tmp/ca.crt"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeOutboundPorts).To(Equal([]uint32{4321, 8765}))
			Expect(cfg.Runtime.Kubernetes.Injector.VirtualProbesEnabled).To(BeFalse())
			Expect(cfg.Runtime.Kubernetes.Injector.VirtualProbesPort).To(Equal(uint32(1111)))
			Expect(cfg.Runtime.Kubernetes.Injector.CNIEnabled).To(BeTrue())
			Expect(cfg.Runtime.Kubernetes.Injector.InitContainer.Image).To(Equal("test-image:test"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.EnvVars).To(Equal(map[string]string{"a": "b", "c": "d"}))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.RedirectPortInbound).To(Equal(uint32(2020)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.RedirectPortOutbound).To(Equal(uint32(1010)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.UID).To(Equal(int64(100)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.GID).To(Equal(int64(1212)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.Image).To(Equal("image:test"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort).To(Equal(uint32(1099)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.DrainTime).To(Equal(33 * time.Second))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.Resources.Requests.Memory).To(Equal("4Gi"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.Resources.Requests.CPU).To(Equal("123m"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.Resources.Limits.Memory).To(Equal("8Gi"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.Resources.Limits.CPU).To(Equal("100m"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.LivenessProbe.InitialDelaySeconds).To(Equal(int32(10)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.LivenessProbe.PeriodSeconds).To(Equal(int32(8)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.LivenessProbe.FailureThreshold).To(Equal(int32(31)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.LivenessProbe.TimeoutSeconds).To(Equal(int32(51)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.ReadinessProbe.SuccessThreshold).To(Equal(int32(17)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.ReadinessProbe.PeriodSeconds).To(Equal(int32(18)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.ReadinessProbe.FailureThreshold).To(Equal(int32(22)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.ReadinessProbe.TimeoutSeconds).To(Equal(int32(24)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.ReadinessProbe.InitialDelaySeconds).To(Equal(int32(41)))

			Expect(cfg.Runtime.Universal.DataplaneCleanupAge).To(Equal(1 * time.Hour))

			Expect(cfg.Reports.Enabled).To(BeFalse())

			Expect(cfg.General.TlsCertFile).To(Equal("/tmp/cert"))
			Expect(cfg.General.TlsKeyFile).To(Equal("/tmp/key"))
			Expect(cfg.General.DNSCacheTTL).To(Equal(19 * time.Second))
			Expect(cfg.General.WorkDir).To(Equal("/custom/work/dir"))

			Expect(cfg.GuiServer.ApiServerUrl).To(Equal("http://localhost:1234"))
			Expect(cfg.Mode).To(Equal(config_core.Remote))
			Expect(cfg.Multizone.Remote.Zone).To(Equal("zone-1"))

			Expect(cfg.Multizone.Global.PollTimeout).To(Equal(750 * time.Millisecond))
			Expect(cfg.Multizone.Global.KDS.GrpcPort).To(Equal(uint32(1234)))
			Expect(cfg.Multizone.Global.KDS.RefreshInterval).To(Equal(time.Second * 2))
			Expect(cfg.Multizone.Global.KDS.ZoneInsightFlushInterval).To(Equal(time.Second * 5))
			Expect(cfg.Multizone.Global.KDS.TlsCertFile).To(Equal("/cert"))
			Expect(cfg.Multizone.Global.KDS.TlsKeyFile).To(Equal("/key"))
			Expect(cfg.Multizone.Remote.GlobalAddress).To(Equal("grpc://1.1.1.1:5685"))
			Expect(cfg.Multizone.Remote.Zone).To(Equal("zone-1"))
			Expect(cfg.Multizone.Remote.KDS.RootCAFile).To(Equal("/rootCa"))
			Expect(cfg.Multizone.Remote.KDS.RefreshInterval).To(Equal(9 * time.Second))

			Expect(cfg.Defaults.SkipMeshCreation).To(BeTrue())

			Expect(cfg.Diagnostics.ServerPort).To(Equal(uint32(5003)))
			Expect(cfg.Diagnostics.DebugEndpoints).To(BeTrue())

			Expect(cfg.DNSServer.Domain).To(Equal("test-domain"))
			Expect(cfg.DNSServer.Port).To(Equal(uint32(15653)))
			Expect(cfg.DNSServer.CIDR).To(Equal("127.1.0.0/16"))

			Expect(cfg.XdsServer.DataplaneStatusFlushInterval).To(Equal(7 * time.Second))
			Expect(cfg.XdsServer.DataplaneConfigurationRefreshInterval).To(Equal(21 * time.Second))
			Expect(cfg.XdsServer.NACKBackoff).To(Equal(10 * time.Second))

			Expect(cfg.Metrics.Zone.Enabled).To(BeFalse())
			Expect(cfg.Metrics.Zone.SubscriptionLimit).To(Equal(23))
			Expect(cfg.Metrics.Mesh.MinResyncTimeout).To(Equal(35 * time.Second))
			Expect(cfg.Metrics.Mesh.MaxResyncTimeout).To(Equal(27 * time.Second))
			Expect(cfg.Metrics.Dataplane.Enabled).To(BeFalse())
			Expect(cfg.Metrics.Dataplane.SubscriptionLimit).To(Equal(47))

			Expect(cfg.DpServer.TlsCertFile).To(Equal("/test/path"))
			Expect(cfg.DpServer.TlsKeyFile).To(Equal("/test/path/key"))
			Expect(cfg.DpServer.Auth.Type).To(Equal("dpToken"))
			Expect(cfg.DpServer.Port).To(Equal(9876))
			Expect(cfg.DpServer.Hds.Enabled).To(BeFalse())
			Expect(cfg.DpServer.Hds.Interval).To(Equal(11 * time.Second))
			Expect(cfg.DpServer.Hds.RefreshInterval).To(Equal(12 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.Timeout).To(Equal(5 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.Interval).To(Equal(6 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.NoTrafficInterval).To(Equal(7 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.HealthyThreshold).To(Equal(uint32(8)))
			Expect(cfg.DpServer.Hds.CheckDefaults.UnhealthyThreshold).To(Equal(uint32(9)))

			Expect(cfg.SdsServer.DataplaneConfigurationRefreshInterval).To(Equal(11 * time.Second))
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
    minReconnectInterval: 44s
    maxReconnectInterval: 55s
    tls:
      mode: verifyFull
      certPath: /path/to/cert
      keyPath: /path/to/key
      caPath: /path/to/rootCert
  kubernetes:
    systemNamespace: test-namespace
  cache:
    enabled: false
    expirationTime: 3s
  upsert:
    conflictRetryBaseBackoff: 4s
    conflictRetryMaxTimes: 10
bootstrapServer:
  apiVersion: v3
  params:
    adminPort: 1234
    adminAccessLogPath: /access/log/test
    adminAddress: 1.1.1.1
    xdsHost: kuma-control-plane
    xdsPort: 4321
    xdsConnectTimeout: 13s
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
  universal:
    dataplaneCleanupAge: 1h
  kubernetes:
    controlPlaneServiceName: custom-control-plane
    admissionServer:
      address: 127.0.0.2
      port: 9443
      certDir: /var/run/secrets/kuma.io/kuma-admission-server/tls-cert
    marshalingCacheExpirationTime: 28s
    injector:
      exceptions:
        labels:
          openshift.io/build.name: value1
          openshift.io/deployer-pod-for.name: value2
      cniEnabled: true
      caCertFile: /tmp/ca.crt
      virtualProbesEnabled: false
      virtualProbesPort: 1111
      initContainer:
        image: test-image:test
      sidecarContainer:
        image: image:test
        redirectPortInbound: 2020
        redirectPortOutbound: 1010
        uid: 100
        gid: 1212
        adminPort: 1099
        drainTime: 33s
        resources:
          requests:
            memory: 4Gi
            cpu: 123m
          limits:
            memory: 8Gi
            cpu: 100m
        livenessProbe:
          initialDelaySeconds: 10
          periodSeconds: 8
          failureThreshold: 31
          timeoutSeconds: 51
        readinessProbe:
          initialDelaySeconds: 41
          successThreshold: 17
          periodSeconds: 18
          failureThreshold: 22
          timeoutSeconds: 24
        envVars:
          a: b
          c: d
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
  tlsKeyFile: /tmp/key
  tlsCertFile: /tmp/cert
  dnsCacheTTL: 19s
  workDir: /custom/work/dir
guiServer:
  apiServerUrl: http://localhost:1234
mode: remote
multizone:
  global:
    pollTimeout: 750ms
    kds:
      grpcPort: 1234
      refreshInterval: 2s
      zoneInsightFlushInterval: 5s
      tlsCertFile: /cert
      tlsKeyFile: /key
  remote:
    globalAddress: "grpc://1.1.1.1:5685"
    zone: "zone-1"
    kds:
      refreshInterval: 9s
      rootCaFile: /rootCa
dnsServer:
  domain: test-domain
  port: 15653
  CIDR: 127.1.0.0/16
defaults:
  skipMeshCreation: true
diagnostics:
  serverPort: 5003
  debugEndpoints: true
xdsServer:
  dataplaneConfigurationRefreshInterval: 21s
  dataplaneStatusFlushInterval: 7s
  nackBackoff: 10s
metrics:
  zone:
    enabled: false
    subscriptionLimit: 23
  mesh:
    minResyncTimeout: 35s
    maxResyncTimeout: 27s
  dataplane:
    subscriptionLimit: 47
    enabled: false
dpServer:
  tlsCertFile: /test/path
  tlsKeyFile: /test/path/key
  port: 9876
  auth:
    type: dpToken
  hds:
    enabled: false
    interval: 11s
    refreshInterval: 12s
    checkDefaults:
      timeout: 5s
      interval: 6s
      noTrafficInterval: 7s
      healthyThreshold: 8
      unhealthyThreshold: 9
sdsServer:
  dataplaneConfigurationRefreshInterval: 11s
`,
		}),
		Entry("from env variables", testCase{
			envVars: map[string]string{
				"KUMA_BOOTSTRAP_SERVER_API_VERSION":                                                        "v3",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT":                                                  "1234",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST":                                                    "kuma-control-plane",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_PORT":                                                    "4321",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_CONNECT_TIMEOUT":                                         "13s",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ACCESS_LOG_PATH":                                       "/access/log/test",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ADDRESS":                                               "1.1.1.1",
				"KUMA_ENVIRONMENT":                                                                         "kubernetes",
				"KUMA_STORE_TYPE":                                                                          "postgres",
				"KUMA_STORE_POSTGRES_HOST":                                                                 "postgres.host",
				"KUMA_STORE_POSTGRES_PORT":                                                                 "5432",
				"KUMA_STORE_POSTGRES_USER":                                                                 "kuma",
				"KUMA_STORE_POSTGRES_PASSWORD":                                                             "kuma",
				"KUMA_STORE_POSTGRES_DB_NAME":                                                              "kuma",
				"KUMA_STORE_POSTGRES_CONNECTION_TIMEOUT":                                                   "10",
				"KUMA_STORE_POSTGRES_MAX_OPEN_CONNECTIONS":                                                 "300",
				"KUMA_STORE_POSTGRES_TLS_MODE":                                                             "verifyFull",
				"KUMA_STORE_POSTGRES_TLS_CERT_PATH":                                                        "/path/to/cert",
				"KUMA_STORE_POSTGRES_TLS_KEY_PATH":                                                         "/path/to/key",
				"KUMA_STORE_POSTGRES_TLS_CA_PATH":                                                          "/path/to/rootCert",
				"KUMA_STORE_POSTGRES_MIN_RECONNECT_INTERVAL":                                               "44s",
				"KUMA_STORE_POSTGRES_MAX_RECONNECT_INTERVAL":                                               "55s",
				"KUMA_STORE_KUBERNETES_SYSTEM_NAMESPACE":                                                   "test-namespace",
				"KUMA_STORE_CACHE_ENABLED":                                                                 "false",
				"KUMA_STORE_CACHE_EXPIRATION_TIME":                                                         "3s",
				"KUMA_STORE_UPSERT_CONFLICT_RETRY_BASE_BACKOFF":                                            "4s",
				"KUMA_STORE_UPSERT_CONFLICT_RETRY_MAX_TIMES":                                               "10",
				"KUMA_API_SERVER_READ_ONLY":                                                                "true",
				"KUMA_API_SERVER_HTTP_PORT":                                                                "15681",
				"KUMA_API_SERVER_HTTP_INTERFACE":                                                           "192.168.0.1",
				"KUMA_API_SERVER_HTTP_ENABLED":                                                             "false",
				"KUMA_API_SERVER_HTTPS_ENABLED":                                                            "false",
				"KUMA_API_SERVER_HTTPS_PORT":                                                               "15682",
				"KUMA_API_SERVER_HTTPS_INTERFACE":                                                          "192.168.0.2",
				"KUMA_API_SERVER_HTTPS_TLS_CERT_FILE":                                                      "/cert",
				"KUMA_API_SERVER_HTTPS_TLS_KEY_FILE":                                                       "/key",
				"KUMA_API_SERVER_AUTH_CLIENT_CERTS_DIR":                                                    "/certs",
				"KUMA_API_SERVER_AUTH_ALLOW_FROM_LOCALHOST":                                                "false",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_GRPC_PORT":                                              "3333",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_ASSIGNMENT_REFRESH_INTERVAL":                            "12s",
				"KUMA_REPORTS_ENABLED":                                                                     "false",
				"KUMA_RUNTIME_KUBERNETES_CONTROL_PLANE_SERVICE_NAME":                                       "custom-control-plane",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_ADDRESS":                                         "127.0.0.2",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_PORT":                                            "9443",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_CERT_DIR":                                        "/var/run/secrets/kuma.io/kuma-admission-server/tls-cert",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_INBOUND_PORTS":                            "1234,5678",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_OUTBOUND_PORTS":                           "4321,8765",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_CA_CERT_FILE":                                            "/tmp/ca.crt",
				"KUMA_RUNTIME_KUBERNETES_MARSHALING_CACHE_EXPIRATION_TIME":                                 "28s",
				"KUMA_INJECTOR_INIT_CONTAINER_IMAGE":                                                       "test-image:test",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_REQUESTS_MEMORY":                                "4Gi",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_REQUESTS_CPU":                                   "123m",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_LIMITS_MEMORY":                                  "8Gi",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_LIMITS_CPU":                                     "100m",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_REDIRECT_PORT_INBOUND":                 "2020",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_REDIRECT_PORT_OUTBOUND":                "1010",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_CNI_ENABLED":                                             "true",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ENV_VARS":                              "a:b,c:d",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_UID":                                   "100",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_ADMIN_PORT":                            "1099",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_DRAIN_TIME":                            "33s",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_GUI":                                   "1212",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IMAGE":                                 "image:test",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_LIVENESS_PROBE_INITIAL_DELAY_SECONDS":  "10",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_LIVENESS_PROBE_PERIOD_SECONDS":         "8",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_LIVENESS_PROBE_FAILURE_THRESHOLD":      "31",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_LIVENESS_PROBE_TIMEOUT_SECONDS":        "51",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_READINESS_PROBE_SUCCESS_THRESHOLD":     "17",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_READINESS_PROBE_PERIOD_SECONDS":        "18",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_READINESS_PROBE_FAILURE_THRESHOLD":     "22",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_READINESS_PROBE_TIMEOUT_SECONDS":       "24",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_READINESS_PROBE_INITIAL_DELAY_SECONDS": "41",
				"KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_ENABLED":                                           "false",
				"KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_PORT":                                              "1111",
				"KUMA_RUNTIME_KUBERNETES_EXCEPTIONS_LABELS":                                                "openshift.io/build.name:value1,openshift.io/deployer-pod-for.name:value2",
				"KUMA_RUNTIME_UNIVERSAL_DATAPLANE_CLEANUP_AGE":                                             "1h",
				"KUMA_GENERAL_TLS_CERT_FILE":                                                               "/tmp/cert",
				"KUMA_GENERAL_TLS_KEY_FILE":                                                                "/tmp/key",
				"KUMA_GENERAL_DNS_CACHE_TTL":                                                               "19s",
				"KUMA_GENERAL_WORK_DIR":                                                                    "/custom/work/dir",
				"KUMA_API_SERVER_CORS_ALLOWED_DOMAINS":                                                     "https://kuma,https://someapi",
				"KUMA_GUI_SERVER_API_SERVER_URL":                                                           "http://localhost:1234",
				"KUMA_DNS_SERVER_DOMAIN":                                                                   "test-domain",
				"KUMA_DNS_SERVER_PORT":                                                                     "15653",
				"KUMA_DNS_SERVER_CIDR":                                                                     "127.1.0.0/16",
				"KUMA_MODE":                                                                                "remote",
				"KUMA_MULTIZONE_GLOBAL_POLL_TIMEOUT":                                                       "750ms",
				"KUMA_MULTIZONE_GLOBAL_KDS_GRPC_PORT":                                                      "1234",
				"KUMA_MULTIZONE_GLOBAL_KDS_REFRESH_INTERVAL":                                               "2s",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_CERT_FILE":                                                  "/cert",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_KEY_FILE":                                                   "/key",
				"KUMA_MULTIZONE_REMOTE_GLOBAL_ADDRESS":                                                     "grpc://1.1.1.1:5685",
				"KUMA_MULTIZONE_REMOTE_ZONE":                                                               "zone-1",
				"KUMA_MULTIZONE_REMOTE_KDS_ROOT_CA_FILE":                                                   "/rootCa",
				"KUMA_MULTIZONE_REMOTE_KDS_REFRESH_INTERVAL":                                               "9s",
				"KUMA_MULTIZONE_GLOBAL_KDS_ZONE_INSIGHT_FLUSH_INTERVAL":                                    "5s",
				"KUMA_DEFAULTS_SKIP_MESH_CREATION":                                                         "true",
				"KUMA_DIAGNOSTICS_SERVER_PORT":                                                             "5003",
				"KUMA_DIAGNOSTICS_DEBUG_ENDPOINTS":                                                         "true",
				"KUMA_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL":                                          "7s",
				"KUMA_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL":                                 "21s",
				"KUMA_XDS_SERVER_NACK_BACKOFF":                                                             "10s",
				"KUMA_METRICS_ZONE_ENABLED":                                                                "false",
				"KUMA_METRICS_ZONE_SUBSCRIPTION_LIMIT":                                                     "23",
				"KUMA_METRICS_MESH_MAX_RESYNC_TIMEOUT":                                                     "27s",
				"KUMA_METRICS_DATAPLANE_ENABLED":                                                           "false",
				"KUMA_METRICS_MESH_MIN_RESYNC_TIMEOUT":                                                     "35s",
				"KUMA_METRICS_DATAPLANE_SUBSCRIPTION_LIMIT":                                                "47",
				"KUMA_DP_SERVER_TLS_CERT_FILE":                                                             "/test/path",
				"KUMA_DP_SERVER_TLS_KEY_FILE":                                                              "/test/path/key",
				"KUMA_DP_SERVER_AUTH_TYPE":                                                                 "dpToken",
				"KUMA_DP_SERVER_PORT":                                                                      "9876",
				"KUMA_DP_SERVER_HDS_ENABLED":                                                               "false",
				"KUMA_DP_SERVER_HDS_INTERVAL":                                                              "11s",
				"KUMA_DP_SERVER_HDS_REFRESH_INTERVAL":                                                      "12s",
				"KUMA_DP_SERVER_HDS_CHECK_TIMEOUT":                                                         "5s",
				"KUMA_DP_SERVER_HDS_CHECK_INTERVAL":                                                        "6s",
				"KUMA_DP_SERVER_HDS_CHECK_NO_TRAFFIC_INTERVAL":                                             "7s",
				"KUMA_DP_SERVER_HDS_CHECK_HEALTHY_THRESHOLD":                                               "8",
				"KUMA_DP_SERVER_HDS_CHECK_UNHEALTHY_THRESHOLD":                                             "9",
				"KUMA_SDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL":                                 "11s",
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
