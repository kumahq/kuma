package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshLoadBalancingStrategy(config *Config) func() {
	GinkgoHelper()

	return func() {
		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshLoadBalancingStrategyResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should use ring hash load balancing strategy", func() {
			Eventually(func(g Gomega) {
				responses, err := client.CollectResponsesByInstance(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(responses).To(HaveLen(3))
			}, "30s", "500ms").Should(Succeed())

			Expect(framework.YamlK8s(fmt.Sprintf(`
kind: MeshLoadBalancingStrategy
apiVersion: kuma.io/v1alpha1
metadata:
  name: ring-hash-delegated
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
      default:
        loadBalancer:
          type: RingHash
          ringHash:
            hashPolicies:
              - type: Header
                header:
                  name: x-header
`, config.CpNamespace, config.Mesh))(kubernetes.Cluster)).To(Succeed())

			Eventually(func(g Gomega) {
				responses, err := client.CollectResponsesByInstance(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
					client.WithHeader("x-header", "value"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(responses).To(HaveLen(1))
			}, "30s", "500ms").Should(Succeed())
		})
	}
}
