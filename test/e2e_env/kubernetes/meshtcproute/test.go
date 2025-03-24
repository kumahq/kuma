package meshtcproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Test() {
	meshName := "meshtcproute"
	namespace := "meshtcproute"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshKubernetes(meshName)).
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
					testserver.WithName("external-http-service"),
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
			core_mesh.ExternalServiceResourceTypeDescriptor,
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
				"test-http-server_meshtcproute_svc_80.mesh",
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("test-http-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should split traffic between internal and external services "+
		"with mixed (tcp and http) protocols", func() {
		// given
		Expect(kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
metadata:
  name: external-http-service-mtcpr
mesh: %s
spec:
  tags:
    kuma.io/service: external-http-service-mtcpr
    kuma.io/protocol: http
  networking:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    address: external-http-service.%s.svc.cluster.local:80
`, meshName, namespace)))).To(Succeed())

		Expect(kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
metadata:
  name: external-tcp-service-mtcpr
mesh: %s
spec:
  tags:
    kuma.io/service: external-tcp-service-mtcpr
    kuma.io/protocol: tcp
  networking:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    address: external-tcp-service.%s.svc.cluster.local:80
`, meshName, namespace)))).To(Succeed())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTCPRoute
metadata:
  name: route-2
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: test-client_%s_svc_80
  to:
  - targetRef:
      kind: MeshService
      name: test-http-server_meshtcproute_svc_80
    rules: 
    - default:
        backendRefs:
        - kind: MeshService
          name: test-http-server_meshtcproute_svc_80
          weight: 25
        - kind: MeshService
          name: test-tcp-server_meshtcproute_svc_80
          weight: 25
        - kind: MeshService
          name: external-http-service-mtcpr
          weight: 25
        - kind: MeshService
          name: external-tcp-service-mtcpr
          weight: 25
`, Config.KumaNamespace, meshName, meshName))(kubernetes.Cluster)).To(Succeed())

		// then receive responses from "test-http-server_meshtcproute_svc_80"
		Eventually(func(g Gomega) {
			response, err := client.CollectResponsesByInstance(
				kubernetes.Cluster,
				"test-client",
				"test-http-server_meshtcproute_svc_80.mesh",
				client.FromKubernetesPod(namespace, "test-client"),
				client.WithNumberOfRequests(100),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response).To(HaveLen(4))
			g.Expect(response).To(And(
				//
				HaveKey(ContainSubstring("test-tcp-server")),
				HaveKey(ContainSubstring("test-http-server")),
				HaveKey(ContainSubstring("external-http-service")),
				HaveKey(ContainSubstring("external-tcp-service")),
			))
		}, "30s", "5s").Should(Succeed())
	})

	It("should use MeshHTTPRoute if both MeshTCPRoute and MeshHTTPRoute "+
		"are present and point to the same source and http destination", func() {
		// given
		Expect(kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
metadata:
  name: external-tcp-service-mtcpr
mesh: %s
spec:
  tags:
    kuma.io/service: external-tcp-service
    kuma.io/protocol: tcp
  networking:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    address: external-tcp-service.%s.svc.cluster.local:80
`, meshName, namespace)))).To(Succeed())

		// when
		meshRoutes := func(meshName string, namespace string) string {
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
    kind: MeshService
    name: test-client_%s_svc_80
  to:
  - targetRef:
      kind: MeshService
      name: test-http-server_meshtcproute_svc_80
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/"
      default:
        backendRefs:
        - kind: MeshService
          name: test-http-server-2_meshtcproute_svc_80
`, Config.KumaNamespace, meshName, namespace)
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
    kind: MeshService
    name: test-client_%s_svc_80
  to:
  - targetRef:
      kind: MeshService
      name: test-http-server_meshtcproute_svc_80
    rules:
    - default:
        backendRefs:
        - kind: MeshService
          name: external-tcp-service
`, Config.KumaNamespace, meshName, namespace)
			return fmt.Sprintf("%s\n---%s", meshTCPRoute, meshHTTPRoute)
		}

		Expect(YamlK8s(meshRoutes(meshName, namespace))(kubernetes.Cluster)).
			To(Succeed())

		// then
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster,
				"test-client",
				"test-http-server_meshtcproute_svc_80.mesh",
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
kind: ExternalService
metadata:
  name: external-tcp-service-mtcpr
mesh: %s
spec:
  tags:
    kuma.io/service: external-tcp-service
    kuma.io/protocol: tcp
  networking:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    address: external-tcp-service.%s.svc.cluster.local:80
`, meshName, namespace)))).To(Succeed())

		// when
		meshRoutes := func(meshName string, namespace string) string {
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
    kind: MeshService
    name: test-client_%s_svc_80
  to:
  - targetRef:
      kind: MeshService
      name: test-tcp-server_meshtcproute_svc_80
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: "/"
      default:
        backendRefs:
        - kind: MeshService
          name: test-http-server-2_meshtcproute_svc_80
`, Config.KumaNamespace, meshName, namespace)
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
    kind: MeshService
    name: test-client_%s_svc_80
  to:
  - targetRef:
      kind: MeshService
      name: test-tcp-server_meshtcproute_svc_80
    rules:
    - default:
        backendRefs:
        - kind: MeshService
          name: external-tcp-service
`, Config.KumaNamespace, meshName, namespace)
			return fmt.Sprintf("%s\n---%s", meshTCPRoute, meshHTTPRoute)
		}

		Expect(YamlK8s(meshRoutes(meshName, namespace))(kubernetes.Cluster)).
			To(Succeed())

		// then
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				kubernetes.Cluster,
				"test-client",
				"test-tcp-server_meshtcproute_svc_80.mesh",
				client.FromKubernetesPod(namespace, "test-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("external-tcp-service"))
		}, "30s", "1s").Should(Succeed())
	})
}
