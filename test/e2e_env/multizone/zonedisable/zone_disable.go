package zonedisable

import (
	"strings"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func ZoneDisable() {
	const nonDefaultMesh = "non-default"

	const clusterName1 = "kuma-disable1"
	const clusterName2 = "kuma-disable2"
	const clusterName3 = "kuma-disable3"
	var global, zone1, zone2 Cluster

	BeforeEach(func() {
		// Global
		global = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(MTLSMeshUniversal(nonDefaultMesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(nonDefaultMesh)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(nonDefaultMesh, multizone.Zones())).To(Succeed())

		globalCP := global.GetKuma()

		// TODO: right now these tests are deliberately run WithHDS(false)
		// even if HDS is enabled without any ServiceProbes it still affects
		// first 2-3 load balancer requests, it's fine but tests should be rewritten

		// Cluster 1
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			zone1 = NewUniversalCluster(NewTestingT(), clusterName2, Silent)
			err = NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
					WithHDS(false),
				)).
				Install(TestServerUniversal("test-server", nonDefaultMesh, WithArgs([]string{"echo", "--instance", "universal1"}))).
				Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, WithTransparentProxy(true))).
				Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
				Setup(zone1)
			Expect(err).ToNot(HaveOccurred())
		}()

		// Cluster 2
		go func() {
			defer GinkgoRecover()
			defer wg.Done()
			zone2 = NewUniversalCluster(NewTestingT(), clusterName3, Silent)
			err = NewClusterSetup().
				Install(Kuma(core.Zone,
					WithGlobalAddress(globalCP.GetKDSServerAddress()),
					WithHDS(false),
				)).
				Install(TestServerUniversal("test-server", nonDefaultMesh, WithArgs([]string{"echo", "--instance", "universal2"}))).
				Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, WithTransparentProxy(true))).
				Install(IngressUniversal(globalCP.GenerateZoneIngressToken)).
				Setup(zone2)
			Expect(err).ToNot(HaveOccurred())
		}()

		wg.Wait()
	})

	AfterEachFailure(func() {
		DebugUniversal(global, nonDefaultMesh)
		DebugUniversal(zone1, nonDefaultMesh)
		DebugUniversal(zone2, nonDefaultMesh)
	})

	E2EAfterEach(func() {
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should access only local service if zone is disabled", func() {
		// given zone 'kuma-disable3' enabled
		// then we should receive responses from both test-server instances
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKey(Equal(`universal1`)),
				HaveKey(Equal(`universal2`)),
			),
		)

		// when disable zone 'kuma-disable3'
		Expect(YamlUniversal(`
name: kuma-disable3
type: Zone
enabled: false
`)(global)).To(Succeed())

		// then 'kuma-disable3.ingress' is deleted from zone 'kuma-disable2'
		Eventually(func() bool {
			output, err := zone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
			if err != nil {
				return false
			}
			return !strings.Contains(output, "kuma-disable3.ingress")
		}, "30s", "10ms").Should(BeTrue())

		// and then responses only from the local service instance
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKey(Equal(`universal1`)),
			),
		)
	})
}
