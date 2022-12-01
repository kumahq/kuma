package inbound_communication

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
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
			Setup(env.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, env.Zones())).To(Succeed())

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
			Setup(env.UniZone1),
		).To(Succeed())

		// Kubernetes Zone 1
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(mesh, namespace)).
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
			Setup(env.KubeZone1),
		).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(env.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.UniZone1.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.Global.DeleteMesh(mesh)).To(Succeed())
	})

	Describe("k8s to k8s communication", func() {
		It("should succeed when application binds to wildcard", func() {
			// when
			response, err := client.CollectResponse(
				env.KubeZone1, "demo-client", "k8s-test-server-wildcard.inbound-passthrough.svc.80.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("k8s-bound-wildcard"))
		})
		It("should succeed when application binds to pod", func() {
			// when
			response, err := client.CollectResponse(
				env.KubeZone1, "demo-client", "k8s-test-server-pod.inbound-passthrough.svc.80.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("k8s-bound-pod"))
		})
		It("should fail when application binds to localhost", func() {
			// given
			podName, err := PodNameOfApp(env.KubeZone1, "demo-client", namespace)
			Expect(err).ToNot(HaveOccurred())

			// when
			_, _, err = env.KubeZone1.Exec(namespace, podName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "k8s-test-server-localhost.inbound-passthrough.svc.80.mesh")

			// then
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("universal to universal communication", func() {
		It("should succeed when application binds to wildcard", func() {
			// when
			response, err := client.CollectResponse(
				env.UniZone1, "uni-demo-client", "uni-test-server-wildcard.mesh",
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("uni-bound-wildcard"))
		})
		It("should succeed when application binds to container ip", func() {
			// when
			response, err := client.CollectResponse(
				env.UniZone1, "uni-demo-client", "uni-test-server-containerip.mesh",
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("uni-bound-containerip"))
		})
		It("should succeed when container is not using transparent proxy", func() {
			// when
			_, _, err := env.UniZone1.Exec("", "", "uni-demo-client",
				"curl", "-v", "-m", "3", "--fail", "uni-test-server-localhost.mesh")

			// then
			Expect(err).To(HaveOccurred())
		})
		It("should fail when application binds to localhost", func() {
			// when
			_, _, err := env.UniZone1.Exec("", "", "uni-demo-client",
				"curl", "-v", "-m", "3", "--fail", "uni-test-server-localhost.mesh")

			// then
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("universal to k8s communication", func() {
		It("should succeed when application binds to wildcard", func() {
			// when
			response, err := client.CollectResponse(
				env.UniZone1, "uni-demo-client", "k8s-test-server-wildcard.inbound-passthrough.svc.80.mesh",
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("k8s-bound-wildcard"))
		})
		It("should succeed when application binds to pod", func() {
			// when
			response, err := client.CollectResponse(
				env.UniZone1, "uni-demo-client", "k8s-test-server-pod.inbound-passthrough.svc.80.mesh",
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("k8s-bound-pod"))
		})
		It("should fail when application binds to localhost", func() {
			// when
			_, _, err := env.UniZone1.Exec("", "", "uni-demo-client",
				"curl", "-v", "-m", "3", "--fail", "k8s-test-server-localhost.inbound-passthrough.svc.80.mesh")

			// then
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("k8s to universal communication", func() {
		It("should succeed when application binds to wildcard", func() {
			// when
			response, err := client.CollectResponse(
				env.KubeZone1, "demo-client", "uni-test-server-wildcard.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("uni-bound-wildcard"))
		})
		It("should succeed when application binds to container ip", func() {
			// when
			response, err := client.CollectResponse(
				env.UniZone1, "uni-demo-client", "uni-test-server-containerip.mesh",
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("uni-bound-containerip"))
		})
		It("should succeed when container is not using transparent proxy", func() {
			// when
			response, err := client.CollectResponse(
				env.KubeZone1, "demo-client", "uni-test-server-wildcard-no-tp.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("uni-bound-wildcard-no-tp"))
		})
		It("should succeed when user expose localhost outside", func() {
			// when
			response, err := client.CollectResponse(
				env.UniZone1, "uni-demo-client", "uni-test-server-localhost-exposed.mesh",
			)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(response.Instance).To(Equal("uni-bound-localhost-exposed"))
		})
		It("should fail when application binds to localhost", func() {
			// given
			podName, err := PodNameOfApp(env.KubeZone1, "demo-client", namespace)
			Expect(err).ToNot(HaveOccurred())

			// when
			_, _, err = env.KubeZone1.Exec(namespace, podName, "demo-client",
				"curl", "-v", "-m", "3", "--fail", "uni-test-server-localhost.mesh")

			// then
			Expect(err).To(HaveOccurred())
		})
	})
}
