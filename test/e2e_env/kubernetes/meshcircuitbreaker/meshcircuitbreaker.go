package meshcircuitbreaker

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/kubernetes/env"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func MeshCircuitBreaker() {
	namespace := "meshcircuitbreaker-namespace"
	mesh := "meshcircuitbreaker"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(DemoClientK8s(mesh, namespace)).
			Install(testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(namespace))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(DeleteMeshResourcesKubernetes(
			env.Cluster,
			mesh,
			"circuitbreakers",
			"retries",
			"meshcircuitbreakers",
		)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResourcesKubernetes(env.Cluster, mesh, "meshcircuitbreakers")).
			To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	DescribeTable("should add timeouts for outbound connections", func(config string) {
		// given no MeshCircuitBreaker
		mcbs, err := env.Cluster.GetKumactlOptions().
			KumactlList("meshcircuitbreakers", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcbs).To(HaveLen(0))

		Eventually(func() ([]FailureResponse, error) {
			return CollectResponsesAndFailures(
				env.Cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				FromKubernetesPod(namespace, "demo-client"),
				WithNumberOfRequests(50),
			)
		}, "30s", "500ms").Should(And(
			HaveLen(15),
			HaveEach(HaveField("ResponseCode", 200)),
		))

		// when
		Expect(YamlK8s(config)(env.Cluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) ([]FailureResponse, error) {
			return CollectResponsesAndFailures(
				env.Cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				FromKubernetesPod(namespace, "demo-client"),
				WithNumberOfRequests(50),
				WithoutRetries(),
			)
		}, "90s", "1s").Should(And(
			HaveLen(15),
			ContainElement(HaveField("ResponseCode", 503)),
		))
	},
		Entry("outbound circuit breaker", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: mcb1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        connectionLimits:
          maxConnectionPools: 1
          maxConnections: 1
          maxPendingRequests: 1
          maxRequests: 1
          maxRetries: 1
`, Config.KumaNamespace, mesh)),
		Entry("inbound circuit breaker", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: mcb1
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        connectionLimits:
          maxConnectionPools: 1
          maxConnections: 1
          maxPendingRequests: 1
          maxRequests: 1
          maxRetries: 1
`, Config.KumaNamespace, mesh)),
	)
}
