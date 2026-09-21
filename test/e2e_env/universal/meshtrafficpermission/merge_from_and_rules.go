package meshtrafficpermission

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

func MergeFromAndRulesUniversal() {
	mtlsMesh := "mtp-merge"
	identityMesh := "mtp-merge-identity"
	trustDomain := fmt.Sprintf("%s.mesh.local", identityMesh)

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(mtlsMesh)).
			Install(TestServerUniversal("merge-test-server", mtlsMesh,
				WithArgs([]string{"echo", "--instance", "merge-test-server"}),
				WithServiceName("merge-test-server"),
			)).
			Install(DemoClientUniversal("merge-demo-client", mtlsMesh, WithTransparentProxy(true))).
			Install(ResourceUniversal(samples.MeshDefaultBuilder().WithName(identityMesh).WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).Build())).
			Install(TestServerUniversal("merge-identity-test-server", identityMesh,
				WithArgs([]string{"echo", "--instance", "merge-identity-test-server"}),
				WithServiceName("merge-identity-test-server"),
				WithWorkload("merge-identity-test-server"),
			)).
			Install(DemoClientUniversal("merge-identity-demo-client", identityMesh,
				WithTransparentProxy(true),
				WithWorkload("merge-identity-demo-client"),
			)).
			Setup(universal.Cluster)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(mtlsMesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mtlsMesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMeshApps(identityMesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(identityMesh)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mtlsMesh)
		DebugUniversal(universal.Cluster, identityMesh)
	})

	trafficAllowed := func(clientName, addr string) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(universal.Cluster, clientName, addr)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
	}

	trafficBlocked := func(clientName, addr string) {
		GinkgoHelper()

		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(universal.Cluster, clientName, addr)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(403))
		}, "30s", "1s").Should(Succeed())
	}

	It("should keep 'from' allows working next to a 'rules' MeshTrafficPermission", func() {
		// given traffic allowed by a 'from' policy
		Expect(universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshTrafficPermission
name: merge-from-allow
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: merge-test-server
  from:
  - targetRef:
      kind: MeshService
      name: merge-demo-client
    default:
      action: Allow
`, mtlsMesh)))).To(Succeed())
		trafficAllowed("merge-demo-client", "merge-test-server.mesh")

		// when a 'rules' policy denies the client
		rulesPolicy := func(spiffeIDPrefix string) string {
			return fmt.Sprintf(`
type: MeshTrafficPermission
name: merge-rules-deny
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: merge-test-server
  rules:
  - default:
      deny:
      - spiffeID:
          type: Prefix
          value: %s
`, mtlsMesh, spiffeIDPrefix)
		}
		Expect(universal.Cluster.Install(YamlUniversal(rulesPolicy(fmt.Sprintf("spiffe://%s/merge-demo-client", mtlsMesh))))).To(Succeed())

		// then the 'rules' deny wins over the 'from' allow
		trafficBlocked("merge-demo-client", "merge-test-server.mesh")

		// when the 'rules' policy denies an unrelated client
		Expect(universal.Cluster.Install(YamlUniversal(rulesPolicy(fmt.Sprintf("spiffe://%s/unrelated", mtlsMesh))))).To(Succeed())

		// then the 'from' allow still applies
		trafficAllowed("merge-demo-client", "merge-test-server.mesh")
	})

	It("should keep 'kind: Mesh' 'from' allows working after MeshIdentity is enabled", func() {
		addr := "merge-identity-test-server.svc.mesh.local"

		// given a mesh-wide 'from' allow and no 'rules' allow-all
		Expect(universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshTrafficPermission
name: merge-from-allow-all
mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
  - targetRef:
      kind: Mesh
    default:
      action: Allow
`, identityMesh)))).To(Succeed())
		trafficAllowed("merge-identity-demo-client", addr)

		// when MeshIdentity is enabled
		Expect(universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshIdentity
name: merge-identity
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: %s
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, identityMesh, trustDomain)))).To(Succeed())
		Eventually(func(g Gomega) {
			output, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshidentity", "-m", identityMesh, "merge-identity", "-o", "json")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(strings.Contains(output, "Successfully initialized")).To(BeTrue())
		}, "30s", "1s").Should(Succeed())

		// and a 'rules' policy denies the client, which proves RBAC runs on identity certificates
		Expect(universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: MeshTrafficPermission
name: merge-identity-rules-deny
mesh: %s
spec:
  targetRef:
    kind: Mesh
  rules:
  - default:
      deny:
      - spiffeID:
          type: Exact
          value: spiffe://%s/workload/merge-identity-demo-client
`, identityMesh, trustDomain)))).To(Succeed())

		// then
		trafficBlocked("merge-identity-demo-client", addr)

		// when the 'rules' deny is removed
		Expect(universal.Cluster.GetKumactlOptions().KumactlDelete("meshtrafficpermission", "merge-identity-rules-deny", identityMesh)).To(Succeed())

		// then the 'from' allow keeps traffic flowing without a 'rules' allow-all
		trafficAllowed("merge-identity-demo-client", addr)
	})
}
