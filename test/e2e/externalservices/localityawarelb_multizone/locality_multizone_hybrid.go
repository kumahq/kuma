package localityawarelb_multizone

import (
	"fmt"
	"net"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func meshMTLSOn(mesh string, zoneEgress string) string {
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
  zoneEgress: %s
`, mesh, zoneEgress)
}

func externalService(mesh string, ip string) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: "%s"
name: external-service-in-both-zones
tags:
  kuma.io/service: external-service-in-both-zones
  kuma.io/protocol: http
networking:
  address: "%s"
`, mesh, net.JoinHostPort(ip, "8080"))
}

func zoneExternalService(mesh string, ip string, name string, zone string) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: "%s"
name: "%s"
tags:
  kuma.io/service: "%s"
  kuma.io/protocol: http
  kuma.io/zone: "%s"
networking:
  address: "%s"
`, mesh, name, name, zone, net.JoinHostPort(ip, "8080"))
}

const defaultMesh = "default"

var global, zone1 Cluster
var zone4 *UniversalCluster

func InstallExternalService(name string) InstallFunc {
	return func(cluster Cluster) error {
		return cluster.DeployApp(
			WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", name}),
			WithName(name),
			WithoutDataplane(),
			WithVerbose())
	}
}

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
		Install(YamlUniversal(meshMTLSOn(defaultMesh, "true"))).
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
		Install(DemoClientK8s(defaultMesh, TestNamespace)).
		Setup(zone1)).To(Succeed())

	E2EDeferCleanup(func() {
		Expect(zone1.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zone1.DeleteKuma()).To(Succeed())
		Expect(zone1.DismissCluster()).To(Succeed())
	})

	// Universal Cluster 4
	zone4 = universalClusters.GetCluster(Kuma4).(*UniversalCluster)
	Expect(err).ToNot(HaveOccurred())

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
		Install(DemoClientUniversal(
			"zone4-demo-client",
			defaultMesh,
			WithTransparentProxy(true),
		)).
		Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
		Install(EgressUniversal(globalCP.GenerateZoneEgressToken)).
		Install(InstallExternalService("external-service-in-zone4")).
		Install(InstallExternalService("external-service-in-zone1")).
		Install(InstallExternalService("external-service-in-both-zones")).
		Setup(zone4),
	).To(Succeed())
	E2EDeferCleanup(zone4.DismissCluster)

	err = NewClusterSetup().
		Install(YamlUniversal(zoneExternalService(defaultMesh, zone4.GetApp("external-service-in-zone4").GetIP(), "external-service-in-zone4", "kuma-4"))).
		Install(YamlUniversal(zoneExternalService(defaultMesh, zone4.GetApp("external-service-in-zone1").GetIP(), "external-service-in-zone1", "kuma-1-zone"))).
		Install(YamlUniversal(externalService(defaultMesh, zone4.GetApp("external-service-in-both-zones").GetIP()))).
		Setup(global)
	Expect(err).ToNot(HaveOccurred())
})

func ExternalServicesOnMultizoneHybridWithLocalityAwareLb() {
	BeforeEach(func() {
		Expect(global.GetKumactlOptions().
			KumactlApplyFromString(meshMTLSOn(defaultMesh, "true")),
		).To(Succeed())

		k8sCluster := zone1.(*K8sCluster)

		Expect(k8sCluster.StartZoneEgress()).To(Succeed())
		Expect(k8sCluster.StartZoneIngress()).To(Succeed())

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

	EgressStats := func(cluster Cluster, filter string) func() (*stats.Stats, error) {
		return func() (*stats.Stats, error) {
			return cluster.GetZoneEgressEnvoyTunnel().GetStats(filter)
		}
	}

	IngressStats := func(cluster Cluster, filter string) func() (*stats.Stats, error) {
		return func() (*stats.Stats, error) {
			return cluster.GetZoneIngressEnvoyTunnel().GetStats(filter)
		}
	}

	It("should route to external-service from universal through k8s", func() {
		// given no request on path
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone1",
		)
		filterIngress := "cluster.external-service-in-zone1.upstream_rq_total"

		Eventually(EgressStats(zone4, filterEgress), "15s", "1s").Should(stats.BeEqualZero())
		Eventually(IngressStats(zone1, filterIngress), "15s", "1s").Should(stats.BeEqualZero())
		Eventually(EgressStats(zone1, filterEgress), "15s", "1s").Should(stats.BeEqualZero())

		// when
		stdout, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// then should route:
		// app -> zone egress (zone4) -> zone ingress (zone1) -> zone egress (zone1) -> external service
		Eventually(EgressStats(zone4, filterEgress), "15s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(IngressStats(zone1, filterIngress), "15s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(EgressStats(zone1, filterEgress), "15s", "1s").Should(stats.BeGreaterThanZero())
	})

	It("should route to external-service from k8s through universal", func() {
		// given no request on path
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone4",
		)

		filterIngress := "cluster.external-service-in-zone4.upstream_rq_total"

		Eventually(EgressStats(zone1, filterEgress), "15s", "1s").Should(stats.BeEqualZero())
		Eventually(IngressStats(zone4, filterIngress), "15s", "1s").Should(stats.BeEqualZero())
		Eventually(EgressStats(zone4, filterEgress), "15s", "1s").Should(stats.BeEqualZero())

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
		Eventually(EgressStats(zone1, filterEgress), "15s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(IngressStats(zone4, filterIngress), "15s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(EgressStats(zone4, filterEgress), "15s", "1s").Should(stats.BeGreaterThanZero())
	})

	It("requests should be routed directly through local sidecar when zone egress disabled", func() {
		// given mesh with locality aware load balancing disabled
		mesh := "demo"
		err := NewClusterSetup().
			Install(YamlUniversal(meshMTLSOn(mesh, "false"))).
			Install(YamlUniversal(zoneExternalService(mesh, zone4.GetApp("external-service-in-zone1").GetIP(), "demo-es-in-zone1", "kuma-1-zone"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		Expect(DemoClientUniversal(
			"zone4-demo-client-2",
			mesh,
			WithTransparentProxy(true),
		)(zone4)).To(Succeed())

		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			mesh,
			"demo-es-in-zone1",
		)
		filterIngress := "cluster.demo-es-in-zone1.upstream_rq_total"

		// and there is no stat because external service is not exposed through ingress
		Eventually(func(g Gomega) {
			s, err := zone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "15s", "1s").Should(Succeed())

		// when doing requests to external service with tag zone1
		stdout, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client-2",
			"curl", "--verbose", "--max-time", "3", "--fail", "demo-es-in-zone1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// then there is no stat because external service is not exposed through egress
		Eventually(func(g Gomega) {
			s, err := zone1.GetZoneIngressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "15s", "1s").Should(Succeed())
		Eventually(func(g Gomega) {
			s, err := zone4.GetZoneIngressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "15s", "1s").Should(Succeed())
	})

	It("should fail request when ingress is down", func() {
		// when ingress is down
		Expect(zone1.(*K8sCluster).StopZoneIngress()).To(Succeed())

		// then service is unreachable
		_, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).Should(HaveOccurred())
	})

	It("should fail request when egress is down", func() {
		// when egress is down
		Expect(zone1.(*K8sCluster).StopZoneEgress()).To(Succeed())

		// then service is unreachable
		_, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).Should(HaveOccurred())
	})
}
