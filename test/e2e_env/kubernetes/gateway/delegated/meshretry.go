package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshRetry(config *Config) func() {
	GinkgoHelper()

	return func() {
		meshRetryPolicy := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshRetry
metadata:
  name: mr
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
        http:
          numRetries: 6
          retryOn: ["503"]
`, config.CpNamespace, config.Mesh)

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshRetryResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should retry on HTTP 503", func() {
			// Given no MeshRetry policies
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshRetryResourceTypeDescriptor,
			)).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithHeader("x-succeed-after-n", "100"),
					client.WithHeader("x-succeed-after-n-id", "1"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(503))
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())

			// and when a MeshRetry policy is applied
			Expect(framework.YamlK8s(meshRetryPolicy)(kubernetes.Cluster)).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithHeader("x-succeed-after-n", "5"),
					client.WithHeader("x-succeed-after-n-id", "2"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
		})
	}
}
