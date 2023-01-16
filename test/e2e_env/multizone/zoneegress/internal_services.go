package zoneegress

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func InternalServices() {
	const meshName = "ze-internal"
	const namespace = "ze-internal"

	mesh := `
type: Mesh
name: ze-internal
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  zoneEgress: true
`

	BeforeAll(func() {
		// Global
		err := multizone.Global.Install(YamlUniversal(mesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		// Universal Zone 1
		err = multizone.UniZone1.Install(DemoClientUniversal(
			"zone3-demo-client",
			meshName,
			WithTransparentProxy(true),
		))
		Expect(err).ToNot(HaveOccurred())

		// Universal Zone 2
		err = multizone.UniZone2.Install(TestServerUniversal("zone4-dp-echo", meshName,
			WithArgs([]string{"echo", "--instance", "echo-v1"}),
			WithServiceName("zone4-test-server"),
		))
		Expect(err).ToNot(HaveOccurred())

		// Kubernetes Zone 1
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		// Kubernetes Zone 2
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
			)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	Context("when the client is from kubernetes cluster", func() {
		var zone1ClientPodName string

		BeforeAll(func() {
			podName, err := PodNameOfApp(multizone.KubeZone1, "demo-client", namespace)
			Expect(err).ToNot(HaveOccurred())
			zone1ClientPodName = podName
		})

		It("should access internal service behind k8s zoneingress through zoneegress", func() {
			filter := fmt.Sprintf(
				"cluster.%s_%s_%s_svc_80.upstream_rq_total",
				meshName,
				"test-server",
				namespace,
			)

			Eventually(func(g Gomega) {
				g.Expect(multizone.KubeZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).To(stats.BeEqualZero())
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				_, stderr, err := multizone.KubeZone1.Exec(namespace, zone1ClientPodName, "demo-client",
					"curl", "--verbose", "--max-time", "3", "--fail", "test-server_ze-internal_svc_80.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(multizone.KubeZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})

	Context("when the client is from universal cluster", func() {
		It("should access internal service behind universal zoneingress through zoneegress", func() {
			filter := fmt.Sprintf(
				"cluster.%s_%s.upstream_rq_total",
				meshName,
				"zone4-test-server",
			)

			Eventually(func(g Gomega) {
				g.Expect(multizone.UniZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeEqualZero())
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				stdout, _, err := multizone.UniZone1.Exec("", "", "zone3-demo-client",
					"curl", "--verbose", "--max-time", "3", "--fail", "zone4-test-server.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(multizone.UniZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})
}
