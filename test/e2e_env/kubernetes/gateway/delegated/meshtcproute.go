package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func MeshTCPRoute(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshTCPRouteResourceTypeDescriptor,
				meshexternalservice_api.MeshExternalServiceResourceTypeDescriptor,
			)).To(Succeed())

			Eventually(func(g Gomega) {
				responses, err := client.CollectResponses(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithNumberOfRequests(10),
					client.WithoutRetries(),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(responses).ToNot(BeEmpty())
				g.Expect(responses).To(HaveEach(HaveField("Instance", HavePrefix("test-server"))))
			}, "30s", "1s").Should(Succeed())
		})

		It("should split traffic between internal and external services "+
			"with mixed (tcp and http) protocols", func() {
			// given
			Expect(kubernetes.Cluster.Install(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-http-service-mtcprd
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    - address: external-service.%s.svc.cluster.local
      port: 80
`, config.CpNamespace, config.Mesh, config.NamespaceOutsideMesh)))).To(Succeed())

			Expect(kubernetes.Cluster.Install(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshExternalService
metadata:
  name: external-tcp-service-mtcprd
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
`, config.CpNamespace, config.Mesh, config.NamespaceOutsideMesh)))).To(Succeed())

			// when
			Expect(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTCPRoute
metadata:
  name: mtr
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Dataplane
    labels:
      app: %[2]s-gateway
  to:
  - targetRef:
      kind: MeshService
      labels:
        kuma.io/display-name: test-server
        k8s.kuma.io/namespace: %[3]s
    rules: 
    - default:
        backendRefs:
        - kind: MeshService
          labels:
            kuma.io/display-name: test-server
            k8s.kuma.io/namespace: %[3]s
          port: 80
        - kind: MeshExternalService
          labels:
            kuma.io/display-name: external-http-service-mtcprd
            k8s.kuma.io/namespace: %[1]s
          port: 80
        - kind: MeshExternalService
          labels:
            kuma.io/display-name: external-tcp-service-mtcprd
            k8s.kuma.io/namespace: %[1]s
          port: 80
`, config.CpNamespace, config.Mesh, config.Namespace))(kubernetes.Cluster)).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				response, err := client.CollectResponsesByInstance(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithNumberOfRequests(100),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response).To(HaveLen(3))
				g.Expect(response).To(And(
					HaveKey(Equal(`test-server`)),
					HaveKey(Equal(`external-service`)),
					HaveKey(Equal(`external-tcp-service`)),
				))
			}, "30s", "5s").Should(Succeed())
		})
	}
}
