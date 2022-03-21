package externalservices

import (
	"fmt"
	"net"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

const defaultMesh = "default"
const nonDefaultMesh = "non-default"

func HybridUniversalGlobal() {
	meshMTLSOn := `
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`

	externalService1 := `
type: ExternalService
mesh: %s
name: external-service-1
tags:
  kuma.io/service: external-service-1
  kuma.io/protocol: http
networking:
  address: es-test-server.default.svc.cluster.local:80
`

	externalService2 := `
type: ExternalService
mesh: %s
name: external-service-2
tags:
  kuma.io/service: external-service-2
  kuma.io/protocol: http
networking:
  address: "%s"
`

	var global, zone1 Cluster
	var zone4 *UniversalCluster

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

		Expect(NewClusterSetup().
			Install(Kuma(config_core.Global)).
			Install(YamlUniversal(fmt.Sprintf(meshMTLSOn, defaultMesh))).
			Install(YamlUniversal(fmt.Sprintf(meshMTLSOn, nonDefaultMesh))).
			Install(YamlUniversal(fmt.Sprintf(externalService1, nonDefaultMesh))).
			Setup(global)).To(Succeed())

		E2EDeferCleanup(global.DismissCluster)

		globalCP := global.GetKuma()

		// K8s Cluster 1
		zone1 = k8sClusters.GetCluster(Kuma1)
		Expect(NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithEgress(true),
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s(nonDefaultMesh)).
			Install(testserver.Install(
				testserver.WithName("es-test-server"),
				testserver.WithNamespace("default"),
				testserver.WithArgs("echo", "--instance", "es-test-server"),
			)).
			Setup(zone1)).To(Succeed())

		E2EDeferCleanup(func() {
			Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
			Expect(zone1.DeleteKuma()).To(Succeed())
			Expect(zone1.DismissCluster()).To(Succeed())
		})

		// Universal Cluster 4
		zone4 = universalClusters.GetCluster(Kuma4).(*UniversalCluster)
		Expect(err).ToNot(HaveOccurred())
		egressTokenZone4, err := globalCP.GenerateZoneEgressToken(Kuma4)
		Expect(err).ToNot(HaveOccurred())
		demoClientTokenZone4, err := globalCP.GenerateDpToken(nonDefaultMesh, "zone4-demo-client")
		Expect(err).ToNot(HaveOccurred())

		Expect(NewClusterSetup().
			Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(DemoClientUniversal(
				"zone4-demo-client",
				nonDefaultMesh,
				demoClientTokenZone4,
				WithTransparentProxy(true),
			)).
			Install(EgressUniversal(egressTokenZone4)).
			Install(
				func(cluster Cluster) error {
					return cluster.DeployApp(
						WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", "es-test-server"}),
						WithName("es-test-server"),
						WithoutDataplane(),
						WithVerbose())
				}).
			Setup(zone4),
		).To(Succeed())

		E2EDeferCleanup(zone4.DismissCluster)

		Expect(global.GetKumactlOptions().
			KumactlApplyFromString(
				fmt.Sprintf(externalService2, nonDefaultMesh, net.JoinHostPort(zone4.GetApp("es-test-server").GetIP(), "8080"))),
		).To(Succeed())
	})

	It("k8s should access external service behind through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			nonDefaultMesh,
			"external-service-1",
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

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "30s", "1s").Should(Succeed())

		_, stderr, err := zone1.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())
	})

	It("universal should access external service through zoneegress", func() {
		filter := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			nonDefaultMesh,
			"external-service-2",
		)

		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "30s", "1s").Should(Succeed())

		stdout, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-2.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filter)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())
	})
}
