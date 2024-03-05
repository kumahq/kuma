package inbound_communication

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func InboundPassthrough() {
	const namespace = "inbound-passthrough"
	const mesh = "inbound-passthrough"

	BeforeAll(func() {
		localhostAddress := "127.0.0.1"
		wildcardAddress := "0.0.0.0"
		if Config.IPV6 {
			localhostAddress = "::1"
			wildcardAddress = "::"
		}
		// Global
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Universal Zone 4
		Expect(NewClusterSetup().
			Install(DemoClientUniversal(
				"uni-demo-client",
				mesh,
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("uni-test-server-localhost", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-localhost", "--ip", localhostAddress}),
				ServiceProbe(),
				WithServiceName("uni-test-server-localhost"),
			)).
			Install(TestServerUniversal("uni-test-server-localhost-exposed", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-localhost-exposed", "--ip", localhostAddress}),
				ServiceProbe(),
				WithServiceAddress(localhostAddress),
				WithServiceName("uni-test-server-localhost-exposed"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard", "--ip", wildcardAddress}),
				ServiceProbe(),
				WithServiceName("uni-test-server-wildcard"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard-no-tp", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard-no-tp", "--ip", wildcardAddress}),
				ServiceProbe(),
				WithTransparentProxy(false),
				WithServiceName("uni-test-server-wildcard-no-tp"),
			)).
			Install(TestServerUniversal("uni-test-server-containerip", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-containerip"}),
				ServiceProbe(),
				BoundToContainerIp(),
				WithServiceName("uni-test-server-containerip"),
			)).
			Setup(multizone.UniZone1),
		).To(Succeed())

		// Kubernetes Zone 1
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(mesh))).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-localhost"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-localhost", "--ip", localhostAddress),
				testserver.WithoutProbes(),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-wildcard"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-wildcard", "--ip", wildcardAddress),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-pod"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-pod", "--ip", "$(POD_IP)"),
			)).
			Setup(multizone.KubeZone1),
		).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	Context("k8s communication", func() {
		DescribeTable("should succeed when application",
			func(url string, expectedInstance string) {
				// when
				Eventually(func(g Gomega) {
					response, err := client.CollectEchoResponse(
						multizone.KubeZone1, "demo-client", url,
						client.FromKubernetesPod(namespace, "demo-client"),
					)

					// then
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Instance).To(Equal(expectedInstance))
				}).Should(Succeed())
			},
			Entry("on k8s binds to wildcard", "k8s-test-server-wildcard.inbound-passthrough.svc.80.mesh", "k8s-bound-wildcard"),
			Entry("on k8s binds to podip", "k8s-test-server-pod.inbound-passthrough.svc.80.mesh", "k8s-bound-pod"),
			Entry("on universal binds to wildcard", "uni-test-server-wildcard.mesh", "uni-bound-wildcard"),
			Entry("on universal binds to podip", "uni-test-server-containerip.mesh", "uni-bound-containerip"),
			Entry("on universal is not using transparent-proxy", "uni-test-server-wildcard-no-tp.mesh", "uni-bound-wildcard-no-tp"),
		)
		DescribeTable("should fail when application",
			func(url string) {
				Consistently(func(g Gomega) {
					_, err := client.CollectEchoResponse(
						multizone.KubeZone1, "demo-client", url,
						client.FromKubernetesPod(namespace, "demo-client"),
					)

					// then
					g.Expect(err).To(HaveOccurred())
				}).Should(Succeed())
			},
			Entry("on k8s binds to localhost", "k8s-test-server-localhost.inbound-passthrough.svc.80.mesh"),
			Entry("on universal binds to localhost", "uni-test-server-localhost.mesh"),
		)
	})

	Context("universal communication", func() {
		DescribeTable("should succeed when application",
			func(url string, expectedInstance string) {
				Eventually(func(g Gomega) {
					// when
					response, err := client.CollectEchoResponse(multizone.UniZone1, "uni-demo-client", url)

					// then
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Instance).To(Equal(expectedInstance))
				}).Should(Succeed())
			},
			Entry("on universal binds to wildcard", "uni-test-server-wildcard.mesh", "uni-bound-wildcard"),
			Entry("on universal binds to container ip", "uni-test-server-containerip.mesh", "uni-bound-containerip"),
			Entry("on universal is not using transparent-proxy", "uni-test-server-wildcard-no-tp.mesh", "uni-bound-wildcard-no-tp"),
			Entry("on k8s binds to wildcard", "k8s-test-server-wildcard.inbound-passthrough.svc.80.mesh", "k8s-bound-wildcard"),
			Entry("on k8s binds to podip", "k8s-test-server-pod.inbound-passthrough.svc.80.mesh", "k8s-bound-pod"),
		)
		DescribeTable("should fail when application",
			func(url string) {
				Consistently(func(g Gomega) {
					// when
					_, err := client.CollectEchoResponse(multizone.UniZone1, "uni-demo-client", url)
					// then
					Expect(err).To(HaveOccurred())
				}).Should(Succeed())
			},
			Entry("on universal binds to localhost", "uni-test-server-localhost.mesh"),
			Entry("on k8s binds to localhost", "k8s-test-server-localhost.inbound-passthrough.svc.80.mesh"),
		)
	})
}
