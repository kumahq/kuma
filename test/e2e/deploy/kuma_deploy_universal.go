package deploy

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

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

	meshMTLSOff := func(mesh string) string {
		return fmt.Sprintf(`
type: Mesh
name: %s
`, mesh)
	}

	const defaultMesh = "default"
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
			Install(YamlUniversal(meshMTLSOff(defaultMesh))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		testServerToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())

		// TODO: right now these tests are deliberately run WithHDS(false)
		// even if HDS is enabled without any ServiceProbes it still affects
		// first 2-3 load balancer requests, it's fine but tests should be rewritten

		// Cluster 1
		zone1 = clusters.GetCluster(Kuma3)
		ingressTokenKuma3, err := globalCP.GenerateZoneIngressToken(Kuma3)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithHDS(false),
			)).
			Install(TestServerUniversal("test-server", nonDefaultMesh, testServerToken, WithArgs([]string{"echo", "--instance", "universal1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(ingressTokenKuma3)).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())

		// Cluster 2
		zone2 = clusters.GetCluster(Kuma4)
		ingressTokenKuma4, err := globalCP.GenerateZoneIngressToken(Kuma4)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone,
				WithGlobalAddress(globalCP.GetKDSServerAddress()),
				WithHDS(false),
			)).
			Install(TestServerUniversal("test-server", nonDefaultMesh, testServerToken, WithArgs([]string{"echo", "--instance", "universal2"}))).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(ingressTokenKuma4)).
			Setup(zone2)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := zone1.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zone1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = zone2.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zone2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should access service locally and remotely", func() {
		retry.DoWithRetry(zone1.GetTesting(), "curl local service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err := zone1.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})

		retry.DoWithRetry(zone2.GetTesting(), "curl remote service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err := zone2.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})
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
