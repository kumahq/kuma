package meshhttproute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func Test() {
	meshName := "meshhttproute"
	namespace := "meshhttproute"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				testserver.Install(
					testserver.WithName("test-client"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithName("test-server"),
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithNamespace(namespace),
					testserver.WithName("external-service"),
				),
			)).
			Setup(kubernetes.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, v1alpha1.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(kubernetes.Cluster, meshName, core_mesh.ExternalServiceResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should use MeshHTTPRoute if no TrafficRoutes are present", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(kubernetes.Cluster, "test-client", "test-server_meshhttproute_svc_80.mesh", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should split traffic between internal and external services", func() {
		// given
		Expect(kubernetes.Cluster.Install(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
metadata:
  name: external-service-mhr
mesh: %s
spec:
  tags:
    kuma.io/service: external-service-mhr
    kuma.io/protocol: http
  networking:
    address: external-service.%s.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
`, meshName, namespace)))).To(Succeed())

		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
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
        name: test-server_meshhttproute_svc_80
      rules: 
        - matches:
            - path: 
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshService
                name: test-server_meshhttproute_svc_80
                weight: 50
              - kind: MeshService
                name: external-service-mhr
                weight: 50
`, Config.KumaNamespace, meshName, meshName))(kubernetes.Cluster)).To(Succeed())

		// then receive responses from 'test-server_meshhttproute_svc_80'
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(kubernetes.Cluster, "test-client", "test-server_meshhttproute_svc_80.mesh", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("test-server"))
		}, "30s", "1s").Should(Succeed())

		// and then receive responses from 'external-service'
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(kubernetes.Cluster, "test-client", "test-server_meshhttproute_svc_80.mesh", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(HavePrefix("external-service"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should configure redirect response", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-3
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
        name: test-server_meshhttproute_svc_80
      rules: 
        - matches:
            - path: 
                type: PathPrefix
                value: /
          default:
            filters:
              - type: RequestRedirect
                requestRedirect:
                  statusCode: 307
                  path:
                    type: ReplaceFullPath
                    replaceFullPath: /new-path
            backendRefs:
              - kind: MeshService
                name: test-server_meshhttproute_svc_80
                weight: 1
`, Config.KumaNamespace, meshName, meshName))(kubernetes.Cluster)).To(Succeed())

		// then receive redirect response
		Eventually(func(g Gomega) {
			failure, err := client.CollectFailure(kubernetes.Cluster, "test-client", "test-server_meshhttproute_svc_80.mesh", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(failure.ResponseCode).To(Equal(307))
			g.Expect(failure.RedirectURL).To(Equal("http://test-server_meshhttproute_svc_80.mesh/new-path"))
		}, "30s", "1s").Should(Succeed())
	})

	It("should configure URLRewrite", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-3
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
        name: test-server_meshhttproute_svc_80
      rules: 
        - matches:
            - path: 
                type: PathPrefix
                value: /prefix
          default:
            filters:
              - type: URLRewrite
                urlRewrite:
                  path:
                    type: ReplacePrefixMatch
                    replacePrefixMatch: /hello/
            backendRefs:
              - kind: MeshService
                name: test-server_meshhttproute_svc_80
                weight: 1
`, Config.KumaNamespace, meshName, meshName))(kubernetes.Cluster)).To(Succeed())

		// then receive redirect response
		Eventually(func(g Gomega) {
			resp, err := client.CollectEchoResponse(kubernetes.Cluster, "test-client", "test-server_meshhttproute_svc_80.mesh/prefix/world", client.FromKubernetesPod(namespace, "test-client"))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Received.Path).To(Equal("/hello/world"))
		}, "30s", "1s").Should(Succeed())
	})
}
