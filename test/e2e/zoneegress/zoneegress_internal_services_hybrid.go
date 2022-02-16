package zoneegress

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func InternalServicesHybrid() {
	meshMTLSOn := func(mesh string) string {
		return fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`, mesh)
	}

	testServerOpts := testserver.DefaultDeploymentOpts()

	var global, zone1, zone2, zone3, zone4 Cluster

	const defaultMesh = "default"
	const nonDefaultMesh = "non-default"

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma3, Kuma4, Kuma5},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = universalClusters.GetCluster(Kuma5)

		err = NewClusterSetup().
			Install(Kuma(config_core.Global)).
			Install(YamlUniversal(meshMTLSOn(defaultMesh))).
			Install(YamlUniversal(meshMTLSOn(nonDefaultMesh))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		zone3TestServerToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "zone3-test-server")
		Expect(err).ToNot(HaveOccurred())
		zone4TestServerToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "zone4-test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 1
		zone1 = k8sClusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithIngress(),
				WithEgress(true),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s(nonDefaultMesh)).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())
		err = zone1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 2
		zone2 = k8sClusters.GetCluster(Kuma2)
		err = NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithIngress(),
				WithEgress(true),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(testserver.Install(
				testserver.WithMesh(nonDefaultMesh),
				testserver.WithServiceAccount("sa-test"),
			)).
			Setup(zone2)
		Expect(err).ToNot(HaveOccurred())
		err = zone2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 3
		zone3 = universalClusters.GetCluster(Kuma3)
		ingressTokenKuma3, err := globalCP.GenerateZoneIngressToken(Kuma3)
		Expect(err).ToNot(HaveOccurred())
		egressTokenKuma3, err := globalCP.GenerateZoneEgressToken(Kuma3)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(TestServerUniversal("dp-echo", nonDefaultMesh, zone3TestServerToken,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceName("zone3-test-server"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(ingressTokenKuma3)).
			Install(EgressUniversal(egressTokenKuma3)).
			Setup(zone3)
		Expect(err).ToNot(HaveOccurred())
		err = zone3.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 4
		zone4 = universalClusters.GetCluster(Kuma4)
		ingressTokenKuma4, err := globalCP.GenerateZoneIngressToken(Kuma4)
		Expect(err).ToNot(HaveOccurred())
		egressTokenKuma4, err := globalCP.GenerateZoneEgressToken(Kuma4)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(TestServerUniversal("dp-echo", nonDefaultMesh, zone4TestServerToken,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceName("zone4-test-server"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(ingressTokenKuma4)).
			Install(EgressUniversal(egressTokenKuma4)).
			Setup(zone4)
		Expect(err).ToNot(HaveOccurred())
		err = zone4.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())

		Expect(zone2.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone2.DeleteKuma()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())

		Expect(zone3.DeleteKuma()).To(Succeed())
		Expect(zone3.DismissCluster()).To(Succeed())

		Expect(zone4.DeleteKuma()).To(Succeed())
		Expect(zone4.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("k8s should access internal service behind k8s zoneingress through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_svc_80.upstream_rq_total",
			nonDefaultMesh,
			testServerOpts.Name,
			TestNamespace,
		)

		stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeEqualZero())

		pods, err := k8s.ListPodsE(
			zone1.GetTesting(),
			zone1.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		_, stderr, err := zone1.Exec(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "--max-time", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		stat, err = zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeGreaterThanZero())
	})

	FIt("universal should access internal service behind universal zoneingress through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			nonDefaultMesh,
			"zone4-test-server",
		)

		stat, err := zone3.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeEqualZero())

		_, stderr, err := zone3.Exec("", "", "demo-client",
			"curl", "-v", "--max-time", "3", "--fail", "zone4-test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		stat, err = zone3.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeGreaterThanZero())
	})

	It("k8s should access internal service behind universal zoneingress through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			nonDefaultMesh,
			"zone3-test-server",
		)

		pods, err := k8s.ListPodsE(
			zone1.GetTesting(),
			zone1.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", "demo-client"),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		clientPod := pods[0]

		stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeEqualZero())

		_, stderr, err := zone1.Exec(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "-v", "--max-time", "3", "--fail", "zone3-test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		stat, err = zone4.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeGreaterThanZero())
	})
}
