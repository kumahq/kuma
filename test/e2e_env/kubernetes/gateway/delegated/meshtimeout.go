package delegated

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshTimeout(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshTimeoutResourceTypeDescriptor,
			)).To(Succeed())
		})

		DescribeTable("should add timeouts", FlakeAttempts(3), func(timeoutConfig string) {
			// given no MeshTimeout
			mts, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("meshtimeouts", config.Mesh)
			Expect(err).ToNot(HaveOccurred())
			Expect(mts).To(BeEmpty())

			Eventually(func(g Gomega) {
				start := time.Now()
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(time.Since(start)).To(BeNumerically(">", 5*time.Second))
			}, "30s", "1s").Should(Succeed())

			// when
			Expect(framework.YamlK8s(timeoutConfig)(kubernetes.Cluster)).To(Succeed())

			// then
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithHeader("x-set-response-delay-ms", "5000"),
					client.WithMaxTime(10), // we don't want 'curl' to return early
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(504))
			}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())

			Expect(framework.DeleteYamlK8s(timeoutConfig)(kubernetes.Cluster)).To(Succeed())
		},
			Entry("outbound timeout", fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTimeout
metadata:
  name: mt1-delegated
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
        idleTimeout: 20s
        http:
          requestTimeout: 2s
          maxStreamDuration: 20s`, config.CpNamespace, config.Mesh)),
		)
	}
}
