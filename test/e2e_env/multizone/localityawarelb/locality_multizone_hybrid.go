package localityawarelb

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
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
	const mesh = "external-service-locality-lb"
	const meshNoZoneEgress = "external-service-locality-lb-no-egress"
	const namespace = "external-service-locality-lb"

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(YamlUniversal(MeshMTLSOnAndZoneEgressAndNoPassthrough(mesh, "true"))).
			Install(YamlUniversal(MeshMTLSOnAndZoneEgressAndNoPassthrough(meshNoZoneEgress, "false"))).
			Setup(env.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, env.Zones())).To(Succeed())

		// Universal Zone 4
		Expect(NewClusterSetup().
			Install(DemoClientUniversal(
				"uni-zone4-demo-client",
				mesh,
				WithTransparentProxy(true),
			)).
			Install(DemoClientUniversal(
				"uni-zone4-demo-client-no-egress",
				meshNoZoneEgress,
				WithTransparentProxy(true),
			)).
			Install(InstallExternalService("external-service-in-uni-zone4")).
			Install(InstallExternalService("external-service-in-kube-zone1")).
			Install(InstallExternalService("external-service-in-both-zones")).
			Setup(env.UniZone1),
		).To(Succeed())

		// Kubernetes Zone 1
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(mesh, namespace)).
			Setup(env.KubeZone1)).ToNot(HaveOccurred())

		Expect(NewClusterSetup().
			Install(YamlUniversal(zoneExternalService(mesh, env.UniZone1.GetApp("external-service-in-uni-zone4").GetIP(), "external-service-in-uni-zone4", "kuma-4"))).
			Install(YamlUniversal(zoneExternalService(mesh, env.UniZone1.GetApp("external-service-in-kube-zone1").GetIP(), "external-service-in-kube-zone1", "kuma-1-zone"))).
			Install(YamlUniversal(externalService(mesh, env.UniZone1.GetApp("external-service-in-both-zones").GetIP()))).
			Install(YamlUniversal(zoneExternalService(meshNoZoneEgress, env.UniZone1.GetApp("external-service-in-uni-zone4").GetIP(), "demo-es-in-uni-zone4", "kuma-4"))).
			Setup(env.Global)).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.UniZone1.DeleteMeshApps(meshNoZoneEgress)).To(Succeed())
		Expect(env.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Global.DeleteMesh(mesh)).To(Succeed())
		Expect(env.Global.DeleteMesh(meshNoZoneEgress)).To(Succeed())
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
		filterIngress := "cluster.external-service-in-kube-zone1.upstream_rq_total"

		Eventually(EgressStats(env.UniZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(IngressStats(env.KubeZone1, filterIngress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(EgressStats(env.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())

		// when
		response, err := client.CollectResponse(env.UniZone1, "uni-zone4-demo-client", "external-service-in-kube-zone1.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).Should(Equal("external-service-in-kube-zone1"))

		// then should route:
		// app -> zone egress (zone4) -> zone ingress (zone1) -> zone egress (zone1) -> external service
		Eventually(EgressStats(env.UniZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(IngressStats(env.KubeZone1, filterIngress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(EgressStats(env.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
	})

	It("should route to external-service from k8s through universal", func() {
		// given no request on path
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			mesh,
			"external-service-in-uni-zone4",
		)
		filterIngress := "cluster.external-service-in-uni-zone4.upstream_rq_total"

		Eventually(EgressStats(env.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(IngressStats(env.UniZone1, filterIngress), "30s", "1s").Should(stats.BeEqualZero())
		Eventually(EgressStats(env.UniZone1, filterEgress), "30s", "1s").Should(stats.BeEqualZero())

		// when request to external service in zone 1
		response, err := client.CollectResponse(
			env.KubeZone1, "demo-client", "external-service-in-uni-zone4.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("external-service-in-uni-zone4"))

		// then should route:
		// app -> zone egress (zone1) -> zone ingress (zone4) -> zone egress (zone4) -> external service
		Eventually(EgressStats(env.KubeZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(IngressStats(env.UniZone1, filterIngress), "30s", "1s").Should(stats.BeGreaterThanZero())
		Eventually(EgressStats(env.UniZone1, filterEgress), "30s", "1s").Should(stats.BeGreaterThanZero())
	})

	It("requests should be routed directly through local sidecar when zone egress disabled", func() {
		filterEgress := fmt.Sprintf(
			"cluster.%s_%s.upstream_rq_total",
			meshNoZoneEgress,
			"demo-es-in-uni-zone4",
		)
		filterIngress := "cluster.demo-es-in-uni-zone4.upstream_rq_total"

		// and there is no stat because external service is not exposed through ingress
		Eventually(func(g Gomega) {
			s, err := env.KubeZone1.GetZoneIngressEnvoyTunnel().GetStats(filterIngress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())

		// when doing requests to external service with tag zone1
		response, err := client.CollectResponse(env.UniZone1, "uni-zone4-demo-client-no-egress", "demo-es-in-uni-zone4.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).Should(Equal("external-service-in-uni-zone4"))

		// then there is no stat because external service is not exposed through egress
		Eventually(func(g Gomega) {
			s, err := env.KubeZone1.GetZoneIngressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())
		Eventually(func(g Gomega) {
			s, err := env.UniZone1.GetZoneIngressEnvoyTunnel().GetStats(filterEgress)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())
	})
}
