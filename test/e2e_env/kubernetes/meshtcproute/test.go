package meshtcproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func Test() {
	meshName := "meshtcproute"
	namespace := "meshtcproute"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			// MeshExternalService traffic is only ever routed through ZoneEgress
			// (see docs/madr/decisions/062-meshexternalservice-and-zoneegress.md),
			// so the mesh needs mTLS + egress routing enabled for the
			// MeshTCPRoute-vs-MeshExternalService precedence case below to have
			// a real backend to route to.
			Install(Combine(
				YamlK8s(samples.MeshMTLSBuilder().WithName(meshName).WithEgressRoutingEnabled().KubeYaml()),
				WaitMeshKubernetesReady(meshName),
			)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithName("test-client"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("test-http-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("test-http-server-2"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("test-tcp-server"),
					testserver.WithServicePortAppProtocol("tcp"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("external-tcp-service"),
					testserver.WithNamespace(namespace),
				),
			)).
			Setup(kubernetes.Cluster),
		).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(
			kubernetes.Cluster,
			meshName,
			v1alpha1.MeshTCPRouteResourceTypeDescriptor,
		)).To(Succeed())
		Expect(DeleteMeshResources(
			kubernetes.Cluster,
			meshName,
			meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
		)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use MeshTCPRoute if no TrafficRoutes are present", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster,
				"test-client",
				"test-http-server.meshtcproute.svc.cluster.local",
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("test-http-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should use MeshHTTPRoute if both MeshTCPRoute and MeshHTTPRoute "+
		"are present and point to the same source and http destination", func() {
		// given
		Expect(kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-tcp-service-mtcpr
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: tcp
  endpoints:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    - address: external-tcp-service.%s.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName, namespace)))).To(Succeed())

		// when
		meshRoutes := func(meshName string) string {
			meshHTTPRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshService
      name: test-http-server
      namespace: %s
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/"
      default:
        backendRefs:
        - kind: MeshService
          name: test-http-server-2
          namespace: %s
          port: 80
`, namespace, meshName, namespace, namespace)
			meshTCPRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTCPRoute
metadata:
  name: tcp-route-external-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshService
      name: test-http-server
      namespace: %s
    rules:
    - default:
        backendRefs:
        - kind: MeshExternalService
          name: external-tcp-service-mtcpr
          # MeshExternalService resources live in the CP's own namespace
          # (see how it's installed above); without an explicit namespace
          # here the backendRef falls back to this policy's namespace and
          # silently fails to resolve, producing an empty outbound listener.
          namespace: %s
          port: 80
`, namespace, meshName, namespace, Config.KumaNamespace)
			return fmt.Sprintf("%s\n---%s", meshTCPRoute, meshHTTPRoute)
		}

		Expect(YamlK8s(meshRoutes(meshName))(kubernetes.Cluster)).
			To(Succeed())

		// then
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster,
				"test-client",
				"test-http-server.meshtcproute.svc.cluster.local",
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("test-http-server-2"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should use MeshTCPRoute if both MeshTCPRoute and MeshHTTPRoute "+
		"are present and point to the same source and tcp destination", func() {
		// given
		Expect(kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-tcp-service-mtcpr
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: tcp
  endpoints:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    - address: external-tcp-service.%s.svc.cluster.local
      port: 80
`, Config.KumaNamespace, meshName, namespace)))).To(Succeed())

		// when
		meshRoutes := func(meshName string) string {
			meshHTTPRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: http-route
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshService
      name: test-tcp-server
      namespace: %s
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/"
      default:
        backendRefs:
        - kind: MeshService
          name: test-http-server-2
          namespace: %s
          port: 80
`, namespace, meshName, namespace, namespace)
			meshTCPRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTCPRoute
metadata:
  name: tcp-route-external
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
  - targetRef:
      kind: MeshService
      name: test-tcp-server
      namespace: %s
    rules:
    - default:
        backendRefs:
        - kind: MeshExternalService
          name: external-tcp-service-mtcpr
          # MeshExternalService resources live in the CP's own namespace
          # (see how it's installed above); without an explicit namespace
          # here the backendRef falls back to this policy's namespace and
          # silently fails to resolve, producing an empty outbound listener.
          namespace: %s
          port: 80
`, namespace, meshName, namespace, Config.KumaNamespace)
			return fmt.Sprintf("%s\n---%s", meshTCPRoute, meshHTTPRoute)
		}

		Expect(YamlK8s(meshRoutes(meshName))(kubernetes.Cluster)).
			To(Succeed())

		// then
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster,
				"test-client",
				"test-tcp-server.meshtcproute.svc.cluster.local",
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("external-tcp-service"))
		}, "30s", "1s").Should(Succeed())
	})
}
