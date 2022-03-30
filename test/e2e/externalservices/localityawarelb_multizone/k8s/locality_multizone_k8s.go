package k8s

import (
	"fmt"

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

func externalServiceInZone2(mesh string, address string, port int) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service-in-zone2
tags:
  kuma.io/service: external-service-in-zone2
  kuma.io/protocol: http
  kuma.io/zone: kuma-2-zone
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

var global Cluster
var zone1, zone2 *K8sCluster

var _ = E2EBeforeSuite(func() {
	k8sClusters, err := NewK8sClusters(
		[]string{Kuma1, Kuma2},
		Silent)
	Expect(err).ToNot(HaveOccurred())

	universalClusters, err := NewUniversalClusters(
		[]string{Kuma5},
		Silent)
	Expect(err).ToNot(HaveOccurred())

	// Global
	global = universalClusters.GetCluster(Kuma5)

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Global)).
		Install(YamlUniversal(meshMTLSOn(defaultMesh, "true", "true"))).
		Install(YamlUniversal(externalServiceInBothZones(defaultMesh, "external-service-in-both-zones.default.svc.cluster.local", 80))).
		Install(YamlUniversal(externalServiceInZone2(defaultMesh, "external-service-in-zone2.default.svc.cluster.local", 80))).
		Install(YamlUniversal(externalServiceInZone1(defaultMesh, "external-service-in-zone1.default.svc.cluster.local", 80))).
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
		Install(testserver.Install(
			testserver.WithName("external-service-in-zone1"),
			testserver.WithNamespace("default"),
			testserver.WithArgs("echo", "--instance", "external-service-in-zone1"),
		)).
		Install(testserver.Install(
			testserver.WithName("external-service-in-both-zones"),
			testserver.WithNamespace("default"),
			testserver.WithArgs("echo", "--instance", "external-service-in-both-zones"),
		)).
		Setup(zone1)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
	})

	// K8s Cluster 2
	zone2 = k8sClusters.GetCluster(Kuma2).(*K8sCluster)
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
		Install(testserver.Install(
			testserver.WithName("external-service-in-zone2"),
			testserver.WithNamespace("default"),
			testserver.WithArgs("echo", "--instance", "external-service-in-zone2"),
		)).
		Setup(zone2)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(zone2.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone2.DeleteKuma()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
	})

	E2EDeferCleanup(zone2.DismissCluster)
})

func ExternalServicesOnMultizoneK8sWithLocalityAwareLb() {
	It("should route to external-service through other zone", func() {
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone2",
		)

		filterIngress := "cluster.external-service-in-zone2.upstream_rq_total"

		// no request on path
		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone2.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone2.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
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

		// when request to external service in zone 2
		_, stderr, err := zone1.ExecWithRetries(TestNamespace, clientPod.GetName(), "demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone2.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		// then should route:
		// app -> zone egress (zone1) -> zone ingress (zone2) -> zone egress (zone2) -> external service
		Eventually(func(g Gomega) {
			stat, err := zone1.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone2.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone2.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())
	})
}
