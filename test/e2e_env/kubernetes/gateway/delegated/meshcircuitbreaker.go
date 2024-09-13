package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func CircuitBreaker(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshCircuitBreakerResourceTypeDescriptor,
			)).To(Succeed())
		})

		DescribeTable("should configure circuit breaker limits and outlier"+
			" detectors for connections", func(yaml string) {
			// given no MeshCircuitBreaker
			mcbs, err := kubernetes.Cluster.GetKumactlOptions().
				KumactlList("meshcircuitbreakers", config.Mesh)
			Expect(err).ToNot(HaveOccurred())
			Expect(mcbs).To(BeEmpty())

			Eventually(func() ([]client.FailureResponse, error) {
				return client.CollectResponsesAndFailures(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithNumberOfRequests(10),
				)
			}, "30s", "1s").Should(And(
				HaveLen(10),
				HaveEach(HaveField("ResponseCode", 200)),
			))

			// when
			Expect(kubernetes.Cluster.Install(framework.YamlK8s(yaml))).To(Succeed())

			// then
			Eventually(func(g Gomega) ([]client.FailureResponse, error) {
				return client.CollectResponsesAndFailures(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithNumberOfRequests(10),
					// increase processing time of a request to increase
					// a probability of triggering maxPendingRequest limit
					client.WithHeader("x-set-response-delay-ms", "1000"),
					client.WithoutRetries(),
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
  name: mcb-outbound-delegated
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
`, config.CpNamespace, config.Mesh)),
		)
	}
}
