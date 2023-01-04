package zoneegress

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
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
		err := env.Global.Install(YamlUniversal(mesh))
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())

		// Universal Zone 1
		err = env.UniZone1.Install(DemoClientUniversal(
			"zone3-demo-client",
			meshName,
			WithTransparentProxy(true),
		))
		Expect(err).ToNot(HaveOccurred())

		// Universal Zone 2
		err = env.UniZone2.Install(TestServerUniversal("zone4-dp-echo", meshName,
			WithArgs([]string{"echo", "--instance", "echo-v1"}),
			WithServiceName("zone4-test-server"),
		))
		Expect(err).ToNot(HaveOccurred())

		// Kubernetes Zone 1
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(env.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		// Kubernetes Zone 2
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
			)).
			Setup(env.KubeZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterAll(func() {
		Expect(env.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
	})

	Context("when the client is from kubernetes cluster", func() {
		var zone1ClientPodName string

		BeforeAll(func() {
			podName, err := PodNameOfApp(env.KubeZone1, "demo-client", namespace)
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
				g.Expect(env.KubeZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).To(stats.BeEqualZero())
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				_, stderr, err := env.KubeZone1.Exec(namespace, zone1ClientPodName, "demo-client",
					"curl", "--verbose", "--max-time", "3", "--fail", "test-server_ze-internal_svc_80.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
			}).Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(env.KubeZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
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
				g.Expect(env.UniZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeEqualZero())
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				stdout, _, err := env.UniZone1.Exec("", "", "zone3-demo-client",
					"curl", "--verbose", "--max-time", "3", "--fail", "zone4-test-server.mesh")
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(env.UniZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).
					To(stats.BeGreaterThanZero())
			}, "30s", "1s").Should(Succeed())
		})
	})
}
