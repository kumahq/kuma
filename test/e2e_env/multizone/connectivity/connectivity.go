package connectivity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func Connectivity() {
	namespace := "connectivity"
	meshName := "connectivity"

	BeforeAll(func() {
		Expect(env.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(meshName, namespace)).
			Setup(env.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server"),
			)).
			Setup(env.KubeZone2)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(DemoClientUniversal("uni-demo-client", meshName, WithTransparentProxy(true))).
			Setup(env.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "uni-test-server"}))).
			Setup(env.UniZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		address          string
		expectedInstance string
	}

	DescribeTable("client from Kubernetes",
		func(given testCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectResponse(env.KubeZone1, "demo-client", given.address,
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
				response, err := client.CollectResponse(env.UniZone1, "uni-demo-client", given.address)
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
