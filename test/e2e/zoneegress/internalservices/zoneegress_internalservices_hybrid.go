package internalservices

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func meshMTLSOn(mesh string) string {
	return fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: true
`, mesh)
}

var testServerOpts = testserver.DefaultDeploymentOpts()

var global, zone1, zone2, zone3, zone4 Cluster

const defaultMesh = "default"
const nonDefaultMesh = "non-default"

var _ = E2EBeforeSuite(func() {
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

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Global)).
		Install(YamlUniversal(meshMTLSOn(defaultMesh))).
		Install(YamlUniversal(meshMTLSOn(nonDefaultMesh))).
		Setup(global)).To(Succeed())

	E2EDeferCleanup(global.DismissCluster)

	globalCP := global.GetKuma()

	// K8s Cluster 1
	zone1 = k8sClusters.GetCluster(Kuma1)
	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone,
			WithEgress(),
			WithEgressEnvoyAdminTunnel(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(DemoClientK8s(nonDefaultMesh, TestNamespace)).
		Setup(zone1)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
	})

	// K8s Cluster 2
	zone2 = k8sClusters.GetCluster(Kuma2)
	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone,
			WithIngress(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(testserver.Install(
			testserver.WithMesh(nonDefaultMesh),
			testserver.WithServiceAccount("sa-test"),
		)).
		Setup(zone2)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(zone2.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone2.DeleteKuma()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
	})

	// Universal Cluster 3
	zone3 = universalClusters.GetCluster(Kuma3)
	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()), WithVerbose())).
		Install(TestServerUniversal("zone3-dp-echo", nonDefaultMesh,
			WithArgs([]string{"echo", "--instance", "echo-v1"}),
			WithServiceName("zone3-test-server"),
		)).
		Install(DemoClientUniversal(
			"zone3-demo-client",
			nonDefaultMesh,
			WithTransparentProxy(true),
		)).
		Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
		Install(EgressUniversal(globalCP.GenerateZoneEgressToken)).
		Setup(zone3)).To(Succeed())

	E2EDeferCleanup(zone3.DismissCluster)

	// Universal Cluster 4
	zone4 = universalClusters.GetCluster(Kuma4)
	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()), WithVerbose())).
		Install(TestServerUniversal("zone4-dp-echo", nonDefaultMesh,
			WithArgs([]string{"echo", "--instance", "echo-v1"}),
			WithServiceName("zone4-test-server"),
		)).
		Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
		Setup(zone4)).To(Succeed())

	E2EDeferCleanup(zone4.DismissCluster)
})

func HybridUniversalGlobal() {
	Context("when the client is from kubernetes cluster", func() {
		var zone1ClientPodName string

		BeforeEach(func() {
			podName, err := PodNameOfApp(zone1, "demo-client", TestNamespace)
			Expect(err).ToNot(HaveOccurred())
			zone1ClientPodName = podName
		})

		JustBeforeEach(func() {
			Eventually(func(g Gomega) {
				g.Expect(zone1.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
			}, "30s", "1s").Should(Succeed())
		})

		It("should access internal service behind k8s zoneingress through zoneegress", func() {
			filter := fmt.Sprintf(
				"cluster.%s_%s_%s_svc_80.upstream_rq_total",
				nonDefaultMesh,
				testServerOpts.Name,
				TestNamespace,
			)

			Eventually(func(g Gomega) {
				g.Expect(zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).To(stats.BeEqualZero())
			}, "30s", "1s").Should(Succeed())

			_, stderr, err := zone1.ExecWithRetries(TestNamespace, zone1ClientPodName, "demo-client",
				"curl", "--verbose", "--max-time", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

			Eventually(func(g Gomega) {
				g.Expect(zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})

		It("should access internal service behind universal zoneingress through zoneegress", func() {
			filter := fmt.Sprintf(
				"cluster.%s_%s.upstream_rq_total",
				nonDefaultMesh,
				"zone3-test-server",
			)

			Eventually(func(g Gomega) {
				g.Expect(zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeEqualZero())
			}, "30s", "1s").Should(Succeed())

			_, stderr, err := zone1.ExecWithRetries(TestNamespace, zone1ClientPodName, "demo-client",
				"curl", "--verbose", "--max-time", "3", "--fail", "zone3-test-server.mesh")
			Expect(err).ToNot(HaveOccurred())
			Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

			Eventually(func(g Gomega) {
				g.Expect(zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("when the client is from universal cluster", func() {
		JustBeforeEach(func() {
			Eventually(func(g Gomega) {
				g.Expect(zone3.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
			}, "30s", "1s").Should(Succeed())
		})

		It("should access internal service behind universal zoneingress through zoneegress", func() {
			filter := fmt.Sprintf(
				"cluster.%s_%s.upstream_rq_total",
				nonDefaultMesh,
				"zone4-test-server",
			)

			Eventually(func(g Gomega) {
				g.Expect(zone3.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeEqualZero())
			}, "30s", "1s").Should(Succeed())

			stdout, _, err := zone3.ExecWithRetries("", "", "zone3-demo-client",
				"curl", "--verbose", "--max-time", "3", "--fail", "zone4-test-server.mesh")
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

			Eventually(func(g Gomega) {
				g.Expect(zone3.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})
}
