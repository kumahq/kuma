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
	"github.com/kumahq/kuma/v3/test/server/types"
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
      name: test-server_%[2]s_svc_80
    rules: 
    - default:
        backendRefs:
        - kind: MeshService
          name: test-server_%[2]s_svc_80
        - kind: MeshExternalService
          name: external-http-service-mtcprd
          port: 80
        - kind: MeshExternalService
          name: external-tcp-service-mtcprd
          port: 80
`, config.CpNamespace, config.Mesh))(kubernetes.Cluster)).To(Succeed())

			// then
			Eventually(func() ([]types.EchoResponse, error) {
				return client.CollectResponses(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithNumberOfRequests(10),
					client.WithHeader("x-set-response-delay-ms", "2000"),
					client.WithoutRetries(),
				)
			}, "30s", "1s").Should(ContainElements(
				HaveField("Instance", HavePrefix("test-server")),
				HaveField("Instance", HavePrefix("external-service")),
				HaveField("Instance", HavePrefix("external-tcp-service")),
			))
		})
	}
}
