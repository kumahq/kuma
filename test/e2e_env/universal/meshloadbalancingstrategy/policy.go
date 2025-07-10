package meshloadbalancingstrategy

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshloadbalancingstrategy_api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "mesh-load-balancing-strategy"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(Yaml(samples.MeshDefaultBuilder().
				WithName(meshName).
				WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			)).
			Install(TestServerUniversal("test-server-1", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(TestServerUniversal("test-server-2", meshName,
				WithArgs([]string{"echo", "--instance", "universal-2"}))).
			Install(TestServerUniversal("test-server-3", meshName,
				WithArgs([]string{"echo", "--instance", "universal-3"}))).
			Install(DemoClientUniversal("demo-client", meshName,
				WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(universal.Cluster, meshName, meshloadbalancingstrategy_api.MeshLoadBalancingStrategyResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(universal.Cluster, meshName, meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
	})

	It("should use ring hash load balancing strategy", func() {
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(3))
		}, "30s", "500ms").Should(Succeed())

		Expect(YamlUniversal(`
type: MeshLoadBalancingStrategy
name: ring-hash
mesh: mesh-load-balancing-strategy
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        loadBalancer:
          type: RingHash
          ringHash:
            hashPolicies:
              - type: Header
                header:
                  name: x-header
`)(universal.Cluster)).To(Succeed())

		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
				client.WithHeader("x-header", "value"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(1))
		}, "30s", "500ms").Should(Succeed())
	})

	It("should use ring hash load balancing strategy with MeshHTTPRoute", func() {
		meshHttpRoute := fmt.Sprintf(`
type: MeshHTTPRoute
mesh: %s
name: test-server-route
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /hash-lb
          default: {}`, meshName)

		mlbsForMS := fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: lb-test-server
mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        hashPolicies:
          - type: Header
            header:
              name: x-common-ms-header
        loadBalancer:
          type: RingHash`, meshName)

		Expect(universal.Cluster.Install(YamlUniversal(meshHttpRoute))).To(Succeed())
		Expect(universal.Cluster.Install(YamlUniversal(mlbsForMS))).To(Succeed())

		By("Checking that requests to /no-hash-lb are balanced across all 3 instances")
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local/no-hash-lb")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(3))
		}, "30s", "500ms").Should(Succeed())

		By("Checking that requests to /hash-lb are balanced across all 3 instances")
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local/hash-lb")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(3))
		}, "30s", "500ms").Should(Succeed())

		mlbsForRoute := fmt.Sprintf(`
type: MeshLoadBalancingStrategy
name: ring-hash
mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshHTTPRoute
        name: test-server-route
      default:
        hashPolicies:
          - type: Header
            header:
              name: x-special-route-header`, meshName)

		Expect(universal.Cluster.Install(YamlUniversal(mlbsForRoute))).To(Succeed())

		By("Checking that requests to /no-hash-lb are still balanced across all 3 instances")
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local/no-hash-lb",
				client.WithHeader("x-special-route-header", "value"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(3))
		}, "30s", "500ms").Should(Succeed())

		By("Checking that requests to /hash-lb with header are balanced to only 1 instance")
		Eventually(func(g Gomega) {
			responses, err := client.CollectResponsesByInstance(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local/hash-lb",
				client.WithHeader("x-special-route-header", "value"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(responses).To(HaveLen(1))
		}, "30s", "500ms").Should(Succeed())
	})
}
