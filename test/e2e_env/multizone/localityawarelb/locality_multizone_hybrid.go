package localityawarelb

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshMTLSOnAndZoneEgressAndNoPassthrough(mesh string, zoneEgress string) string {
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

func InstallExternalService(name string) InstallFunc {
	return func(cluster Cluster) error {
		return cluster.DeployApp(
			WithArgs([]string{"test-server", "echo", "--port", "8080", "--instance", name}),
			WithName(name),
			WithoutDataplane(),
			WithVerbose())
	}
}

func ExternalServicesWithLocalityAwareLb() {
	zoneExternalService := func(mesh string, ip string, name string, zone string) string {
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

	const mesh = "es-locality-lb"
	const meshNoZoneEgress = "es-locality-lb-no-egress"
	const namespace = "es-locality-lb"

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(YamlUniversal(MeshMTLSOnAndZoneEgressAndNoPassthrough(mesh, "true"))).
			Install(YamlUniversal(MeshMTLSOnAndZoneEgressAndNoPassthrough(meshNoZoneEgress, "false"))).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshNoZoneEgress)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		// Universal Zone 4
		NewClusterSetup().
			Install(Parallel(
				DemoClientUniversal(
					"uni-zone4-demo-client",
					mesh,
					WithTransparentProxy(true),
				),
				DemoClientUniversal(
					"uni-zone4-demo-client-no-egress",
					meshNoZoneEgress,
					WithTransparentProxy(true),
				),
				InstallExternalService("external-service-in-uni-zone4"),
				InstallExternalService("external-service-in-kube-zone1"),
				InstallExternalService("external-service-in-both-zones"),
			)).
			SetupInGroup(multizone.UniZone1, &group)

		// Kubernetes Zone 1
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(mesh))).
			SetupInGroup(multizone.KubeZone1, &group)

		Expect(group.Wait()).To(Succeed())

		Expect(NewClusterSetup().
			Install(YamlUniversal(zoneExternalService(mesh, multizone.UniZone1.GetApp("external-service-in-uni-zone4").GetIP(), "external-service-in-uni-zone4", "kuma-4"))).
			Install(YamlUniversal(zoneExternalService(mesh, multizone.UniZone1.GetApp("external-service-in-kube-zone1").GetIP(), "external-service-in-kube-zone1", "kuma-1"))).
			Install(YamlUniversal(externalService(mesh, multizone.UniZone1.GetApp("external-service-in-both-zones").GetIP()))).
			Install(YamlUniversal(zoneExternalService(meshNoZoneEgress, multizone.UniZone1.GetApp("external-service-in-uni-zone4").GetIP(), "demo-es-in-uni-zone4", "kuma-4"))).
			Setup(multizone.Global)).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugUniversal(multizone.Global, meshNoZoneEgress)
		DebugUniversal(multizone.UniZone1, mesh)
		DebugUniversal(multizone.UniZone2, meshNoZoneEgress)
		DebugKube(multizone.KubeZone1, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshNoZoneEgress)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshNoZoneEgress)).To(Succeed())
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
			mesh,
			"external-service-in-kube-zone1",
		)
		filterIngress := fmt.Sprintf("cluster.%s_external-service-in-kube-zone1.upstream_rq_total", mesh)

		Eventually(EgressStats(multizone.UniZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(IngressStats(multizone.KubeZone1, filterIngress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(EgressStats(multizone.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())

		// when
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(multizone.UniZone1, "uni-zone4-demo-client", "external-service-in-kube-zone1.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).Should(Equal("external-service-in-kube-zone1"))
		}, "30s", "1s").Should(Succeed())

		// then should route:
		// app -> zone egress (zone4) -> zone ingress (zone1) -> zone egress (zone1) -> external service
		Eventually(EgressStats(multizone.UniZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(IngressStats(multizone.KubeZone1, filterIngress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(EgressStats(multizone.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
	})

	It("should route to external-service from k8s through universal", func() {
		// given no request on path
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			mesh,
			"external-service-in-uni-zone4",
		)
		filterIngress := fmt.Sprintf("cluster.%s_external-service-in-uni-zone4.upstream_rq_total", mesh)

		Eventually(EgressStats(multizone.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(IngressStats(multizone.UniZone1, filterIngress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(EgressStats(multizone.UniZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())

		// when request to external service in zone 1
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "external-service-in-uni-zone4.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("external-service-in-uni-zone4"))
		}, "30s", "1s").Should(Succeed())

		// then should route:
		// app -> zone egress (zone1) -> zone ingress (zone4) -> zone egress (zone4) -> external service
		Eventually(EgressStats(multizone.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(IngressStats(multizone.UniZone1, filterIngress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(EgressStats(multizone.UniZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
	})

	It("requests should be routed directly through local sidecar when zone egress disabled", func() {
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			meshNoZoneEgress,
			"demo-es-in-uni-zone4",
		)
		filterIngress := fmt.Sprintf("cluster.%s_demo-es-in-uni-zone4.upstream_rq_total", meshNoZoneEgress)

		// and there is no stat because external service is not exposed through ingress
		Eventually(func(g Gomega) {
			s, err := multizone.KubeZone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())

		// when doing requests to external service with tag zone1
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(multizone.UniZone1, "uni-zone4-demo-client-no-egress", "demo-es-in-uni-zone4.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).Should(Equal("external-service-in-uni-zone4"))
		}, "30s", "1s").Should(Succeed())

		// then there is no stat because external service is not exposed through egress
		Eventually(func(g Gomega) {
			s, err := multizone.KubeZone1.GetZoneIngressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())
		Eventually(func(g Gomega) {
			s, err := multizone.UniZone1.GetZoneIngressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())
	})
}
