package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshHTTPRoute(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshHTTPRouteResourceTypeDescriptor,
			)).To(Succeed())

			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				core_mesh.ExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should split traffic between internal and external services", func() {
			// given
			Expect(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
metadata:
  name: external-service-mhr-delegated
mesh: %s
spec:
  tags:
    kuma.io/service: external-service-mhr
    kuma.io/protocol: http
  networking:
    address: external-service.%s.svc.cluster.local:80 # .svc.cluster.local is needed, otherwise Kubernetes will resolve this to the real IP
`, config.Mesh, config.NamespaceOutsideMesh))(kubernetes.Cluster)).To(Succeed())

			// when
			Expect(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: mhr-delegated
  namespace: %s
  labels:
    kuma.io/mesh: %[2]s
spec:
  targetRef:
    kind: MeshService
    name: %[2]s-gateway-admin_%[2]s_svc_8444
  to:
    - targetRef:
        kind: MeshService
        name: test-server_%[2]s_svc_80
      rules: 
        - matches:
            - path: 
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshService
                name: test-server_%[2]s_svc_80
                weight: 50
              - kind: MeshService
                name: external-service-mhr
                weight: 50
`, config.CpNamespace, config.Mesh))(kubernetes.Cluster)).To(Succeed())

			// then receive responses from 'test-server_delegated-gateway_svc_80'
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(HavePrefix("test-server"))
			}, "30s", "1s").Should(Succeed())

			// and then receive responses from 'external-service'
			Eventually(func(g Gomega) {
				response, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.Instance).To(HavePrefix("external-service"))
			}, "30s", "1s").Should(Succeed())
		})
	}
}

func MeshHTTPRouteMeshService(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshHTTPRouteResourceTypeDescriptor,
			)).To(Succeed())

			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should split traffic between MeshService and MeshExternalServices", func() {
			// given
			Expect(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: plain-external-service-delegated-ms
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: external-service.%s.svc.cluster.local
      port: 80
`, config.CpNamespace, config.Mesh, config.NamespaceOutsideMesh))(kubernetes.Cluster)).To(Succeed())
			// when
			Expect(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: mhr-delegated-ms
  namespace: %s
  labels:
    kuma.io/mesh: %[2]s
spec:
  targetRef:
    kind: MeshSubset
    tags:
      kuma.io/service: %[2]s-gateway-admin_%[2]s_svc_8444
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        namespace: %[3]s
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /
          default:
            backendRefs:
              - kind: MeshService
                name: test-server
                namespace: %[3]s
                port: 80
                weight: 50
              - kind: MeshExternalService
                name: plain-external-service-delegated-ms
                port: 80
                weight: 50
`, config.CpNamespace, config.Mesh, config.Namespace))(kubernetes.Cluster)).To(Succeed())

			// and then receive responses from 'external-service'
			Eventually(func(g Gomega) {
				response, err := client.CollectResponsesByInstance(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithNumberOfRequests(100),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response).To(HaveLen(4))
				g.Expect(response).To(And(
					HaveKeyWithValue(Equal(`test-server-0`), BeNumerically("~", 16, 6)),
					HaveKeyWithValue(Equal(`test-server-1`), BeNumerically("~", 16, 6)),
					HaveKeyWithValue(Equal(`test-server-2`), BeNumerically("~", 16, 6)),
					HaveKeyWithValue(ContainSubstring(`external-service`), BeNumerically("~", 50, 15)),
				))
			}, "30s", "5s").Should(Succeed())
		})
	}
}
