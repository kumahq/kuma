package meshidentity

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/spire"
	"github.com/kumahq/kuma/v2/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func Spire() {
	meshName := "meshidentity-spire"
	trustDomain := "example.org"
	testServerName := "meshidentity-spire-test-server"

	BeforeEach(func() {
		// Install Spire server
		Expect(spire.Install(
			spire.WithTrustDomain(trustDomain),
			spire.WithName("spire-server"),
		)(universal.Cluster)).To(Succeed())

		// Get Spire deployment to interact with it
		spireDeployment := spire.From(spire.AppSpireServer, universal.Cluster)
		spireIP, err := spireDeployment.GetIP()
		Expect(err).ToNot(HaveOccurred())

		Logf("Spire server IP: %s", spireIP)

		// Create Spire agent entries for demo-client and test-server containers
		// These would be the "nodes" that can attest to the Spire server
		demoClientToken, err := spireDeployment.GetAgentJoinToken(
			universal.Cluster,
			fmt.Sprintf("spiffe://%s/agent/demo-client", trustDomain),
		)
		Expect(demoClientToken).ToNot(BeEmpty())
		Expect(err).ToNot(HaveOccurred())

		testServerToken, err := spireDeployment.GetAgentJoinToken(
			universal.Cluster,
			fmt.Sprintf("spiffe://%s/agent/test-server", trustDomain),
		)
		Expect(testServerToken).ToNot(BeEmpty())
		Expect(err).ToNot(HaveOccurred())

		// Install Kuma with Spire configuration
		Expect(NewClusterSetup().
			Install(ResourceUniversal(samples.MeshDefaultBuilder().WithName(meshName).WithMeshServicesEnabled(v1alpha1.Mesh_MeshServices_Exclusive).Build())).
			Install(DemoClientUniversal("demo-client", meshName,
				WithTransparentProxy(true),
				WithWorkload("demo-client"),
				WithSpireAgent(demoClientToken, spireIP, strconv.FormatUint(uint64(spire.SpireServerPort), 10), trustDomain),
			)).
			Install(TestServerUniversal(testServerName, meshName,
				WithArgs([]string{"echo", "--instance", "spire-test-server"}),
				WithWorkload("test-server"),
				WithSpireAgent(testServerToken, spireIP, strconv.FormatUint(uint64(spire.SpireServerPort), 10), trustDomain),
			)).
			Setup(universal.Cluster)).To(Succeed())

		Expect(spireDeployment.RegisterWorkload(
			universal.Cluster,
			fmt.Sprintf("spiffe://%s/workload/demo-client", trustDomain),
			fmt.Sprintf("spiffe://%s/agent/demo-client", trustDomain),
			"unix:uid:5678",
		)).To(Succeed())
		Expect(spireDeployment.RegisterWorkload(
			universal.Cluster,
			fmt.Sprintf("spiffe://%s/workload/test-server", trustDomain),
			fmt.Sprintf("spiffe://%s/agent/test-server", trustDomain),
			"unix:uid:5678",
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	isMeshIdentityReady := func(cluster *UniversalCluster, name string) (bool, error) {
		output, err := cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshidentity", "-m", meshName, name, "-o", "json")
		if err != nil {
			return false, err
		}
		return strings.Contains(output, "Successfully initialized"), nil
	}

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use identity from Spire for secure communication", func() {
		mtp := fmt.Sprintf(`
type: MeshTrafficPermission
name: demo-client-to-test-server
mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/workload: test-server
  rules:
  - default:
      allow:
      - spiffeID:
          type: Exact
          value: spiffe://%s/workload/demo-client
`, meshName, trustDomain)
		admin := universal.Cluster.GetApp(testServerName).GetEnvoyAdminTunnel()

		// given
		// communication works
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("spire-test-server"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.*_80.ssl.handshake")
			g.Expect(err).ToNot(HaveOccurred())
			// tls is not enabled
			g.Expect(s.Stats).To(BeEmpty())
		}, "30s", "1s").Should(Succeed())

		// when
		// enable MeshIdentity and MeshTrafficPermissions
		Expect(NewClusterSetup().
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshIdentity
name: spire-identity
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: %s
    path: "/workload/{{ .Workload }}"
  provider:
    type: Spire
    spire: {}
`, meshName, trustDomain))).
			Install(YamlUniversal(mtp)).
			Setup(universal.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			ready, err := isMeshIdentityReady(universal.Cluster, "spire-identity")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(ready).To(BeTrue())
		}, "30s", "1s").MustPassRepeatedly(3).Should(Succeed())

		// then
		// communication works
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("spire-test-server"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		Eventually(func(g Gomega) {
			s, err := admin.GetStats("listener.*_80.ssl.handshake")
			Expect(err).ToNot(HaveOccurred())
			Expect(s).To(stats.BeGreaterThanZero())
		}, "30s", "1s").Should(Succeed())
	})
}
