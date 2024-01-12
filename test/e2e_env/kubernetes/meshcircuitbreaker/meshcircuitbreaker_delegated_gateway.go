package meshcircuitbreaker

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/kic"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshCircuitBreakerDelegatedGateway() {
	// IPv6 currently not supported by Kong Ingress Controller
	// https://github.com/Kong/kubernetes-ingress-controller/issues/1017
	if Config.IPV6 {
		fmt.Println("Test not supported on IPv6")
		return
	}
	if Config.K8sType == KindK8sType {
		// KIC 2.0 when started with service type LoadBalancer requires external
		// IP to be provisioned before it's healthy. KIND cannot provision
		// external IP, K3D can.
		fmt.Println("Test not supported on KIND")
		return
	}

	namespace := "meshcircuitbreaker"
	namespaceOutsideMesh := "meshcircuitbreaker-outside-mesh"
	mesh := "meshcircuitbreaker"

	var kicIP string

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MTLSMeshKubernetes(mesh)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceOutsideMesh)).
			Install(democlient.Install(
				democlient.WithNamespace(namespaceOutsideMesh),
			)).
			Install(testserver.Install(
				testserver.WithMesh(mesh),
				testserver.WithNamespace(namespace),
				testserver.WithName("test-server"),
			)).
			Install(kic.KongIngressController(
				kic.WithNamespace(namespace),
				kic.WithMesh(mesh),
			)).
			Install(kic.KongIngressService(kic.WithNamespace(namespace))).
			Install(YamlK8s(fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  namespace: %s
  name: %s-ingress
  annotations:
    kubernetes.io/ingress.class: kong
spec:
  rules:
  - http:
      paths:
      - path: /test-server
        pathType: Prefix
        backend:
          service:
            name: test-server
            port:
              number: 80
`, namespace, mesh))).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Expect(DeleteMeshResources(kubernetes.Cluster, mesh,
			core_mesh.CircuitBreakerResourceTypeDescriptor,
			core_mesh.RetryResourceTypeDescriptor,
			v1alpha1.MeshCircuitBreakerResourceTypeDescriptor,
		)).To(Succeed())

		kicIP, err = kic.From(kubernetes.Cluster).IP(namespace)
		Expect(err).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(
			kubernetes.Cluster,
			mesh,
			v1alpha1.MeshCircuitBreakerResourceTypeDescriptor,
		)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).
			To(Succeed())
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespaceOutsideMesh)).
			To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	DescribeTable("should configure circuit breaker limits and outlier"+
		" detectors for connections", func(config string) {
		// given no MeshCircuitBreaker
		mcbs, err := kubernetes.Cluster.GetKumactlOptions().
			KumactlList("meshcircuitbreakers", mesh)
		Expect(err).ToNot(HaveOccurred())
		Expect(mcbs).To(BeEmpty())

		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				kubernetes.Cluster,
				"demo-client",
				fmt.Sprintf("http://%s/test-server", kicIP),
				client.FromKubernetesPod(namespaceOutsideMesh, "demo-client"),
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
				fmt.Sprintf("http://%s/test-server", kicIP),
				client.FromKubernetesPod(namespaceOutsideMesh, "demo-client"),
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
