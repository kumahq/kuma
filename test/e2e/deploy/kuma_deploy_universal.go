package deploy

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/client"
)

func UniversalDeployment() {
	meshMTLSOn := func(mesh, localityAware string) string {
		return fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  localityAwareLoadBalancing: %s
`, mesh, localityAware)
	}

	const nonDefaultMesh = "non-default"

	var global, zone1, zone2 Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3, Kuma4, Kuma5},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma5)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Install(YamlUniversal(meshMTLSOn(nonDefaultMesh, "false"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		// TODO: right now these tests are deliberately run WithHDS(false)
		// even if HDS is enabled without any ServiceProbes it still affects
		// first 2-3 load balancer requests, it's fine but tests should be rewritten

		// Cluster 1
		zone1 = clusters.GetCluster(Kuma3)
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

		// Cluster 2
		zone2 = clusters.GetCluster(Kuma4)
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
	})

	E2EAfterEach(func() {
		Expect(zone1.DismissCluster()).To(Succeed())
		Expect(zone2.DismissCluster()).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
	})

	It("should access only local service if zone is disabled", func() {
		// given zone 'kuma-4' enabled
		// then we should receive responses from both test-server instances
		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(zone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKey(Equal(`universal1`)),
				HaveKey(Equal(`universal2`)),
			),
		)

		// when disable zone 'kuma-4'
		Expect(YamlUniversal(`
name: kuma-4
type: Zone
enabled: false
`)(global)).To(Succeed())

		// then 'kuma-4.ingress' is deleted from zone 'kuma-3'
		Eventually(func() bool {
			output, err := zone1.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zone-ingresses")
			if err != nil {
				return false
			}
			return !strings.Contains(output, "kuma-4.ingress")
		}, "30s", "10ms").Should(BeTrue())

		// and then responses only from the local service instance
		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(zone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKey(Equal(`universal1`)),
			),
		)
	})
}
