package universal

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

func externalServiceInZone3(mesh string, address string, port int) string {
	return fmt.Sprintf(`
type: ExternalService
mesh: %s
name: external-service-in-zone3
tags:
  kuma.io/service: external-service-in-zone3
  kuma.io/protocol: http
  kuma.io/zone: kuma-3
networking:
  address: %s:%d
`, mesh, address, port)
}

const defaultMesh = "default"

var global Cluster
var zone3, zone4 *UniversalCluster

var _ = E2EBeforeSuite(func() {
	universalClusters, err := NewUniversalClusters(
		[]string{Kuma3, Kuma4, Kuma5},
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

	// Universal Cluster 3
	zone3 = universalClusters.GetCluster(Kuma3).(*UniversalCluster)
	Expect(err).ToNot(HaveOccurred())
	egressTokenZone3, err := globalCP.GenerateZoneEgressToken(Kuma3)
	Expect(err).ToNot(HaveOccurred())
	ingressTokenZone3, err := globalCP.GenerateZoneIngressToken(Kuma3)
	Expect(err).ToNot(HaveOccurred())
	demoClientTokenZone3, err := globalCP.GenerateDpToken(defaultMesh, "zone3-demo-client")
	Expect(err).ToNot(HaveOccurred())

	Expect(NewClusterSetup().
		Install(Kuma(config_core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
		Install(DemoClientUniversal(
			"zone3-demo-client",
			defaultMesh,
			demoClientTokenZone3,
			WithTransparentProxy(true),
		)).
		Install(IngressUniversal(ingressTokenZone3)).
		Install(EgressUniversal(egressTokenZone3)).
		Install(
			func(cluster Cluster) error {
				return cluster.DeployApp(
					WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", "external-service-in-zone3"}),
					WithName("external-service-in-zone3"),
					WithoutDataplane(),
					WithVerbose())
			}).
		Setup(zone3),
	).To(Succeed())

	E2EDeferCleanup(zone3.DismissCluster)

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
					WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", "external-service-in-both-zones"}),
					WithName("external-service-in-both-zones"),
					WithoutDataplane(),
					WithVerbose())
			}).
		Setup(zone4),
	).To(Succeed())

	E2EDeferCleanup(zone4.DismissCluster)

	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(externalServiceInZone3(defaultMesh, zone3.GetApp("external-service-in-zone3").GetIP(), 8080)),
	).To(Succeed())
	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(externalServiceInZone4(defaultMesh, zone4.GetApp("external-service-in-zone4").GetIP(), 8080)),
	).To(Succeed())
	Expect(global.GetKumactlOptions().
		KumactlApplyFromString(externalServiceInBothZones(defaultMesh, zone4.GetApp("external-service-in-both-zones").GetIP(), 8080)),
	).To(Succeed())
})

func ExternalServicesOnMultizoneUniversalWithLocalityAwareLb() {
	JustBeforeEach(func() {
		Eventually(func(g Gomega) {
			g.Expect(zone3.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(zone3.GetZoneIngressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(zone4.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(zone4.GetZoneIngressEnvoyTunnel().ResetCounters()).To(Succeed())
		}, "15s", "1s").Should(Succeed())
	})

	It("should route to external-service in other zone", func() {
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			defaultMesh,
			"external-service-in-zone3",
		)

		filterIngress := "cluster.external-service-in-zone3.upstream_rq_total"

		// no request on path
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone3.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone3.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeEqualZero())
		}, "15s", "1s").Should(Succeed())

		stdout, _, err := zone4.ExecWithRetries("", "", "zone4-demo-client",
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone3.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		// then should route:
		// app -> zone egress (zone4) -> zone ingress (zone3) -> zone egress (zone3) -> external service
		Eventually(func(g Gomega) {
			stat, err := zone4.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone3.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			stat, err := zone3.GetZoneEgressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stat).To(stats.BeGreaterThanZero())
		}, "15s", "1s").Should(Succeed())
	})
}
