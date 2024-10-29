package inbound_communication

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func InboundPassthroughDisabled() {
	const namespace = "inbound-passthrough-disabled"
	const mesh = "inbound-passthrough-disabled"

	BeforeAll(func() {
		localhostAddress := "127.0.0.1"
		wildcardAddress := "0.0.0.0"
		if Config.IPV6 {
			wildcardAddress = "::"
		}

		// Global
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Universal Zone 4 We pick the second set of zones to test with passthrough disabled
		group := errgroup.Group{}
		group.Go(func() error {
			err := NewClusterSetup().
				Install(Parallel(
					DemoClientUniversal(
						"uni-demo-client",
						mesh,
						WithTransparentProxy(true),
					),
					TestServerUniversal("uni-test-server-localhost", mesh,
						WithArgs([]string{"echo", "--instance", "uni-bound-localhost", "--ip", localhostAddress}),
						ServiceProbe(),
						WithServiceName("uni-test-server-localhost"),
					),
					TestServerUniversal("uni-test-server-wildcard", mesh,
						WithArgs([]string{"echo", "--instance", "uni-bound-wildcard", "--ip", wildcardAddress}),
						ServiceProbe(),
						WithServiceName("uni-test-server-wildcard"),
					),
					TestServerUniversal("uni-test-server-wildcard-no-tp", mesh,
						WithArgs([]string{"echo", "--instance", "uni-bound-wildcard-no-tp", "--ip", wildcardAddress}),
						ServiceProbe(),
						WithTransparentProxy(false),
						WithServiceName("uni-test-server-wildcard-no-tp"),
					),
					TestServerUniversal("uni-test-server-containerip", mesh,
						WithArgs([]string{"echo", "--instance", "uni-bound-containerip"}),
						ServiceProbe(),
						BoundToContainerIp(),
						WithServiceName("uni-test-server-containerip"),
					),
				)).
				Setup(multizone.UniZone2)
			return errors.Wrap(err, multizone.UniZone2.Name())
		})

		// Kubernetes Zone 1
		group.Go(func() error {
			err := NewClusterSetup().
				Install(NamespaceWithSidecarInjection(namespace)).
				Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(mesh))).
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
					testserver.WithoutProbes(),
				)).
				Setup(multizone.KubeZone2)
			return errors.Wrap(err, multizone.KubeZone2.Name())
		})
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugUniversal(multizone.UniZone2, mesh)
		DebugKube(multizone.KubeZone2, mesh, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(mesh)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	Context("k8s communication", func() {
		DescribeTable("should success when application",
			func(url string, expectedInstance string) {
				Eventually(func(g Gomega) {
					// when
					response, err := client.CollectEchoResponse(
						multizone.KubeZone2, "demo-client", url,
						client.FromKubernetesPod(namespace, "demo-client"),
					)

					// then
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Instance).To(Equal(expectedInstance))
				}).Should(Succeed())
			},
			Entry("on k8s binds to wildcard", "k8s-test-server-wildcard.inbound-passthrough-disabled.svc.80.mesh", "k8s-bound-wildcard"),
			Entry("on universal binds to wildcard", "uni-test-server-wildcard.mesh", "uni-bound-wildcard"),
			Entry("on universal is not using transparent-proxy", "uni-test-server-wildcard-no-tp.mesh", "uni-bound-wildcard-no-tp"),
			Entry("on k8s binds to pod", "k8s-test-server-pod.inbound-passthrough-disabled.svc.80.mesh", "k8s-bound-pod"),
			Entry("on universal binds to containerip", "uni-test-server-containerip.mesh", "uni-bound-containerip"),
		)
	})

	Context("universal communication", func() {
		DescribeTable("should succeed when application",
			func(url string, expectedInstance string) {
				Eventually(func(g Gomega) {
					// when
					response, err := client.CollectEchoResponse(multizone.UniZone2, "uni-demo-client", url)

					// then
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(response.Instance).To(Equal(expectedInstance))
				}).Should(Succeed())
			},
			Entry("on universal binds to wildcard", "uni-test-server-wildcard.mesh", "uni-bound-wildcard"),
			Entry("on universal is not using transparent-proxy", "uni-test-server-wildcard-no-tp.mesh", "uni-bound-wildcard-no-tp"),
			Entry("on k8s binds to wildcard", "k8s-test-server-wildcard.inbound-passthrough-disabled.svc.80.mesh", "k8s-bound-wildcard"),
			Entry("on universal binds to containerip", "uni-test-server-containerip.mesh", "uni-bound-containerip"),
			Entry("on k8s binds to pod", "k8s-test-server-pod.inbound-passthrough-disabled.svc.80.mesh", "k8s-bound-pod"),
		)
	})
}
