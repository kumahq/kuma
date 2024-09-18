package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
	"github.com/kumahq/kuma/test/server/types"
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
				core_mesh.ExternalServiceResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should split traffic between internal and external services "+
			"with mixed (tcp and http) protocols", func() {
			// given
			Expect(kubernetes.Cluster.Install(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
metadata:
  name: external-http-service-mtcprd
mesh: %s
spec:
  tags:
    kuma.io/service: external-http-service-mtcprd
    kuma.io/protocol: http
  networking:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    address: external-service.%s.svc.cluster.local:80
`, config.Mesh, config.NamespaceOutsideMesh)))).To(Succeed())

			Expect(kubernetes.Cluster.Install(framework.YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: ExternalService
metadata:
  name: external-tcp-service-mtcprd
mesh: %s
spec:
  tags:
    kuma.io/service: external-tcp-service-mtcprd
    kuma.io/protocol: tcp
  networking:
    # .svc.cluster.local is needed, otherwise Kubernetes will resolve this
    # to the real IP
    address: external-tcp-service.%s.svc.cluster.local:80
`, config.Mesh, config.NamespaceOutsideMesh)))).To(Succeed())

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
    kind: MeshService
    name: %[2]s-gateway-admin_%[2]s_svc_8444
  to:
  - targetRef:
      kind: MeshService
      name: test-server_%[2]s_svc_80
    rules: 
    - default:
        backendRefs:
        - kind: MeshService
          name: test-server_%[2]s_svc_80
        - kind: MeshService
          name: external-http-service-mtcprd
        - kind: MeshService
          name: external-tcp-service-mtcprd
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
