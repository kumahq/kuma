package config_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/maps"

	"github.com/kumahq/kuma/pkg/config"
	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/postgres"
	"github.com/kumahq/kuma/test/testenvconfig"
)

var _ = Describe("Config loader", func() {
	var configFile *os.File

	BeforeEach(func() {
		os.Clearenv()
		file, err := os.CreateTemp("", "*")
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
			file, err := os.CreateTemp("", "*")
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
			if len(given.envVars) != 0 {
				infos, err := testenvconfig.GatherInfo("", &cfg)
				Expect(err).ToNot(HaveOccurred())

				configEnvs := map[string]struct{}{}
				for _, info := range infos {
					if info.Alt != "" {
						configEnvs[info.Alt] = struct{}{}
					}
				}

				testEnvs := map[string]struct{}{}
				for key := range given.envVars {
					testEnvs[key] = struct{}{}
				}

				Expect(maps.Keys(testEnvs)).To(ConsistOf(maps.Keys(configEnvs)), "config values are not overridden in the test. Add overrides for them with a value that is different than default.")
			}

			Expect(cfg.BootstrapServer.Params.AdminPort).To(Equal(uint32(1234)))
			Expect(cfg.BootstrapServer.Params.XdsHost).To(Equal("kuma-control-plane"))
			Expect(cfg.BootstrapServer.Params.XdsPort).To(Equal(uint32(4321)))
			Expect(cfg.BootstrapServer.Params.XdsConnectTimeout.Duration).To(Equal(13 * time.Second))
			Expect(cfg.BootstrapServer.Params.AdminAccessLogPath).To(Equal("/access/log/test"))
			Expect(cfg.BootstrapServer.Params.AdminAddress).To(Equal("1.1.1.1"))

			Expect(cfg.Environment).To(Equal(config_core.KubernetesEnvironment))

			Expect(cfg.Store.Type).To(Equal(store.PostgresStore))
			Expect(cfg.Store.UnsafeDelete).To(BeTrue())
			Expect(cfg.Store.Postgres.Host).To(Equal("postgres.host"))
			Expect(cfg.Store.Postgres.Port).To(Equal(5432))
			Expect(cfg.Store.Postgres.User).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.Password).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.DbName).To(Equal("kuma"))
			Expect(cfg.Store.Postgres.DriverName).To(Equal("postgres"))
			Expect(cfg.Store.Postgres.ConnectionTimeout).To(Equal(10))
			Expect(cfg.Store.Postgres.MaxConnectionLifetime.Duration).To(Equal(123 * time.Minute))
			Expect(cfg.Store.Postgres.MaxConnectionLifetimeJitter.Duration).To(Equal(456 * time.Second))
			Expect(cfg.Store.Postgres.HealthCheckInterval.Duration).To(Equal(78 * time.Second))
			Expect(cfg.Store.Postgres.MinOpenConnections).To(Equal(3))
			Expect(cfg.Store.Postgres.MaxOpenConnections).To(Equal(300))
			Expect(cfg.Store.Postgres.MaxIdleConnections).To(Equal(300))
			Expect(cfg.Store.Postgres.MaxListQueryElements).To(Equal(uint32(111)))
			Expect(cfg.Store.Postgres.MaxConnectionIdleTime.Duration).To(Equal(99 * time.Second))

			Expect(cfg.Store.Kubernetes.SystemNamespace).To(Equal("test-namespace"))

			Expect(cfg.Store.Cache.Enabled).To(BeFalse())
			Expect(cfg.Store.Cache.ExpirationTime.Duration).To(Equal(3 * time.Second))

			Expect(cfg.Store.Upsert.ConflictRetryBaseBackoff.Duration).To(Equal(4 * time.Second))
			Expect(cfg.Store.Upsert.ConflictRetryMaxTimes).To(Equal(uint(15)))
			Expect(cfg.Store.Upsert.ConflictRetryJitterPercent).To(Equal(uint(10)))

			Expect(cfg.Store.Postgres.TLS.Mode).To(Equal(postgres.VerifyFull))
			Expect(cfg.Store.Postgres.TLS.CertPath).To(Equal("/path/to/cert"))
			Expect(cfg.Store.Postgres.TLS.KeyPath).To(Equal("/path/to/key"))
			Expect(cfg.Store.Postgres.TLS.CAPath).To(Equal("/path/to/rootCert"))
			Expect(cfg.Store.Postgres.TLS.DisableSSLSNI).To(BeTrue())

			Expect(cfg.Store.Postgres.ReadReplica.Host).To(Equal("ro.host"))
			Expect(cfg.Store.Postgres.ReadReplica.Port).To(Equal(uint(35432)))
			Expect(cfg.Store.Postgres.ReadReplica.Ratio).To(Equal(uint(80)))

			Expect(cfg.ApiServer.ReadOnly).To(BeTrue())
			Expect(cfg.ApiServer.HTTP.Enabled).To(BeFalse())
			Expect(cfg.ApiServer.HTTP.Interface).To(Equal("192.168.0.1"))
			Expect(cfg.ApiServer.HTTP.Port).To(Equal(uint32(15681)))
			Expect(cfg.ApiServer.HTTPS.Enabled).To(BeFalse())
			Expect(cfg.ApiServer.HTTPS.Interface).To(Equal("192.168.0.2"))
			Expect(cfg.ApiServer.HTTPS.Port).To(Equal(uint32(15682)))
			Expect(cfg.ApiServer.HTTPS.TlsCertFile).To(Equal("/cert"))
			Expect(cfg.ApiServer.HTTPS.TlsKeyFile).To(Equal("/key"))
			Expect(cfg.ApiServer.HTTPS.TlsCaFile).To(Equal("/ca"))
			Expect(cfg.ApiServer.HTTPS.TlsMinVersion).To(Equal("TLSv1_3"))
			Expect(cfg.ApiServer.HTTPS.TlsMaxVersion).To(Equal("TLSv1_3"))
			Expect(cfg.ApiServer.HTTPS.TlsCipherSuites).To(Equal([]string{"TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"}))
			Expect(cfg.ApiServer.HTTPS.RequireClientCert).To(BeTrue())
			Expect(cfg.ApiServer.Auth.ClientCertsDir).To(Equal("/certs"))
			Expect(cfg.ApiServer.Authn.LocalhostIsAdmin).To(BeFalse())
			Expect(cfg.ApiServer.Authn.Type).To(Equal("custom-authn"))
			Expect(cfg.ApiServer.Authn.Tokens.BootstrapAdminToken).To(BeFalse())
			Expect(cfg.ApiServer.Authn.Tokens.EnableIssuer).To(BeFalse())
			Expect(cfg.ApiServer.Authn.Tokens.Validator.UseSecrets).To(BeFalse())
			Expect(cfg.ApiServer.CorsAllowedDomains).To(Equal([]string{"https://kuma", "https://someapi"}))
			Expect(cfg.ApiServer.BasePath).To(Equal("/api"))
			Expect(cfg.ApiServer.RootUrl).To(Equal("https://foo.com"))
			Expect(cfg.ApiServer.GUI.RootUrl).To(Equal("https://bar.com"))
			Expect(cfg.ApiServer.GUI.Enabled).To(BeFalse())
			Expect(cfg.ApiServer.GUI.BasePath).To(Equal("/ui"))

			Expect(cfg.MonitoringAssignmentServer.Port).To(Equal(uint32(2222)))
			Expect(cfg.MonitoringAssignmentServer.AssignmentRefreshInterval.Duration).To(Equal(12 * time.Second))
			Expect(cfg.MonitoringAssignmentServer.DefaultFetchTimeout.Duration).To(Equal(45 * time.Second))
			Expect(cfg.MonitoringAssignmentServer.ApiVersions).To(HaveLen(1))
			Expect(cfg.MonitoringAssignmentServer.ApiVersions).To(ContainElements("v1"))
			Expect(cfg.MonitoringAssignmentServer.TlsEnabled).To(BeTrue())
			Expect(cfg.MonitoringAssignmentServer.TlsCertFile).To(Equal("/cert"))
			Expect(cfg.MonitoringAssignmentServer.TlsKeyFile).To(Equal("/key"))
			Expect(cfg.MonitoringAssignmentServer.TlsMinVersion).To(Equal("TLSv1_3"))
			Expect(cfg.MonitoringAssignmentServer.TlsMaxVersion).To(Equal("TLSv1_3"))
			Expect(cfg.MonitoringAssignmentServer.TlsCipherSuites).To(Equal([]string{"TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"}))

			Expect(cfg.Runtime.Kubernetes.ControlPlaneServiceName).To(Equal("custom-control-plane"))
			Expect(cfg.Runtime.Kubernetes.ServiceAccountName).To(Equal("custom-sa"))
			Expect(cfg.Runtime.Kubernetes.AllowedUsers).To(Equal([]string{"allowed-usr-1", "allowed-usr-2"}))
			Expect(cfg.Runtime.Kubernetes.NodeTaintController.Enabled).To(BeTrue())
			Expect(cfg.Runtime.Kubernetes.NodeTaintController.CniApp).To(Equal("kuma-cni"))

			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Address).To(Equal("127.0.0.2"))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.Port).To(Equal(uint32(9443)))
			Expect(cfg.Runtime.Kubernetes.AdmissionServer.CertDir).To(Equal("/var/run/secrets/kuma.io/kuma-admission-server/tls-cert"))
			Expect(cfg.Runtime.Kubernetes.MarshalingCacheExpirationTime.Duration).To(Equal(28 * time.Second))

			Expect(cfg.Runtime.Kubernetes.Injector.Exceptions.Labels).To(Equal(map[string]string{"openshift.io/build.name": "value1", "openshift.io/deployer-pod-for.name": "value2"}))
			Expect(cfg.Runtime.Kubernetes.Injector.CaCertFile).To(Equal("/tmp/ca.crt"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeInboundPorts).To(Equal([]uint32{1234, 5678}))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeOutboundPorts).To(Equal([]uint32{4321, 8765}))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeInboundIPs).To(Equal([]string{"192.168.0.1", "172.32.16.8/16", "a81b:a033:6399:73c7:72b6:aa8c:6f22:7098", "fe80::/10"}))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarTraffic.ExcludeOutboundIPs).To(Equal([]string{"10.0.0.1", "172.16.0.0/16", "fe80::1", "fe80::/10"}))
			Expect(cfg.Runtime.Kubernetes.Injector.VirtualProbesEnabled).To(BeFalse())
			Expect(cfg.Runtime.Kubernetes.Injector.VirtualProbesPort).To(Equal(uint32(1111)))
			Expect(cfg.Runtime.Kubernetes.Injector.ApplicationProbeProxyPort).To(Equal(uint32(1112)))
			Expect(cfg.Runtime.Kubernetes.Injector.CNIEnabled).To(BeTrue())
			Expect(cfg.Runtime.Kubernetes.Injector.ContainerPatches).To(Equal([]string{"patch1", "patch2"}))
			Expect(cfg.Runtime.Kubernetes.Injector.InitContainer.Image).To(Equal("test-image:test"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.EnvVars).To(Equal(map[string]string{"a": "b", "c": "d"}))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.IpFamilyMode).To(Equal("dualstack"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.RedirectPortInbound).To(Equal(uint32(2020)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.RedirectPortOutbound).To(Equal(uint32(1010)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.UID).To(Equal(int64(100)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.GID).To(Equal(int64(1212)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.Image).To(Equal("image:test"))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.AdminPort).To(Equal(uint32(1099)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.DrainTime.Duration).To(Equal(33 * time.Second))
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
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.StartupProbe.InitialDelaySeconds).To(Equal(int32(11)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.StartupProbe.PeriodSeconds).To(Equal(int32(9)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.StartupProbe.FailureThreshold).To(Equal(int32(32)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.StartupProbe.TimeoutSeconds).To(Equal(int32(52)))
			Expect(cfg.Runtime.Kubernetes.Injector.SidecarContainer.WaitForDataplaneReady).To(BeTrue())
			Expect(cfg.Runtime.Kubernetes.Injector.BuiltinDNS.Enabled).To(BeTrue())
			Expect(cfg.Runtime.Kubernetes.Injector.BuiltinDNS.Port).To(Equal(uint32(1053)))
			Expect(cfg.Runtime.Kubernetes.Injector.EBPF.Enabled).To(BeTrue())
			Expect(cfg.Runtime.Kubernetes.Injector.EBPF.InstanceIPEnvVarName).To(Equal("FOO"))
			Expect(cfg.Runtime.Kubernetes.Injector.EBPF.BPFFSPath).To(Equal("/run/kuma/bar"))
			Expect(cfg.Runtime.Kubernetes.Injector.EBPF.CgroupPath).To(Equal("/faz/daz/zaz"))
			Expect(cfg.Runtime.Kubernetes.Injector.EBPF.TCAttachIface).To(Equal("veth1"))
			Expect(cfg.Runtime.Kubernetes.Injector.EBPF.ProgramsSourcePath).To(Equal("/kuma/baz"))
			Expect(cfg.Runtime.Kubernetes.Injector.IgnoredServiceSelectorLabels).To(Equal([]string{"x", "y"}))
			Expect(cfg.Runtime.Kubernetes.Injector.NodeLabelsToCopy).To(Equal([]string{"label-1", "label-2"}))
			Expect(cfg.Runtime.Kubernetes.Injector.TransparentProxyConfigMapName).To(Equal("foo"))
			Expect(cfg.Runtime.Kubernetes.NodeTaintController.CniNamespace).To(Equal("kuma-system"))
			Expect(cfg.Runtime.Kubernetes.ControllersConcurrency.PodController).To(Equal(10))
			Expect(cfg.Runtime.Kubernetes.ClientConfig.Qps).To(Equal(100))
			Expect(cfg.Runtime.Kubernetes.ClientConfig.BurstQps).To(Equal(100))
			Expect(cfg.Runtime.Kubernetes.LeaderElection.LeaseDuration.Duration).To(Equal(199 * time.Second))
			Expect(cfg.Runtime.Kubernetes.LeaderElection.RenewDeadline.Duration).To(Equal(99 * time.Second))
			Expect(cfg.Runtime.Kubernetes.SkipMeshOwnerReference).To(BeTrue())
			Expect(cfg.Runtime.Kubernetes.SupportGatewaySecretsInAllNamespaces).To(BeTrue())

			Expect(cfg.Runtime.Universal.DataplaneCleanupAge.Duration).To(Equal(1 * time.Hour))
			Expect(cfg.Runtime.Universal.VIPRefreshInterval.Duration).To(Equal(10 * time.Second))

			Expect(cfg.Reports.Enabled).To(BeFalse())

			Expect(cfg.General.TlsCertFile).To(Equal("/tmp/cert"))
			Expect(cfg.General.TlsKeyFile).To(Equal("/tmp/key"))
			Expect(cfg.General.TlsMinVersion).To(Equal("TLSv1_3"))
			Expect(cfg.General.TlsMaxVersion).To(Equal("TLSv1_3"))
			Expect(cfg.General.TlsCipherSuites).To(Equal([]string{"TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"}))
			Expect(cfg.General.DNSCacheTTL.Duration).To(Equal(19 * time.Second))
			Expect(cfg.General.WorkDir).To(Equal("/custom/work/dir"))

			Expect(cfg.Mode).To(Equal(config_core.Zone))
			Expect(cfg.Multizone.Zone.Name).To(Equal("zone-1"))

			Expect(cfg.Multizone.Global.KDS.GrpcPort).To(Equal(uint32(1234)))
			Expect(cfg.Multizone.Global.KDS.RefreshInterval.Duration).To(Equal(time.Second * 2))
			Expect(cfg.Multizone.Global.KDS.ZoneInsightFlushInterval.Duration).To(Equal(time.Second * 5))
			Expect(cfg.Multizone.Global.KDS.TlsMinVersion).To(Equal("TLSv1_3"))
			Expect(cfg.Multizone.Global.KDS.TlsMaxVersion).To(Equal("TLSv1_3"))
			Expect(cfg.Multizone.Global.KDS.TlsCipherSuites).To(Equal([]string{"TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"}))
			Expect(cfg.Multizone.Global.KDS.TlsEnabled).To(BeFalse())
			Expect(cfg.Multizone.Global.KDS.TlsCertFile).To(Equal("/cert"))
			Expect(cfg.Multizone.Global.KDS.TlsKeyFile).To(Equal("/key"))
			Expect(cfg.Multizone.Global.KDS.MaxMsgSize).To(Equal(uint32(1)))
			Expect(cfg.Multizone.Global.KDS.MsgSendTimeout.Duration).To(Equal(10 * time.Second))
			Expect(cfg.Multizone.Global.KDS.NackBackoff.Duration).To(Equal(11 * time.Second))
			Expect(cfg.Multizone.Global.KDS.ResponseBackoff.Duration).To(Equal(time.Second))
			Expect(cfg.Multizone.Global.KDS.ZoneHealthCheck.PollInterval.Duration).To(Equal(11 * time.Second))
			Expect(cfg.Multizone.Global.KDS.ZoneHealthCheck.Timeout.Duration).To(Equal(110 * time.Second))
			Expect(cfg.Multizone.Global.KDS.Tracing.Enabled).To(BeFalse())
			Expect(cfg.Multizone.Zone.GlobalAddress).To(Equal("grpc://1.1.1.1:5685"))
			Expect(cfg.Multizone.Zone.Name).To(Equal("zone-1"))
			Expect(cfg.Multizone.Zone.KDS.RootCAFile).To(Equal("/rootCa"))
			Expect(cfg.Multizone.Zone.KDS.RefreshInterval.Duration).To(Equal(9 * time.Second))
			Expect(cfg.Multizone.Zone.KDS.MaxMsgSize).To(Equal(uint32(2)))
			Expect(cfg.Multizone.Zone.KDS.MsgSendTimeout.Duration).To(Equal(20 * time.Second))
			Expect(cfg.Multizone.Zone.KDS.NackBackoff.Duration).To(Equal(21 * time.Second))
			Expect(cfg.Multizone.Zone.KDS.ResponseBackoff.Duration).To(Equal(2 * time.Second))
			Expect(cfg.Multizone.Zone.KDS.TlsSkipVerify).To(BeTrue())

			Expect(cfg.Defaults.SkipMeshCreation).To(BeTrue())
			Expect(cfg.Defaults.SkipTenantResources).To(BeTrue())
			Expect(cfg.Defaults.CreateMeshRoutingResources).To(BeTrue())
			Expect(cfg.Defaults.SkipHostnameGenerators).To(BeTrue())

			Expect(cfg.Diagnostics.ServerPort).To(Equal(uint32(5003)))
			Expect(cfg.Diagnostics.DebugEndpoints).To(BeTrue())
			Expect(cfg.Diagnostics.TlsEnabled).To(BeTrue())
			Expect(cfg.Diagnostics.TlsCertFile).To(Equal("/cert"))
			Expect(cfg.Diagnostics.TlsKeyFile).To(Equal("/key"))
			Expect(cfg.Diagnostics.TlsMinVersion).To(Equal("TLSv1_3"))
			Expect(cfg.Diagnostics.TlsMaxVersion).To(Equal("TLSv1_3"))
			Expect(cfg.Diagnostics.TlsCipherSuites).To(Equal([]string{"TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"}))

			Expect(cfg.DNSServer.Domain).To(Equal("test-domain"))
			Expect(cfg.DNSServer.CIDR).To(Equal("127.1.0.0/16"))
			Expect(cfg.DNSServer.ServiceVipEnabled).To(BeFalse())
			Expect(cfg.DNSServer.ServiceVipPort).To(Equal(uint32(9090)))

			Expect(cfg.XdsServer.DataplaneStatusFlushInterval.Duration).To(Equal(7 * time.Second))
			Expect(cfg.XdsServer.DataplaneConfigurationRefreshInterval.Duration).To(Equal(21 * time.Second))
			Expect(cfg.XdsServer.DataplaneDeregistrationDelay.Duration).To(Equal(11 * time.Second))
			Expect(cfg.XdsServer.NACKBackoff.Duration).To(Equal(10 * time.Second))

			Expect(cfg.Metrics.Zone.SubscriptionLimit).To(Equal(23))
			Expect(cfg.Metrics.Zone.IdleTimeout.Duration).To(Equal(2 * time.Minute))
			Expect(cfg.Metrics.Zone.CompactFinishedSubscriptions).To(BeTrue())
			Expect(cfg.Metrics.Mesh.MinResyncInterval.Duration).To(Equal(27 * time.Second))
			Expect(cfg.Metrics.Mesh.FullResyncInterval.Duration).To(Equal(35 * time.Second))
			Expect(cfg.Metrics.Mesh.BufferSize).To(Equal(23))
			Expect(cfg.Metrics.Mesh.EventProcessors).To(Equal(2))
			Expect(cfg.Metrics.Dataplane.SubscriptionLimit).To(Equal(47))
			Expect(cfg.Metrics.Dataplane.IdleTimeout.Duration).To(Equal(1 * time.Minute))
			Expect(cfg.Metrics.ControlPlane.ReportResourcesCount).To(BeTrue())

			Expect(cfg.DpServer.TlsCertFile).To(Equal("/test/path"))
			Expect(cfg.DpServer.TlsKeyFile).To(Equal("/test/path/key"))
			Expect(cfg.DpServer.Authn.DpProxy.Type).To(Equal("dpToken"))
			Expect(cfg.DpServer.Authn.DpProxy.DpToken.EnableIssuer).To(BeFalse())
			Expect(cfg.DpServer.Authn.DpProxy.DpToken.Validator.UseSecrets).To(BeFalse())
			Expect(cfg.DpServer.Authn.ZoneProxy.Type).To(Equal("zoneToken"))
			Expect(cfg.DpServer.Authn.ZoneProxy.ZoneToken.EnableIssuer).To(BeFalse())
			Expect(cfg.DpServer.Authn.ZoneProxy.ZoneToken.Validator.UseSecrets).To(BeFalse())
			Expect(cfg.DpServer.Authn.EnableReloadableTokens).To(BeTrue())
			Expect(cfg.DpServer.Port).To(Equal(9876))
			Expect(cfg.DpServer.TlsMinVersion).To(Equal("TLSv1_3"))
			Expect(cfg.DpServer.TlsMaxVersion).To(Equal("TLSv1_3"))
			Expect(cfg.DpServer.TlsCipherSuites).To(Equal([]string{"TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"}))
			Expect(cfg.DpServer.ReadHeaderTimeout.Duration).To(Equal(11 * time.Second))
			Expect(cfg.DpServer.Hds.Enabled).To(BeFalse())
			Expect(cfg.DpServer.Hds.Interval.Duration).To(Equal(11 * time.Second))
			Expect(cfg.DpServer.Hds.RefreshInterval.Duration).To(Equal(12 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.Timeout.Duration).To(Equal(5 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.Interval.Duration).To(Equal(6 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.NoTrafficInterval.Duration).To(Equal(7 * time.Second))
			Expect(cfg.DpServer.Hds.CheckDefaults.HealthyThreshold).To(Equal(uint32(8)))
			Expect(cfg.DpServer.Hds.CheckDefaults.UnhealthyThreshold).To(Equal(uint32(9)))

			Expect(cfg.InterCp.Catalog.InstanceAddress).To(Equal("192.168.0.1"))
			Expect(cfg.InterCp.Catalog.HeartbeatInterval.Duration).To(Equal(time.Second))
			Expect(cfg.InterCp.Catalog.WriterInterval.Duration).To(Equal(2 * time.Second))
			Expect(cfg.InterCp.Server.Port).To(Equal(uint16(15683)))
			Expect(cfg.InterCp.Server.TlsMinVersion).To(Equal("TLSv1_3"))
			Expect(cfg.InterCp.Server.TlsMaxVersion).To(Equal("TLSv1_3"))
			Expect(cfg.InterCp.Server.TlsCipherSuites).To(Equal([]string{"TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"}))

			Expect(cfg.Access.Type).To(Equal("custom-rbac"))
			Expect(cfg.Access.Static.AdminResources.Users).To(Equal([]string{"ar-admin1", "ar-admin2"}))
			Expect(cfg.Access.Static.AdminResources.Groups).To(Equal([]string{"ar-group1", "ar-group2"}))
			Expect(cfg.Access.Static.GenerateDPToken.Users).To(Equal([]string{"dp-admin1", "dp-admin2"}))
			Expect(cfg.Access.Static.GenerateDPToken.Groups).To(Equal([]string{"dp-group1", "dp-group2"}))
			Expect(cfg.Access.Static.GenerateUserToken.Users).To(Equal([]string{"ut-admin1", "ut-admin2"}))
			Expect(cfg.Access.Static.GenerateUserToken.Groups).To(Equal([]string{"ut-group1", "ut-group2"}))
			Expect(cfg.Access.Static.GenerateZoneToken.Users).To(Equal([]string{"zt-admin1", "zt-admin2"}))
			Expect(cfg.Access.Static.GenerateZoneToken.Groups).To(Equal([]string{"zt-group1", "zt-group2"}))
			Expect(cfg.Access.Static.ViewConfigDump.Users).To(Equal([]string{"zt-admin1", "zt-admin2"}))
			Expect(cfg.Access.Static.ViewConfigDump.Groups).To(Equal([]string{"zt-group1", "zt-group2"}))
			Expect(cfg.Access.Static.ViewStats.Users).To(Equal([]string{"zt-admin1", "zt-admin2"}))
			Expect(cfg.Access.Static.ViewStats.Groups).To(Equal([]string{"zt-group1", "zt-group2"}))
			Expect(cfg.Access.Static.ViewClusters.Users).To(Equal([]string{"zt-admin1", "zt-admin2"}))
			Expect(cfg.Access.Static.ViewClusters.Groups).To(Equal([]string{"zt-group1", "zt-group2"}))
			Expect(cfg.Access.Static.ControlPlaneMetadata.Users).To(Equal([]string{"cp-admin1", "cp-admin2"}))
			Expect(cfg.Access.Static.ControlPlaneMetadata.Groups).To(Equal([]string{"cp-group1", "cp-group2"}))

			Expect(cfg.Experimental.KubeOutboundsAsVIPs).To(BeTrue())
			Expect(cfg.Experimental.UseTagFirstVirtualOutboundModel).To(BeFalse())
			Expect(cfg.Experimental.IngressTagFilters).To(ContainElements("kuma.io/service"))
			Expect(cfg.Experimental.KDSEventBasedWatchdog.Enabled).To(BeTrue())
			Expect(cfg.Experimental.KDSEventBasedWatchdog.FlushInterval.Duration).To(Equal(10 * time.Second))
			Expect(cfg.Experimental.KDSEventBasedWatchdog.FullResyncInterval.Duration).To(Equal(15 * time.Second))
			Expect(cfg.Experimental.KDSEventBasedWatchdog.DelayFullResync).To(BeTrue())
			Expect(cfg.Experimental.AutoReachableServices).To(BeTrue())
			Expect(cfg.Experimental.SidecarContainers).To(BeTrue())
<<<<<<< HEAD
			Expect(cfg.Experimental.SkipPersistedVIPs).To(BeTrue())
			Expect(cfg.Experimental.GenerateMeshServices).To(BeTrue())
			Expect(cfg.Experimental.DeltaXds).To(BeTrue())
=======
>>>>>>> upstream/master

			Expect(cfg.Proxy.Gateway.GlobalDownstreamMaxConnections).To(BeNumerically("==", 1))
			Expect(cfg.EventBus.BufferSize).To(Equal(uint(30)))

			Expect(cfg.IPAM.MeshService.CIDR).To(Equal("251.0.0.0/8"))
			Expect(cfg.IPAM.MeshExternalService.CIDR).To(Equal("252.0.0.0/8"))
			Expect(cfg.IPAM.MeshMultiZoneService.CIDR).To(Equal("253.0.0.0/8"))
			Expect(cfg.IPAM.AllocationInterval.Duration).To(Equal(7 * time.Second))
			Expect(cfg.MeshService.GenerationInterval.Duration).To(Equal(8 * time.Second))
			Expect(cfg.MeshService.DeletionGracePeriod.Duration).To(Equal(11 * time.Second))

			Expect(cfg.CoreResources.Enabled).To(Equal([]string{"meshservice"}))
			Expect(cfg.CoreResources.Status.MeshServiceInterval.Duration).To(Equal(6 * time.Second))
			Expect(cfg.CoreResources.Status.MeshMultiZoneServiceInterval.Duration).To(Equal(7 * time.Second))
		},
		Entry("from config file", testCase{
			envVars: map[string]string{},
			yamlFileConfig: `
environment: kubernetes
store:
  type: postgres
  unsafeDelete: true
  postgres:
    host: postgres.host
    port: 5432
    user: kuma
    password: kuma
    dbName: kuma
    driverName: postgres
    connectionTimeout: 10
    minOpenConnections: 3
    maxOpenConnections: 300
    maxConnectionLifetime: 123m
    maxConnectionLifetimeJitter: 456s
    healthCheckInterval: 78s
    maxIdleConnections: 300
    minReconnectInterval: 44s
    maxReconnectInterval: 55s
    maxListQueryElements: 111
    maxConnectionIdleTime: 99s
    tls:
      mode: verifyFull
      certPath: /path/to/cert
      keyPath: /path/to/key
      caPath: /path/to/rootCert
      disableSSLSNI: true
    readReplica:
      host: ro.host
      port: 35432
      ratio: 80
  kubernetes:
    systemNamespace: test-namespace
  cache:
    enabled: false
    expirationTime: 3s
  upsert:
    conflictRetryBaseBackoff: 4s
    conflictRetryMaxTimes: 15
    conflictRetryJitterPercent: 10
bootstrapServer:
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
    tlsCaFile: "/ca"
    tlsMinVersion: TLSv1_3
    tlsMaxVersion: TLSv1_3
    tlsCipherSuites: ["TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"]
    requireClientCert: true
  auth:
    clientCertsDir: "/certs" # ENV: KUMA_API_SERVER_AUTH_CLIENT_CERTS_DIR
  authn:
    type: custom-authn
    localhostIsAdmin: false
    tokens:
      bootstrapAdminToken: false
      enableIssuer: false
      validator:
        useSecrets: false
  readOnly: true
  corsAllowedDomains:
    - https://kuma
    - https://someapi
  rootUrl: https://foo.com
  basePath: /api
  gui:
    enabled: false
    rootUrl: https://bar.com
    basePath: /ui
monitoringAssignmentServer:
  port: 2222
  defaultFetchTimeout: 45s
  apiVersions: [v1]
  assignmentRefreshInterval: 12s
  tlsEnabled: true
  tlsCertFile: "/cert"
  tlsKeyFile: "/key"
  tlsMinVersion: TLSv1_3
  tlsMaxVersion: TLSv1_3
  tlsCipherSuites: ["TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"]
runtime:
  universal:
    dataplaneCleanupAge: 1h
    vipRefreshInterval: 10s
  kubernetes:
    serviceAccountName: custom-sa
    allowedUsers: ["allowed-usr-1", "allowed-usr-2"]
    controlPlaneServiceName: custom-control-plane
    nodeTaintController:
      enabled: true
      cniApp: kuma-cni
      cniNamespace: kuma-system
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
      applicationProbeProxyPort: 1112
      containerPatches: ["patch1", "patch2"]
      initContainer:
        image: test-image:test
      sidecarContainer:
        waitForDataplaneReady: true
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
        startupProbe:
          initialDelaySeconds: 11
          periodSeconds: 9
          failureThreshold: 32
          timeoutSeconds: 52
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
        excludeInboundIPs:
        - 192.168.0.1
        - 172.32.16.8/16
        - a81b:a033:6399:73c7:72b6:aa8c:6f22:7098
        - fe80::/10
        excludeOutboundIPs:
        - 10.0.0.1
        - 172.16.0.0/16
        - fe80::1
        - fe80::/10
      builtinDNS:
        enabled: true
        port: 1053
        logging: false
      ebpf:
        enabled: true
        instanceIPEnvVarName: FOO
        bpffsPath: /run/kuma/bar
        cgroupPath: /faz/daz/zaz
        tcAttachIface: veth1
        programsSourcePath: /kuma/baz
      ignoredServiceSelectorLabels: ["x", "y"]
      nodeLabelsToCopy: ["label-1", "label-2"]
      transparentProxyConfigMap: foo
    controllersConcurrency: 
      podController: 10
    clientConfig:
      qps: 100
      burstQps: 100
    leaderElection:
      leaseDuration: 199s
      renewDeadline: 99s
    skipMeshOwnerReference: true
    supportGatewaySecretsInAllNamespaces: true
reports:
  enabled: false
general:
  tlsKeyFile: /tmp/key
  tlsCertFile: /tmp/cert
  tlsMinVersion: TLSv1_3
  tlsMaxVersion: TLSv1_3
  tlsCipherSuites: ["TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"]
  dnsCacheTTL: 19s
  workDir: /custom/work/dir
mode: zone
multizone:
  global:
    kds:
      grpcPort: 1234
      refreshInterval: 2s
      zoneInsightFlushInterval: 5s
      tlsEnabled: false
      tlsCertFile: /cert
      tlsKeyFile: /key
      tlsMinVersion: TLSv1_3
      tlsMaxVersion: TLSv1_3
      tlsCipherSuites: ["TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"]
      maxMsgSize: 1
      msgSendTimeout: 10s
      nackBackoff: 11s
      responseBackoff: 1s
      zoneHealthCheck:
        pollInterval: 11s
        timeout: 110s
      tracing:
        enabled: false
  zone:
    globalAddress: "grpc://1.1.1.1:5685"
    name: "zone-1"
    kds:
      refreshInterval: 9s
      rootCaFile: /rootCa
      maxMsgSize: 2
      msgSendTimeout: 20s
      nackBackoff: 21s
      responseBackoff: 2s
      tlsSkipVerify: true
    disableOriginLabelValidation: true
    ingressUpdateInterval: 2s
dnsServer:
  domain: test-domain
  CIDR: 127.1.0.0/16
  serviceVipEnabled: false
  serviceVipPort: 9090
defaults:
  skipMeshCreation: true
  skipHostnameGenerators: true
  skipTenantResources: true
  createMeshRoutingResources: true
diagnostics:
  serverPort: 5003
  debugEndpoints: true
  tlsEnabled: true
  tlsCertFile: "/cert"
  tlsKeyFile: "/key"
  tlsMinVersion: TLSv1_3
  tlsMaxVersion: TLSv1_3
  tlsCipherSuites: ["TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"]
xdsServer:
  dataplaneConfigurationRefreshInterval: 21s
  dataplaneStatusFlushInterval: 7s
  dataplaneDeregistrationDelay: 11s
  nackBackoff: 10s
metrics:
  zone:
    subscriptionLimit: 23
    idleTimeout: 2m
    compactFinishedSubscriptions: true
  mesh:
    fullResyncInterval: 35s
    minResyncInterval: 27s
    bufferSize: 23
    eventProcessors: 2
  dataplane:
    subscriptionLimit: 47
    idleTimeout: 1m
  controlPlane:
    reportResourcesCount: true
dpServer:
  tlsCertFile: /test/path
  tlsKeyFile: /test/path/key
  tlsMinVersion: TLSv1_3
  tlsMaxVersion: TLSv1_3
  tlsCipherSuites: ["TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"]
  readHeaderTimeout: 11s
  port: 9876
  authn:
    dpProxy:
      type: dpToken
      dpToken:
        enableIssuer: false
        validator:
          useSecrets: false
    zoneProxy:
      type: zoneToken
      zoneToken:
        enableIssuer: false
        validator:
          useSecrets: false
    enableReloadableTokens: true
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
interCp:
  catalog:
    instanceAddress: "192.168.0.1"
    heartbeatInterval: 1s
    writerInterval: 2s
  server:
    port: 15683
    tlsMinVersion: "TLSv1_3"
    tlsMaxVersion: "TLSv1_3"
    tlsCipherSuites: ["TLS_RSA_WITH_AES_128_CBC_SHA", "TLS_AES_256_GCM_SHA384"]
access:
  type: custom-rbac
  static:
    adminResources:
      users: ["ar-admin1", "ar-admin2"]
      groups: ["ar-group1", "ar-group2"]
    generateDpToken:
      users: ["dp-admin1", "dp-admin2"]
      groups: ["dp-group1", "dp-group2"]
    generateUserToken:
      users: ["ut-admin1", "ut-admin2"]
      groups: ["ut-group1", "ut-group2"]
    generateZoneToken:
      users: ["zt-admin1", "zt-admin2"]
      groups: ["zt-group1", "zt-group2"]
    viewConfigDump:
      users: ["zt-admin1", "zt-admin2"]
      groups: ["zt-group1", "zt-group2"]
    viewStats:
      users: ["zt-admin1", "zt-admin2"]
      groups: ["zt-group1", "zt-group2"]
    viewClusters:
      users: ["zt-admin1", "zt-admin2"]
      groups: ["zt-group1", "zt-group2"]
    controlPlaneMetadata:
      users: ["cp-admin1", "cp-admin2"]
      groups: ["cp-group1", "cp-group2"]
experimental:
  kubeOutboundsAsVIPs: true
  cniApp: "kuma-cni"
  useTagFirstVirtualOutboundModel: false
  ingressTagFilters: ["kuma.io/service"]
  kdsEventBasedWatchdog:
    enabled: true
    flushInterval: 10s
    fullResyncInterval: 15s
    delayFullResync: true
  autoReachableServices: true
  sidecarContainers: true
  generateMeshServices: true
  skipPersistedVIPs: true
  deltaXds: true
proxy:
  gateway:
    globalDownstreamMaxConnections: 1
eventBus:
  bufferSize: 30
coreResources:
  enabled:
  - meshservice
  status:
    meshServiceInterval: 6s
    meshMultiZoneServiceInterval: 7s
policies:
  pluginPoliciesEnabled:
    - meshaccesslog
    - meshcircuitbreaker
tracing:
  openTelemetry:
    enabled: true
    endpoint: collector:4317
ipam:
  meshService:
    cidr: 251.0.0.0/8
  meshExternalService:
    cidr: 252.0.0.0/8
  meshMultiZoneService:
    cidr: 253.0.0.0/8
  allocationInterval: 7s
meshService:
  generationInterval: 8s
  deletionGracePeriod: 11s
`,
		}),
		Entry("from env variables", testCase{
			envVars: map[string]string{
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_PORT":             "1234",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_HOST":               "kuma-control-plane",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_PORT":               "4321",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_XDS_CONNECT_TIMEOUT":    "13s",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ACCESS_LOG_PATH":  "/access/log/test",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_ADMIN_ADDRESS":          "1.1.1.1",
				"KUMA_BOOTSTRAP_SERVER_PARAMS_COREFILE_TEMPLATE_PATH": "/etc/resolv.conf",
				"KUMA_ENVIRONMENT":                                                                         "kubernetes",
				"KUMA_STORE_TYPE":                                                                          "postgres",
				"KUMA_STORE_UNSAFE_DELETE":                                                                 "true",
				"KUMA_STORE_POSTGRES_HOST":                                                                 "postgres.host",
				"KUMA_STORE_POSTGRES_PORT":                                                                 "5432",
				"KUMA_STORE_POSTGRES_USER":                                                                 "kuma",
				"KUMA_STORE_POSTGRES_PASSWORD":                                                             "kuma",
				"KUMA_STORE_POSTGRES_DB_NAME":                                                              "kuma",
				"KUMA_STORE_POSTGRES_DRIVER_NAME":                                                          "postgres",
				"KUMA_STORE_POSTGRES_CONNECTION_TIMEOUT":                                                   "10",
				"KUMA_STORE_POSTGRES_MIN_OPEN_CONNECTIONS":                                                 "3",
				"KUMA_STORE_POSTGRES_MAX_OPEN_CONNECTIONS":                                                 "300",
				"KUMA_STORE_POSTGRES_MAX_IDLE_CONNECTIONS":                                                 "300",
				"KUMA_STORE_POSTGRES_MAX_CONNECTION_LIFETIME":                                              "123m",
				"KUMA_STORE_POSTGRES_MAX_CONNECTION_LIFETIME_JITTER":                                       "456s",
				"KUMA_STORE_POSTGRES_HEALTH_CHECK_INTERVAL":                                                "78s",
				"KUMA_STORE_POSTGRES_TLS_MODE":                                                             "verifyFull",
				"KUMA_STORE_POSTGRES_TLS_CERT_PATH":                                                        "/path/to/cert",
				"KUMA_STORE_POSTGRES_TLS_KEY_PATH":                                                         "/path/to/key",
				"KUMA_STORE_POSTGRES_TLS_CA_PATH":                                                          "/path/to/rootCert",
				"KUMA_STORE_POSTGRES_TLS_DISABLE_SSLSNI":                                                   "true",
				"KUMA_STORE_POSTGRES_MAX_LIST_QUERY_ELEMENTS":                                              "111",
				"KUMA_STORE_POSTGRES_READ_REPLICA_HOST":                                                    "ro.host",
				"KUMA_STORE_POSTGRES_READ_REPLICA_PORT":                                                    "35432",
				"KUMA_STORE_POSTGRES_READ_REPLICA_RATIO":                                                   "80",
				"KUMA_STORE_POSTGRES_MAX_CONNECTION_IDLE_TIME":                                             "99s",
				"KUMA_STORE_KUBERNETES_SYSTEM_NAMESPACE":                                                   "test-namespace",
				"KUMA_STORE_CACHE_ENABLED":                                                                 "false",
				"KUMA_STORE_CACHE_EXPIRATION_TIME":                                                         "3s",
				"KUMA_STORE_UPSERT_CONFLICT_RETRY_BASE_BACKOFF":                                            "4s",
				"KUMA_STORE_UPSERT_CONFLICT_RETRY_MAX_TIMES":                                               "15",
				"KUMA_STORE_UPSERT_CONFLICT_RETRY_JITTER_PERCENT":                                          "10",
				"KUMA_API_SERVER_READ_ONLY":                                                                "true",
				"KUMA_API_SERVER_HTTP_PORT":                                                                "15681",
				"KUMA_API_SERVER_HTTP_INTERFACE":                                                           "192.168.0.1",
				"KUMA_API_SERVER_HTTP_ENABLED":                                                             "false",
				"KUMA_API_SERVER_HTTPS_ENABLED":                                                            "false",
				"KUMA_API_SERVER_HTTPS_PORT":                                                               "15682",
				"KUMA_API_SERVER_HTTPS_INTERFACE":                                                          "192.168.0.2",
				"KUMA_API_SERVER_HTTPS_TLS_CERT_FILE":                                                      "/cert",
				"KUMA_API_SERVER_HTTPS_TLS_KEY_FILE":                                                       "/key",
				"KUMA_API_SERVER_HTTPS_TLS_CA_FILE":                                                        "/ca",
				"KUMA_API_SERVER_HTTPS_TLS_MIN_VERSION":                                                    "TLSv1_3",
				"KUMA_API_SERVER_HTTPS_TLS_MAX_VERSION":                                                    "TLSv1_3",
				"KUMA_API_SERVER_HTTPS_TLS_CIPHER_SUITES":                                                  "TLS_RSA_WITH_AES_128_CBC_SHA,TLS_AES_256_GCM_SHA384",
				"KUMA_API_SERVER_HTTPS_REQUIRE_CLIENT_CERT":                                                "true",
				"KUMA_API_SERVER_AUTH_CLIENT_CERTS_DIR":                                                    "/certs",
				"KUMA_API_SERVER_AUTHN_TYPE":                                                               "custom-authn",
				"KUMA_API_SERVER_AUTHN_LOCALHOST_IS_ADMIN":                                                 "false",
				"KUMA_API_SERVER_AUTHN_TOKENS_BOOTSTRAP_ADMIN_TOKEN":                                       "false",
				"KUMA_API_SERVER_AUTHN_TOKENS_ENABLE_ISSUER":                                               "false",
				"KUMA_API_SERVER_AUTHN_TOKENS_VALIDATOR_USE_SECRETS":                                       "false",
				"KUMA_API_SERVER_ROOT_URL":                                                                 "https://foo.com",
				"KUMA_API_SERVER_BASE_PATH":                                                                "/api",
				"KUMA_API_SERVER_GUI_ENABLED":                                                              "false",
				"KUMA_API_SERVER_GUI_ROOT_URL":                                                             "https://bar.com",
				"KUMA_API_SERVER_GUI_BASE_PATH":                                                            "/ui",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_PORT":                                                   "2222",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_DEFAULT_FETCH_TIMEOUT":                                  "45s",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_API_VERSIONS":                                           "v1",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_ASSIGNMENT_REFRESH_INTERVAL":                            "12s",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_TLS_ENABLED":                                            "true",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_TLS_CERT_FILE":                                          "/cert",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_TLS_KEY_FILE":                                           "/key",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_TLS_MIN_VERSION":                                        "TLSv1_3",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_TLS_MAX_VERSION":                                        "TLSv1_3",
				"KUMA_MONITORING_ASSIGNMENT_SERVER_TLS_CIPHER_SUITES":                                      "TLS_RSA_WITH_AES_128_CBC_SHA,TLS_AES_256_GCM_SHA384",
				"KUMA_REPORTS_ENABLED":                                                                     "false",
				"KUMA_RUNTIME_KUBERNETES_CONTROL_PLANE_SERVICE_NAME":                                       "custom-control-plane",
				"KUMA_RUNTIME_KUBERNETES_SERVICE_ACCOUNT_NAME":                                             "custom-sa",
				"KUMA_RUNTIME_KUBERNETES_ALLOWED_USERS":                                                    "allowed-usr-1,allowed-usr-2",
				"KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_ENABLED":                                    "true",
				"KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_CNI_APP":                                    "kuma-cni",
				"KUMA_RUNTIME_KUBERNETES_NODE_TAINT_CONTROLLER_CNI_NAMESPACE":                              "kuma-system",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_ADDRESS":                                         "127.0.0.2",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_PORT":                                            "9443",
				"KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_CERT_DIR":                                        "/var/run/secrets/kuma.io/kuma-admission-server/tls-cert",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_INBOUND_PORTS":                            "1234,5678",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_OUTBOUND_PORTS":                           "4321,8765",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_INBOUND_IPS":                              "192.168.0.1,172.32.16.8/16,a81b:a033:6399:73c7:72b6:aa8c:6f22:7098,fe80::/10",
				"KUMA_RUNTIME_KUBERNETES_SIDECAR_TRAFFIC_EXCLUDE_OUTBOUND_IPS":                             "10.0.0.1,172.16.0.0/16,fe80::1,fe80::/10",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_CA_CERT_FILE":                                            "/tmp/ca.crt",
				"KUMA_RUNTIME_KUBERNETES_MARSHALING_CACHE_EXPIRATION_TIME":                                 "28s",
				"KUMA_INJECTOR_INIT_CONTAINER_IMAGE":                                                       "test-image:test",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_REQUESTS_MEMORY":                                "4Gi",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_REQUESTS_CPU":                                   "123m",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_LIMITS_MEMORY":                                  "8Gi",
				"KUMA_INJECTOR_SIDECAR_CONTAINER_RESOURCES_LIMITS_CPU":                                     "100m",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_CONTAINER_PATCHES":                                       "patch1,patch2",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_REDIRECT_PORT_INBOUND":                 "2020",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_IP_FAMILY_MODE":                        "dualstack",
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
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_STARTUP_PROBE_INITIAL_DELAY_SECONDS":   "11",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_STARTUP_PROBE_PERIOD_SECONDS":          "9",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_STARTUP_PROBE_FAILURE_THRESHOLD":       "32",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_STARTUP_PROBE_TIMEOUT_SECONDS":         "52",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_SIDECAR_CONTAINER_WAIT_FOR_DATAPLANE_READY":              "true",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED":                                     "true",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_PORT":                                        "1053",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_LOGGING":                                     "false",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_ENABLED":                                            "true",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_INSTANCE_IP_ENV_VAR_NAME":                           "FOO",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_BPFFS_PATH":                                         "/run/kuma/bar",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_CGROUP_PATH":                                        "/faz/daz/zaz",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_TC_ATTACH_IFACE":                                    "veth1",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_EBPF_PROGRAMS_SOURCE_PATH":                               "/kuma/baz",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_IGNORED_SERVICE_SELECTOR_LABELS":                         "x,y",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_TRANSPARENT_PROXY_CONFIGMAP_NAME":                        "foo",
				"KUMA_RUNTIME_KUBERNETES_INJECTOR_NODE_LABELS_TO_COPY":                                     "label-1,label-2",
				"KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_ENABLED":                                           "false",
				"KUMA_RUNTIME_KUBERNETES_VIRTUAL_PROBES_PORT":                                              "1111",
				"KUMA_RUNTIME_KUBERNETES_APPLICATION_PROBE_PROXY_PORT":                                     "1112",
				"KUMA_RUNTIME_KUBERNETES_EXCEPTIONS_LABELS":                                                "openshift.io/build.name:value1,openshift.io/deployer-pod-for.name:value2",
				"KUMA_RUNTIME_KUBERNETES_CONTROLLERS_CONCURRENCY_POD_CONTROLLER":                           "10",
				"KUMA_RUNTIME_KUBERNETES_CLIENT_CONFIG_QPS":                                                "100",
				"KUMA_RUNTIME_KUBERNETES_CLIENT_CONFIG_BURST_QPS":                                          "100",
				"KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_LEASE_DURATION":                                   "199s",
				"KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_RENEW_DEADLINE":                                   "99s",
				"KUMA_RUNTIME_KUBERNETES_SKIP_MESH_OWNER_REFERENCE":                                        "true",
				"KUMA_RUNTIME_KUBERNETES_SUPPORT_GATEWAY_SECRETS_IN_ALL_NAMESPACES":                        "true",
				"KUMA_RUNTIME_UNIVERSAL_DATAPLANE_CLEANUP_AGE":                                             "1h",
				"KUMA_RUNTIME_UNIVERSAL_VIP_REFRESH_INTERVAL":                                              "10s",
				"KUMA_GENERAL_TLS_CERT_FILE":                                                               "/tmp/cert",
				"KUMA_GENERAL_TLS_KEY_FILE":                                                                "/tmp/key",
				"KUMA_GENERAL_TLS_MAX_VERSION":                                                             "TLSv1_3",
				"KUMA_GENERAL_TLS_MIN_VERSION":                                                             "TLSv1_3",
				"KUMA_GENERAL_TLS_CIPHER_SUITES":                                                           "TLS_RSA_WITH_AES_128_CBC_SHA,TLS_AES_256_GCM_SHA384",
				"KUMA_GENERAL_DNS_CACHE_TTL":                                                               "19s",
				"KUMA_GENERAL_WORK_DIR":                                                                    "/custom/work/dir",
				"KUMA_GENERAL_RESILIENT_COMPONENT_BASE_BACKOFF":                                            "1s",
				"KUMA_GENERAL_RESILIENT_COMPONENT_MAX_BACKOFF":                                             "3m",
				"KUMA_API_SERVER_CORS_ALLOWED_DOMAINS":                                                     "https://kuma,https://someapi",
				"KUMA_DNS_SERVER_DOMAIN":                                                                   "test-domain",
				"KUMA_DNS_SERVER_CIDR":                                                                     "127.1.0.0/16",
				"KUMA_DNS_SERVER_SERVICE_VIP_ENABLED":                                                      "false",
				"KUMA_DNS_SERVER_SERVICE_VIP_PORT":                                                         "9090",
				"KUMA_MODE":                                                                                "zone",
				"KUMA_MULTIZONE_GLOBAL_KDS_GRPC_PORT":                                                      "1234",
				"KUMA_MULTIZONE_GLOBAL_KDS_REFRESH_INTERVAL":                                               "2s",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_ENABLED":                                                    "false",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_CERT_FILE":                                                  "/cert",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_KEY_FILE":                                                   "/key",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_MIN_VERSION":                                                "TLSv1_3",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_MAX_VERSION":                                                "TLSv1_3",
				"KUMA_MULTIZONE_GLOBAL_KDS_TLS_CIPHER_SUITES":                                              "TLS_RSA_WITH_AES_128_CBC_SHA,TLS_AES_256_GCM_SHA384",
				"KUMA_MULTIZONE_GLOBAL_KDS_MAX_MSG_SIZE":                                                   "1",
				"KUMA_MULTIZONE_GLOBAL_KDS_MSG_SEND_TIMEOUT":                                               "10s",
				"KUMA_MULTIZONE_GLOBAL_KDS_NACK_BACKOFF":                                                   "11s",
				"KUMA_MULTIZONE_GLOBAL_KDS_RESPONSE_BACKOFF":                                               "1s",
				"KUMA_MULTIZONE_GLOBAL_KDS_ZONE_HEALTH_CHECK_POLL_INTERVAL":                                "11s",
				"KUMA_MULTIZONE_GLOBAL_KDS_ZONE_HEALTH_CHECK_TIMEOUT":                                      "110s",
				"KUMA_MULTIZONE_GLOBAL_KDS_TRACING_ENABLED":                                                "false",
				"KUMA_MULTIZONE_ZONE_GLOBAL_ADDRESS":                                                       "grpc://1.1.1.1:5685",
				"KUMA_MULTIZONE_ZONE_NAME":                                                                 "zone-1",
				"KUMA_MULTIZONE_ZONE_KDS_ROOT_CA_FILE":                                                     "/rootCa",
				"KUMA_MULTIZONE_ZONE_KDS_REFRESH_INTERVAL":                                                 "9s",
				"KUMA_MULTIZONE_ZONE_KDS_MAX_MSG_SIZE":                                                     "2",
				"KUMA_MULTIZONE_ZONE_KDS_MSG_SEND_TIMEOUT":                                                 "20s",
				"KUMA_MULTIZONE_ZONE_KDS_NACK_BACKOFF":                                                     "21s",
				"KUMA_MULTIZONE_ZONE_KDS_RESPONSE_BACKOFF":                                                 "2s",
				"KUMA_MULTIZONE_ZONE_KDS_TLS_SKIP_VERIFY":                                                  "true",
				"KUMA_MULTIZONE_ZONE_DISABLE_ORIGIN_LABEL_VALIDATION":                                      "true",
				"KUMA_MULTIZONE_ZONE_INGRESS_UPDATE_INTERVAL":                                              "2s",
				"KUMA_MULTIZONE_GLOBAL_KDS_ZONE_INSIGHT_FLUSH_INTERVAL":                                    "5s",
				"KUMA_DEFAULTS_SKIP_MESH_CREATION":                                                         "true",
				"KUMA_DEFAULTS_SKIP_HOSTNAME_GENERATORS":                                                   "true",
				"KUMA_DEFAULTS_SKIP_TENANT_RESOURCES":                                                      "true",
				"KUMA_DEFAULTS_CREATE_MESH_ROUTING_RESOURCES":                                              "true",
				"KUMA_DIAGNOSTICS_SERVER_PORT":                                                             "5003",
				"KUMA_DIAGNOSTICS_DEBUG_ENDPOINTS":                                                         "true",
				"KUMA_DIAGNOSTICS_TLS_ENABLED":                                                             "true",
				"KUMA_DIAGNOSTICS_TLS_CERT_FILE":                                                           "/cert",
				"KUMA_DIAGNOSTICS_TLS_KEY_FILE":                                                            "/key",
				"KUMA_DIAGNOSTICS_TLS_MIN_VERSION":                                                         "TLSv1_3",
				"KUMA_DIAGNOSTICS_TLS_MAX_VERSION":                                                         "TLSv1_3",
				"KUMA_DIAGNOSTICS_TLS_CIPHER_SUITES":                                                       "TLS_RSA_WITH_AES_128_CBC_SHA,TLS_AES_256_GCM_SHA384",
				"KUMA_XDS_SERVER_DATAPLANE_STATUS_FLUSH_INTERVAL":                                          "7s",
				"KUMA_XDS_SERVER_DATAPLANE_CONFIGURATION_REFRESH_INTERVAL":                                 "21s",
				"KUMA_XDS_DATAPLANE_DEREGISTRATION_DELAY":                                                  "11s",
				"KUMA_XDS_SERVER_NACK_BACKOFF":                                                             "10s",
				"KUMA_METRICS_ZONE_SUBSCRIPTION_LIMIT":                                                     "23",
				"KUMA_METRICS_ZONE_IDLE_TIMEOUT":                                                           "2m",
				"KUMA_METRICS_ZONE_COMPACT_FINISHED_SUBSCRIPTIONS":                                         "true",
				"KUMA_METRICS_MESH_MIN_RESYNC_TIMEOUT":                                                     "27s",
				"KUMA_METRICS_MESH_MAX_RESYNC_TIMEOUT":                                                     "35s",
				"KUMA_METRICS_MESH_MIN_RESYNC_INTERVAL":                                                    "27s",
				"KUMA_METRICS_MESH_FULL_RESYNC_INTERVAL":                                                   "35s",
				"KUMA_METRICS_MESH_BUFFER_SIZE":                                                            "23",
				"KUMA_METRICS_MESH_EVENT_PROCESSORS":                                                       "2",
				"KUMA_METRICS_DATAPLANE_SUBSCRIPTION_LIMIT":                                                "47",
				"KUMA_METRICS_DATAPLANE_IDLE_TIMEOUT":                                                      "1m",
				"KUMA_METRICS_CONTROL_PLANE_REPORT_RESOURCES_COUNT":                                        "true",
				"KUMA_DP_SERVER_TLS_CERT_FILE":                                                             "/test/path",
				"KUMA_DP_SERVER_TLS_KEY_FILE":                                                              "/test/path/key",
				"KUMA_DP_SERVER_TLS_MIN_VERSION":                                                           "TLSv1_3",
				"KUMA_DP_SERVER_TLS_MAX_VERSION":                                                           "TLSv1_3",
				"KUMA_DP_SERVER_TLS_CIPHER_SUITES":                                                         "TLS_RSA_WITH_AES_128_CBC_SHA,TLS_AES_256_GCM_SHA384",
				"KUMA_DP_SERVER_READ_HEADER_TIMEOUT":                                                       "11s",
				"KUMA_DP_SERVER_AUTHN_DP_PROXY_TYPE":                                                       "dpToken",
				"KUMA_DP_SERVER_AUTHN_DP_PROXY_DP_TOKEN_ENABLE_ISSUER":                                     "false",
				"KUMA_DP_SERVER_AUTHN_DP_PROXY_DP_TOKEN_VALIDATOR_USE_SECRETS":                             "false",
				"KUMA_DP_SERVER_AUTHN_ZONE_PROXY_TYPE":                                                     "zoneToken",
				"KUMA_DP_SERVER_AUTHN_ZONE_PROXY_ZONE_TOKEN_ENABLE_ISSUER":                                 "false",
				"KUMA_DP_SERVER_AUTHN_ZONE_PROXY_ZONE_TOKEN_VALIDATOR_USE_SECRETS":                         "false",
				"KUMA_DP_SERVER_AUTHN_ENABLE_RELOADABLE_TOKENS":                                            "true",
				"KUMA_DP_SERVER_PORT":                                                                      "9876",
				"KUMA_DP_SERVER_HDS_ENABLED":                                                               "false",
				"KUMA_DP_SERVER_HDS_INTERVAL":                                                              "11s",
				"KUMA_DP_SERVER_HDS_REFRESH_INTERVAL":                                                      "12s",
				"KUMA_DP_SERVER_HDS_CHECK_TIMEOUT":                                                         "5s",
				"KUMA_DP_SERVER_HDS_CHECK_INTERVAL":                                                        "6s",
				"KUMA_DP_SERVER_HDS_CHECK_NO_TRAFFIC_INTERVAL":                                             "7s",
				"KUMA_DP_SERVER_HDS_CHECK_HEALTHY_THRESHOLD":                                               "8",
				"KUMA_DP_SERVER_HDS_CHECK_UNHEALTHY_THRESHOLD":                                             "9",
				"KUMA_INTER_CP_CATALOG_INSTANCE_ADDRESS":                                                   "192.168.0.1",
				"KUMA_INTER_CP_CATALOG_HEARTBEAT_INTERVAL":                                                 "1s",
				"KUMA_INTER_CP_CATALOG_WRITER_INTERVAL":                                                    "2s",
				"KUMA_INTER_CP_SERVER_PORT":                                                                "15683",
				"KUMA_INTER_CP_SERVER_TLS_MIN_VERSION":                                                     "TLSv1_3",
				"KUMA_INTER_CP_SERVER_TLS_MAX_VERSION":                                                     "TLSv1_3",
				"KUMA_INTER_CP_SERVER_TLS_CIPHER_SUITES":                                                   "TLS_RSA_WITH_AES_128_CBC_SHA,TLS_AES_256_GCM_SHA384",
				"KUMA_ACCESS_TYPE":                                                                         "custom-rbac",
				"KUMA_ACCESS_STATIC_ADMIN_RESOURCES_USERS":                                                 "ar-admin1,ar-admin2",
				"KUMA_ACCESS_STATIC_ADMIN_RESOURCES_GROUPS":                                                "ar-group1,ar-group2",
				"KUMA_ACCESS_STATIC_GENERATE_DP_TOKEN_USERS":                                               "dp-admin1,dp-admin2",
				"KUMA_ACCESS_STATIC_GENERATE_DP_TOKEN_GROUPS":                                              "dp-group1,dp-group2",
				"KUMA_ACCESS_STATIC_GENERATE_USER_TOKEN_USERS":                                             "ut-admin1,ut-admin2",
				"KUMA_ACCESS_STATIC_GENERATE_USER_TOKEN_GROUPS":                                            "ut-group1,ut-group2",
				"KUMA_ACCESS_STATIC_GENERATE_ZONE_TOKEN_USERS":                                             "zt-admin1,zt-admin2",
				"KUMA_ACCESS_STATIC_GENERATE_ZONE_TOKEN_GROUPS":                                            "zt-group1,zt-group2",
				"KUMA_ACCESS_STATIC_GET_CONFIG_DUMP_USERS":                                                 "zt-admin1,zt-admin2",
				"KUMA_ACCESS_STATIC_GET_CONFIG_DUMP_GROUPS":                                                "zt-group1,zt-group2",
				"KUMA_ACCESS_STATIC_VIEW_STATS_USERS":                                                      "zt-admin1,zt-admin2",
				"KUMA_ACCESS_STATIC_VIEW_STATS_GROUPS":                                                     "zt-group1,zt-group2",
				"KUMA_ACCESS_STATIC_VIEW_CLUSTERS_USERS":                                                   "zt-admin1,zt-admin2",
				"KUMA_ACCESS_STATIC_VIEW_CLUSTERS_GROUPS":                                                  "zt-group1,zt-group2",
				"KUMA_ACCESS_STATIC_CONTROL_PLANE_METADATA_USERS":                                          "cp-admin1,cp-admin2",
				"KUMA_ACCESS_STATIC_CONTROL_PLANE_METADATA_GROUPS":                                         "cp-group1,cp-group2",
				"KUMA_EXPERIMENTAL_KUBE_OUTBOUNDS_AS_VIPS":                                                 "true",
				"KUMA_EXPERIMENTAL_USE_TAG_FIRST_VIRTUAL_OUTBOUND_MODEL":                                   "false",
				"KUMA_EXPERIMENTAL_INGRESS_TAG_FILTERS":                                                    "kuma.io/service",
				"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_ENABLED":                                       "true",
				"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_FLUSH_INTERVAL":                                "10s",
				"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_FULL_RESYNC_INTERVAL":                          "15s",
				"KUMA_EXPERIMENTAL_KDS_EVENT_BASED_WATCHDOG_DELAY_FULL_RESYNC":                             "true",
				"KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES":                                                "true",
				"KUMA_EXPERIMENTAL_SIDECAR_CONTAINERS":                                                     "true",
				"KUMA_EXPERIMENTAL_DELTA_XDS":                                                              "true",
				"KUMA_PROXY_GATEWAY_GLOBAL_DOWNSTREAM_MAX_CONNECTIONS":                                     "1",
				"KUMA_TRACING_OPENTELEMETRY_ENDPOINT":                                                      "otel-collector:4317",
				"KUMA_TRACING_OPENTELEMETRY_ENABLED":                                                       "true",
				"KUMA_EVENT_BUS_BUFFER_SIZE":                                                               "30",
				"KUMA_PLUGIN_POLICIES_ENABLED":                                                             "meshaccesslog,meshcircuitbreaker",
				"KUMA_CORE_RESOURCES_ENABLED":                                                              "meshservice",
				"KUMA_CORE_RESOURCES_STATUS_MESH_SERVICE_INTERVAL":                                         "6s",
				"KUMA_CORE_RESOURCES_STATUS_MESH_MULTI_ZONE_SERVICE_INTERVAL":                              "7s",
				"KUMA_IPAM_MESH_SERVICE_CIDR":                                                              "251.0.0.0/8",
				"KUMA_IPAM_MESH_EXTERNAL_SERVICE_CIDR":                                                     "252.0.0.0/8",
				"KUMA_IPAM_MESH_MULTI_ZONE_SERVICE_CIDR":                                                   "253.0.0.0/8",
				"KUMA_IPAM_ALLOCATION_INTERVAL":                                                            "7s",
				"KUMA_MESH_SERVICE_GENERATION_INTERVAL":                                                    "8s",
				"KUMA_MESH_SERVICE_DELETION_GRACE_PERIOD":                                                  "11s",
			},
			yamlFileConfig: "",
		}),
	)

	It("should override via env var", func() {
		// given file with sample cfg
		file, err := os.CreateTemp("", "*")
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
