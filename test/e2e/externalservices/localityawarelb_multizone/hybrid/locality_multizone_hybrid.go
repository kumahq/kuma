package hybrid

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func meshMTLSOn(mesh string, localityLb string, zoneEgress string) string {
	return fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
    - name: ca-1
      type: builtin
networking:
  outbound:
    passthrough: false
routing:
  localityAwareLoadBalancing: %s
  zoneEgress: %s
`, mesh, localityLb, zoneEgress)
}

func externalServiceInBothZones(mesh string, address string, port int) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service-in-both-zones
tags:
  kuma.io/service: external-service-in-both-zones
  kuma.io/protocol: http
networking:
  address: %s:%d
`, mesh, address, port)
}

func externalServiceInZone4(mesh string, address string, port int) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service-in-zone4
tags:
  kuma.io/service: external-service-in-zone4
  kuma.io/protocol: http
  kuma.io/zone: kuma-4
networking:
  address: %s:%d
`, mesh, address, port)
}

func externalServiceInZone1(mesh string, address string, port int) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service-in-zone1
tags:
  kuma.io/service: external-service-in-zone1
  kuma.io/protocol: http
  kuma.io/zone: kuma-1-zone
networking:
  address: %s:%d
`, mesh, address, port)
}

const defaultMesh = "default"

var global, zone1 Cluster
var zone4 *UniversalCluster

var _ = E2EBeforeSuite(func() {
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
		Install(YamlUniversal(meshMTLSOn(defaultMesh, "true", "true"))).
		Setup(global)).To(Succeed())

	E2EDeferCleanup(global.DismissCluster)

	globalCP := global.GetKuma()

	// K8s Cluster 1
	zone1 = k8sClusters.GetCluster(Kuma1).(*K8sCluster)
	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone,
			WithIngress(),
			WithIngressEnvoyAdminTunnel(),
			WithEgress(),
			WithEgressEnvoyAdminTunnel(),
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(DemoClientK8s(defaultMesh)).
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
	ingressTokenZone4, err := globalCP.GenerateZoneIngressToken(Kuma4)
	Expect(err).ToNot(HaveOccurred())
	demoClientTokenZone4, err := globalCP.GenerateDpToken(defaultMesh, "zone4-demo-client")
	Expect(err).ToNot(HaveOccurred())

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
		Install(DemoClientUniversal(
			"zone4-demo-client",
			defaultMesh,
			demoClientTokenZone4,
			WithTransparentProxy(true),
		)).
		Install(IngressUniversal(ingressTokenZone4)).
		Install(EgressUniversal(egressTokenZone4)).
		Install(
			func(cluster Cluster) error {
				return cluster.DeployApp(
					WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", "external-service-in-zone4"}),
					WithName("external-service-in-zone4"),
					WithoutDataplane(),
					WithVerbose())
			}).
		Install(
			func(cluster Cluster) error {
				return cluster.DeployApp(
					WithArgs([]string{"test-server", "echo", "--port", "8081", "--instance", "external-service-in-zone1"}),
					WithName("external-service-in-zone1"),
					WithoutDataplane(),
					WithVerbose())
			}).
		Install(
			func(cluster Cluster) error {
				return cluster.DeployApp(
					WithArgs([]string{"test-server", "echo", "--port", "8082", "--instance", "external-service-in-both-zones"}),
					WithName("external-service-in-both-zones"),
					WithoutDataplane(),
					WithVerbose())
			}).
		Setup(zone4),
	).To(Succeed())
	E2EDeferCleanup(zone4.DismissCluster)

	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(externalServiceInZone4(defaultMesh, zone4.GetApp("external-service-in-zone4").GetIP(), 8080)),
	).To(Succeed())
	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(externalServiceInZone1(defaultMesh, zone4.GetApp("external-service-in-zone1").GetIP(), 8081)),
	).To(Succeed())
	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(externalServiceInBothZones(defaultMesh, zone4.GetApp("external-service-in-both-zones").GetIP(), 8082)),
	).To(Succeed())
})

func ExternalServicesOnMultizoneHybridWithLocalityAwareLb() {
	BeforeEach(func() {
		Expect(global.GetKumactlOptions().
			KumactlApplyFromString(meshMTLSOn(defaultMesh, "true", "true")),
		).To(Succeed())

		k8sCluster := zone1.(*K8sCluster)

		err := k8sCluster.StartZoneEgress()
		Expect(err).ToNot(HaveOccurred())

		err = k8sCluster.StartZoneIngress()
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			g.Expect(zone1.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(zone1.GetZoneIngressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(zone4.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(zone4.GetZoneIngressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())
	})

	It("should route to external-service from universal through k8s", func() {
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone1",
		)

		filterIngress := "cluster.external-service-in-zone1.upstream_rq_total"

		// no request on path
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		stdout, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// then should route:
		// app -> zone egress (zone4) -> zone ingress (zone1) -> zone egress (zone1) -> external service
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())
	})

	It("should route to external-service from k8s through universal", func() {
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone4",
		)

		filterIngress := "cluster.external-service-in-zone4.upstream_rq_total"

		// no request on path
		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

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

		// when request to external service in zone 4
		_, stderr, err := zone1.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone4.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// then should route:
		// app -> zone egress (zone1) -> zone ingress (zone4) -> zone egress (zone4) -> external service
		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())
	})

	It("should fail request when ingress is down", func() {
		k8sCluster := zone1.(*K8sCluster)

		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone1",
		)
		filterIngress := "cluster.external-service-in-zone1.upstream_rq_total"

		// no request on path
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		// when ingress is down
		Expect(k8sCluster.StopZoneIngress()).To(Succeed())

		// then service is unreachable
		_, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).Should(HaveOccurred())
	})

	It("should fail request when egress is down", func() {
		k8sCluster := zone1.(*K8sCluster)

		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone1",
		)
		filterIngress := "cluster.external-service-in-zone1.upstream_rq_total"

		// no request on path
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		// when ingress is down
		Expect(k8sCluster.StopZoneEgress()).To(Succeed())

		// then service is unreachable
		_, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).Should(HaveOccurred())
	})

	It("requests should be routed through local zone egress when locality aware load balancing is disabled", func() {
		// given locality aware load balancing turned off
		Expect(global.GetKumactlOptions().
			KumactlApplyFromString(meshMTLSOn(defaultMesh, "false", "true")),
		).To(Succeed())

		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone1",
		)
		filterIngress := "cluster.external-service-in-zone1.upstream_rq_total"

		// and no request on path through ingress
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").ShouldNot(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		// when doing requests to external service with tag zone1
		stdout, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// then should route:
		// app -> zone egress (zone4) -> external service
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())
	})
}
