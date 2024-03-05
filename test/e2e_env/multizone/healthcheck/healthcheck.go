package healthcheck

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func ApplicationOnUniversalClientOnK8s() {
	namespace := "healthcheck-app-on-universal"
	meshName := "healthcheck-app-on-universal"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(TrafficRouteUniversal(meshName)).
			Install(TrafficPermissionUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(TestServerUniversal("test-server-1", meshName,
				WithArgs([]string{"echo", "--instance", "dp-universal-1"}),
				WithProtocol("tcp"))).
			// This instance doesn't actually start the app
			Install(TestServerUniversal("test-server-2", meshName,
				WithArgs([]string{"echo", "--instance", "dp-universal-2"}),
				WithProtocol("tcp"),
				ProxyOnly(),
				ServiceProbe())).
			Install(TestServerUniversal("test-server-3", meshName,
				WithArgs([]string{"echo", "--instance", "dp-universal-3"}),
				WithProtocol("tcp"))).
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should not load balance requests to unhealthy instance", func() {
		expectHealthyInstances := func(g Gomega) {
			instances, err := client.CollectResponsesByInstance(multizone.KubeZone1, "demo-client", "test-server.mesh",
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithNumberOfRequests(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(instances).To(HaveLen(2))
			g.Expect(instances).To(And(HaveKey("dp-universal-1"), HaveKey("dp-universal-3")))
		}

		Eventually(expectHealthyInstances).WithTimeout(30 * time.Second).WithPolling(time.Second / 2).Should(Succeed())
		Consistently(expectHealthyInstances).WithTimeout(10 * time.Second).WithPolling(time.Second / 2).Should(Succeed())
	})
}
