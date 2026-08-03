package zonedisable

import (
	"fmt"
	"sync"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/zoneproxy"
)

func ZoneDisable() {
	const nonDefaultMesh = "non-default"
	const identityName = "zone-disable-identity"

	const clusterName1 = "kuma-disable1"
	const clusterName2 = "kuma-disable2"
	const clusterName3 = "kuma-disable3"
	var global, zone1, zone2 Cluster

	zoneIngress := func() InstallFunc {
		return zoneproxy.Install(
			zoneproxy.WithMesh(nonDefaultMesh),
			zoneproxy.WithIngress(),
		)
	}

	BeforeEach(func() {
		// Global
		global = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(Yaml(builders.Mesh().WithName(nonDefaultMesh))).
			Install(MeshIdentityBundled(nonDefaultMesh, identityName)).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshMultiZoneService
name: test-server
mesh: %s
spec:
  selector:
    meshService:
      matchLabels:
        kuma.io/display-name: test-server
  ports:
  - port: 80
    appProtocol: http
`, nonDefaultMesh))).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: disable-la-to-test-server
mesh: %s
spec:
  to:
  - targetRef:
      kind: MeshMultiZoneService
      labels:
        kuma.io/display-name: test-server
      sectionName: '80'
    default:
      localityAwareness:
        disabled: true
`, nonDefaultMesh))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

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
				// The zone joins the multizone deployment only now, so it gets
				// the mesh after its control plane is up and before the proxies
				// that need it start.
				Install(func(cluster Cluster) error {
					return WaitForMesh(nonDefaultMesh, []Cluster{cluster})
				}).
				Install(TestServerUniversal("test-server", nonDefaultMesh, WithArgs([]string{"echo", "--instance", "universal1"}))).
				Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, WithTransparentProxy(true))).
				Install(zoneIngress()).
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
				Install(func(cluster Cluster) error {
					return WaitForMesh(nonDefaultMesh, []Cluster{cluster})
				}).
				Install(TestServerUniversal("test-server", nonDefaultMesh, WithArgs([]string{"echo", "--instance", "universal2"}))).
				Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, WithTransparentProxy(true))).
				Install(zoneIngress()).
				Setup(zone2)
			Expect(err).ToNot(HaveOccurred())
		}()

		wg.Wait()

		// The zones exist only now, so the identity-based MeshTrafficPermission
		// can finally name their trust domains.
		Expect(MeshTrafficPermissionAllowAllUniversalWorkloadIdentity(nonDefaultMesh,
			MeshIdentityTrustDomain(nonDefaultMesh, zone1),
			MeshIdentityTrustDomain(nonDefaultMesh, zone2),
		)(global)).To(Succeed())

		Expect(DistributeMeshTrusts(global, nonDefaultMesh, identityName, zone1, zone2)).To(Succeed())
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

	// meshZoneAddresses returns the addresses of the ingresses zone
	// 'kuma-disable2' knows about. Zone proxies are mesh-scoped Dataplanes now,
	// so a MeshZoneAddress, not a ZoneIngress, carries the address of a zone
	// across the mesh.
	meshZoneAddresses := func(g Gomega) string {
		out, err := zone1.GetKumactlOptions().RunKumactlAndGetOutput(
			"get", "meshzoneaddresses", "--mesh", nonDefaultMesh, "-o", "json",
		)
		g.Expect(err).ToNot(HaveOccurred())
		return out
	}

	It("should access only local service if zone is disabled", func() {
		// given zone 'kuma-disable3' enabled
		// then we should receive responses from both test-server instances
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "test-server.mzsvc.mesh.local")
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKey(Equal(`universal1`)),
				HaveKey(Equal(`universal2`)),
			),
		)

		// and zone 'kuma-disable2' knows the ingress of zone 'kuma-disable3'
		Eventually(func(g Gomega) {
			g.Expect(meshZoneAddresses(g)).To(ContainSubstring(clusterName3))
		}, "30s", "1s").Should(Succeed())

		// when disable zone 'kuma-disable3'
		Expect(YamlUniversal(`
name: kuma-disable3
type: Zone
enabled: false
`)(global)).To(Succeed())

		// then the ingress of 'kuma-disable3' is deleted from zone 'kuma-disable2'
		Eventually(func(g Gomega) {
			g.Expect(meshZoneAddresses(g)).ToNot(ContainSubstring(clusterName3))
		}, "30s", "1s").Should(Succeed())

		// and then responses only from the local service instance
		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(zone1, "demo-client", "test-server.mzsvc.mesh.local")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKey(Equal(`universal1`)),
			),
		)
	})
}
