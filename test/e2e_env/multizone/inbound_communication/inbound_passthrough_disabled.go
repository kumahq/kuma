package inbound_communication

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func InboundPassthroughDisabled() {
	const namespace = "inbound-passthrough-disabled"
	const mesh = "inbound-passthrough-disabled"

	BeforeAll(func() {
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
			// TODO: bind to docker ip
			Install(TestServerUniversal("uni-test-server-localhost", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-localhost", "--ip", "127.0.0.1"}),
				ServiceProbe(),
				WithServiceName("uni-test-server-localhost"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard", "--ip", "0.0.0.0"}),
				ServiceProbe(),
				WithServiceName("uni-test-server-wildcard"),
			)).
			Install(TestServerUniversal("uni-test-server-wildcard-no-tp", mesh,
				WithArgs([]string{"echo", "--instance", "uni-bound-wildcard-no-tp", "--ip", "0.0.0.0"}),
				ServiceProbe(),
				WithTransparentProxy(false),
				WithServiceName("uni-test-server-wildcard-no-tp"),
			)).
			Setup(env.UniZone2),
		).To(Succeed())

		// Kubernetes Zone 1
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(mesh, namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-localhost"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-localhost", "--ip", "127.0.0.1"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-wildcard"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-wildcard", "--ip", "0.0.0.0"),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(mesh),
				testserver.WithName("k8s-test-server-pod"),
				testserver.WithEchoArgs("echo", "--instance", "k8s-bound-pod", "--ip", "$(POD_IP)"),
				testserver.WithoutProbe(),
			)).
			Setup(env.KubeZone2),
		).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(env.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.KubeZone2.DeleteMesh(mesh))
		Expect(env.UniZone2.DeleteMeshApps(mesh)).To(Succeed())
		Expect(env.Global.DeleteMeshApps(mesh)).To(Succeed())
	})

	It("should check communication k8s to k8s", func() {
		// given
		podName, err := PodNameOfApp(env.KubeZone2, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())

		// when
		response, err := client.CollectResponse(
			env.KubeZone2, "demo-client", "k8s-test-server-wildcard.inbound-passthrough-disabled.svc.80.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			env.KubeZone2, "demo-client", "k8s-test-server-localhost.inbound-passthrough-disabled.svc.80.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-localhost"))

		// when
		_, _, err = env.KubeZone2.Exec(namespace, podName, "demo-client",
			"curl", "-v", "-m", "3", "--fail", "k8s-test-server-pod.inbound-passthrough-disabled.svc.80.mesh")

		// then
		Expect(err).To(HaveOccurred())
	})

	It("should check communication universal to universal", func() {
		// when
		response, err := client.CollectResponse(
			env.UniZone2, "uni-demo-client", "uni-test-server-wildcard.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			env.UniZone2, "uni-demo-client", "uni-test-server-wildcard-no-tp.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard-no-tp"))

		// when
		response, err = client.CollectResponse(
			env.UniZone2, "uni-demo-client", "uni-test-server-localhost.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-localhost"))
	})

	It("should check communication k8s to universal", func() {
		// when
		response, err := client.CollectResponse(
			env.KubeZone2, "demo-client", "uni-test-server-wildcard.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			env.KubeZone2, "demo-client", "uni-test-server-wildcard-no-tp.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-wildcard-no-tp"))

		// when
		response, err = client.CollectResponse(
			env.KubeZone2, "demo-client", "uni-test-server-localhost.mesh",
			client.FromKubernetesPod(namespace, "demo-client"),
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("uni-bound-localhost"))
	})

	It("should check communication universal to k8s", func() {
		// when
		response, err := client.CollectResponse(
			env.UniZone2, "uni-demo-client", "k8s-test-server-wildcard.inbound-passthrough-disabled.svc.80.mesh",
		)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-wildcard"))

		// when
		response, err = client.CollectResponse(
			env.UniZone2, "uni-demo-client", "k8s-test-server-localhost.inbound-passthrough-disabled.svc.80.mesh",
		)
		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(response.Instance).To(Equal("k8s-bound-localhost"))

		// when
		_, _, err = env.UniZone2.Exec("", "", "uni-demo-client",
			"curl", "-v", "-m", "3", "--fail", "k8s-test-server-pod.inbound-passthrough-disabled.svc.80.mesh")

		// then
		Expect(err).To(HaveOccurred())
	})
}
