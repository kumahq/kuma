package zoneegress

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/externalservice"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func ExternalServicesHybrid() {
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

	var global, zone1, zone4 Cluster

	const defaultMesh = "default"
	const nonDefaultMesh = "non-default"

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma4, Kuma5},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = universalClusters.GetCluster(Kuma5)

		err = NewClusterSetup().
			Install(Kuma(config_core.Global)).
			Install(YamlUniversal(meshMTLSOn(defaultMesh))).
			Install(YamlUniversal(meshMTLSOn(nonDefaultMesh))).
			Install(externalservice.Install(externalservice.HttpServer, externalservice.UniversalAppEchoServer)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		demoClientToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())

		// K8s Cluster 1
		zone1 = k8sClusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithEgress(true),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s(nonDefaultMesh)).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())
		err = zone1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Universal Cluster 4
		zone4 = universalClusters.GetCluster(Kuma4)
		Expect(err).ToNot(HaveOccurred())
		egressTokenKuma4, err := globalCP.GenerateZoneEgressToken(Kuma4)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(EgressUniversal(egressTokenKuma4)).
			Setup(zone4)
		Expect(err).ToNot(HaveOccurred())
		err = zone4.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
		time.Sleep(10 * time.Second)
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}

		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())

		Expect(zone4.DeleteKuma()).To(Succeed())
		Expect(zone4.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("k8s should access external service behind through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s_%s_svc_80.upstream_rq_total",
			nonDefaultMesh,
			"external-service",
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
			"curl", "-v", "--max-time", "3", "--fail", "external-service.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		stat, err = zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeGreaterThanZero())
	})

	It("universal should access external service through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			nonDefaultMesh,
			"external-service",
		)

		stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeEqualZero())

		_, stderr, err := zone4.Exec("", "", "demo-client",
			"curl", "-v", "--max-time", "3", "--fail", "external-service.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		stat, err = zone4.GetZoneEgressEnvoyTunnel().GetStats(filter)
		Expect(err).ToNot(HaveOccurred())
		Expect(stat).To(stats.BeGreaterThanZero())
	})
}
