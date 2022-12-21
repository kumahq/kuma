package meshcircuitbreaker

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
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

		Expect(DeleteMeshResources(env.Cluster, mesh,
			core_mesh.CircuitBreakerResourceTypeDescriptor,
			core_mesh.RetryResourceTypeDescriptor,
			v1alpha1.MeshCircuitBreakerResourceTypeDescriptor,
		)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(env.Cluster, mesh, v1alpha1.MeshCircuitBreakerResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	DescribeTable("should configure circuit breaker limits and outlier"+
		" detectors for connections", func(config string) {
		// given no MeshCircuitBreaker
		mcbs, err := env.Cluster.GetKumactlOptions().KumactlList("meshcircuitbreakers", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcbs).To(HaveLen(0))

		Eventually(func() ([]FailureResponse, error) {
			return CollectResponsesAndFailures(
				env.Cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				FromKubernetesPod(namespace, "demo-client"),
				WithNumberOfRequests(10),
			)
		}, "30s", "1s").Should(And(
			HaveLen(10),
			HaveEach(HaveField("ResponseCode", 200)),
		))

		// when
		Expect(env.Cluster.Install(YamlK8s(config))).To(Succeed())

		// then
		Eventually(func(g Gomega) ([]FailureResponse, error) {
			return CollectResponsesAndFailures(
				env.Cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				FromKubernetesPod(namespace, "demo-client"),
				WithNumberOfRequests(10),
				// increase processing time of a request to increase a probability of triggering maxPendingRequest limit
				WithHeader("x-set-response-delay-ms", "1000"),
				WithoutRetries(),
			)
		}, "30s", "1s").Should(And(
			HaveLen(10),
			ContainElement(HaveField("ResponseCode", 503)),
		))
	},
		Entry("outbound circuit breaker", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: mcb-outbound
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
  name: mcb-inbound
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
