package meshcircuitbreaker

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshCircuitBreaker() {
	namespace := "meshcircuitbreaker-namespace"
	mesh := "meshcircuitbreaker"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(
				Yaml(
					builders.Mesh().
						WithName(mesh).
						WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Everywhere),
				),
			).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(mesh)),
				testserver.Install(testserver.WithMesh(mesh), testserver.WithNamespace(namespace)),
			)).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: HostnameGenerator
metadata:
  name: circuitbreaker-connectivity
  namespace: %s
spec:
  template: '{{ .DisplayName }}.{{ .Namespace }}.{{ .Zone }}.meshcircuitbreaker'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/managed-by: k8s-controller
`, Config.KumaNamespace))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			meshretry_api.MeshRetryResourceTypeDescriptor,
			fmt.Sprintf("mesh-retry-all-%s", mesh),
		)).To(Succeed())

		Expect(DeleteMeshPolicyOrError(
			kubernetes.Cluster,
			v1alpha1.MeshCircuitBreakerResourceTypeDescriptor,
			fmt.Sprintf("mesh-circuit-breaker-all-%s", mesh),
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, mesh, namespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, mesh,
			v1alpha1.MeshCircuitBreakerResourceTypeDescriptor,
			meshretry_api.MeshRetryResourceTypeDescriptor,
		)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	DescribeTable("should configure circuit breaker limits and outlier"+
		" detectors for connections", func(config string) {
		// given no MeshCircuitBreaker
		mcbs, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshcircuitbreakers", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcbs).To(BeEmpty())

		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				kubernetes.Cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithNumberOfRequests(10),
			)
		}, "30s", "1s").Should(And(
			HaveLen(10),
			HaveEach(HaveField("ResponseCode", 200)),
		))

		// when
		Expect(kubernetes.Cluster.Install(YamlK8s(config))).To(Succeed())

		// then
		Eventually(func(g Gomega) ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				kubernetes.Cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithNumberOfRequests(10),
				// increase processing time of a request to increase a probability of triggering maxPendingRequest limit
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
  rules:
    - default:
        connectionLimits:
          maxConnectionPools: 1
          maxConnections: 1
          maxPendingRequests: 1
          maxRequests: 1
          maxRetries: 1
`, Config.KumaNamespace, mesh)),
	)

	It("should configure circuit breaker limits and outlier detectors for connections", func() {
		// given no MeshCircuitBreaker
		mcbs, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshcircuitbreakers", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcbs).To(BeEmpty())

		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				kubernetes.Cluster,
				"demo-client",
				fmt.Sprintf("test-server.%s.default.meshcircuitbreaker", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithNumberOfRequests(10),
			)
		}, "30s", "1s").Should(And(
			HaveLen(10),
			HaveEach(HaveField("ResponseCode", 200)),
		))

		// when
		Expect(kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshCircuitBreaker
metadata:
  name: mcb-outbound-ms
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        namespace: %s
      default:
        connectionLimits:
          maxConnectionPools: 1
          maxConnections: 1
          maxPendingRequests: 1
          maxRequests: 1
          maxRetries: 1
        outlierDetection:
          interval: 1s
          baseEjectionTime: 30s
          maxEjectionPercent: 100
          detectors:
            totalFailures:
              consecutive: 1
          healthyPanicThreshold: 0`, namespace, mesh, namespace)))).To(Succeed())

		// then
		Eventually(func(g Gomega) ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				kubernetes.Cluster,
				"demo-client",
				fmt.Sprintf("test-server.%s.default.meshcircuitbreaker", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
				client.WithNumberOfRequests(10),
				// increase processing time of a request to increase a probability of triggering maxPendingRequest limit
				client.WithHeader("x-set-response-delay-ms", "1000"),
				client.WithoutRetries(),
			)
		}, "30s", "1s").Should(And(
			HaveLen(10),
			ContainElement(HaveField("ResponseCode", 503)),
		))

		// then
		// with the above 10 times 503, we're in panic mode.
		Consistently(func(g Gomega) ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				kubernetes.Cluster,
				"demo-client",
				fmt.Sprintf("test-server.%s.default.meshcircuitbreaker", namespace),
				// errors returned by circuit breaker don't count for outlier detection
				// only upstream returned errors
				client.WithHeader("x-set-response-code", "500"),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			// we expect 503 because we are in the panic mode
		}, "10s", "1s").Should(ContainElement(HaveField("ResponseCode", 503)))
	})
}
