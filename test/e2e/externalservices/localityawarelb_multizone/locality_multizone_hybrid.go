package localityawarelb_multizone

import (
	"fmt"
	"net"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
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
		Install(InstallExternalService("external-service-in-zone1")).
		Setup(zone4),
	).To(Succeed())
	E2EDeferCleanup(zone4.DismissCluster)

	err = NewClusterSetup().
		Install(YamlUniversal(zoneExternalService(defaultMesh, zone4.GetApp("external-service-in-zone1").GetIP(), "external-service-in-zone1", "kuma-1-zone"))).
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
	})
	It("should fail request when ingress is down", func() {
		// when ingress is down
		Expect(zone1.(*K8sCluster).StopZoneIngress()).To(Succeed())

		// then service is unreachable
		_, _, err := zone4.ExecWithCustomRetries("", "", "zone4-demo-client", 3, time.Second*3,
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).Should(HaveOccurred())
	})

	It("should fail request when egress is down", func() {
		// when egress is down
		Expect(zone1.(*K8sCluster).StopZoneEgress()).To(Succeed())

		// then service is unreachable
		_, _, err := zone4.ExecWithCustomRetries("", "", "zone4-demo-client", 3, time.Second*3,
			"curl", "--verbose", "--max-time", "3", "--fail", "external-service-in-zone1.mesh")
		Expect(err).Should(HaveOccurred())
	})
}
