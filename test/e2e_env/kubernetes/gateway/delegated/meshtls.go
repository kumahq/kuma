package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtls/api/v1alpha1"
	"github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envs/kubernetes"
)

func MeshTLS(config *Config) func() {
	GinkgoHelper()

	return func() {
		meshTls := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTLS
metadata:
  name: meshtls-delegated
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  rules:
    - default:
        tlsVersion:
          min: TLS13
          max: TLS13`, config.CpNamespace, config.Mesh)

		framework.AfterEachFailure(func() {
			framework.DebugKube(kubernetes.Cluster, config.Mesh, config.Namespace, config.ObservabilityDeploymentName)
		})

		framework.E2EAfterEach(func() {
			Expect(framework.DeleteMeshResources(
				kubernetes.Cluster,
				config.Mesh,
				v1alpha1.MeshTLSResourceTypeDescriptor,
			)).To(Succeed())
		})

		It("should not break communication once switched to TLS 1.3", func() {
			// check that communication to test-server works
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s", MustPassRepeatedly(5)).Should(Succeed())

			// change TLS version to 1.3
			Expect(framework.YamlK8s(meshTls)(kubernetes.Cluster)).To(Succeed())

			// check that the policy reached the data path the gateway uses.
			// The gateway proxy has no inbounds of its own, so mesh traffic
			// entering through it is secured by the backend's inbound. The
			// gateway load balances over every replica, so a request fails
			// until the last of them has the policy, not the first.
			Eventually(func(g Gomega) {
				for i := range TestServerReplicas {
					stdout, err := kubernetes.Cluster.GetKumactlOptions().RunKumactlAndGetOutput(
						"inspect", "dataplane",
						"-m", config.Mesh,
						fmt.Sprintf("test-server-%d.%s", i, config.Namespace),
						"--type=config-dump",
					)
					g.Expect(err).ToNot(HaveOccurred())
					g.Expect(stdout).To(ContainSubstring(`"tls_minimum_protocol_version": "TLSv1_3"`))
					g.Expect(stdout).To(ContainSubstring(`"tls_maximum_protocol_version": "TLSv1_3"`))
				}
			}, "60s", "1s").Should(Succeed())

			// check that communication to test-server works
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster,
					"demo-client",
					fmt.Sprintf("http://%s/test-server", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s", MustPassRepeatedly(5)).Should(Succeed())
		})
	}
}
