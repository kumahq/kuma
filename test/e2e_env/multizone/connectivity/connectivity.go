package connectivity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func Connectivity() {
	namespace := "connectivity"
	meshName := "connectivity"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server"),
			)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(DemoClientUniversal("uni-demo-client", meshName, WithTransparentProxy(true))).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		address          string
		expectedInstance string
	}

	DescribeTable("client from Kubernetes",
		func(given testCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.KubeZone1, "demo-client", given.address,
					client.FromKubernetesPod(meshName, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("should access service in another Kubernetes cluster", testCase{
			address:          "http://test-server_connectivity_svc_80.mesh",
			expectedInstance: "kube-test-server",
		}),
		Entry("should access service in another Kubernetes cluster", testCase{
			address:          "http://test-server.mesh",
			expectedInstance: "uni-test-server",
		}),
	)

	DescribeTable("client from Universal",
		func(given testCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(multizone.UniZone1, "uni-demo-client", given.address)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(Equal(given.expectedInstance))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("should access service in another Kubernetes cluster", testCase{
			address:          "http://test-server_connectivity_svc_80.mesh",
			expectedInstance: "kube-test-server",
		}),
		Entry("should access service in another Kubernetes cluster", testCase{
			address:          "http://test-server.mesh",
			expectedInstance: "uni-test-server",
		}),
	)
}
