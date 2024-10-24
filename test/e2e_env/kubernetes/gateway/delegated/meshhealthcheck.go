package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshHealthCheck(config *Config) func() {
	GinkgoHelper()

	return func() {
		healthCheck := func(path, status string) string {
			return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHealthCheck
metadata:
  name: mhc
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
        interval: 3s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        failTrafficOnPanic: true
        noTrafficInterval: 1s
        healthyPanicThreshold: 0
        reuseConnection: true
        http: 
          path: %s
          expectedStatuses: 
          - %s`, config.CpNamespace, config.Mesh, path, status)
		}

		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshHealthCheckResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should mark host as unhealthy if it doesn't reply on health checks", func() {
			// check that test-server is healthy
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}).Should(Succeed())

			// update HealthCheck policy to check for another status code
			Expect(framework.YamlK8s(healthCheck("/are-you-healthy", "500"))(kubernetes.Cluster)).
				To(Succeed())

			// check that test-server is unhealthy
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "30s", "1s").Should(Succeed())
		})
	}
}
