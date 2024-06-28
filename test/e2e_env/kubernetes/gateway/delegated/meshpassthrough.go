package delegated

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshPassthrough(config *Config) func() {
	GinkgoHelper()

	return func() {
		AfterEach(func() {
			Expect(framework.DeleteMeshResources(kubernetes.Cluster, config.Mesh, v1alpha1.MeshPassthroughResourceTypeDescriptor)).To(Succeed())
		})

		// we cannot test in a stable way first doing a successful request, because once connection is estabilished we don't break it after applying the policy (similar case to egress) the connection exists and works.
		It("should control traffic to domains and IPs", func() {
			esIP, err := framework.ServiceIP(kubernetes.Cluster, "external-service", config.NamespaceOutsideMesh)
			Expect(err).ToNot(HaveOccurred())
			// given
			meshPassthrough := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshPassthrough
metadata:
  name: disable-passthrough
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshSubset
    proxyTypes: ["Gateway"]
    tags:
      kuma.io/service: delegated-gateway-admin_delegated-gateway_svc_8444
  default:
    passthroughMode: None
`, config.CpNamespace, config.Mesh)

			// when
			err = kubernetes.Cluster.Install(framework.YamlK8s(meshPassthrough))

			// we need to wait for a config to arrive because once request is done, connection is estabilished and it won't return 502
			time.Sleep(5 * time.Second)

			// then
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				resp, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					fmt.Sprintf("http://%s/external-service", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.ResponseCode).To(Equal(502))
			}, "30s", "1s").Should(Succeed())

			meshPassthrough = fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshPassthrough
metadata:
  name: allow-specified
  namespace: %s
  labels:
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: MeshSubset
    proxyTypes: ["Gateway"]
    tags:
      kuma.io/service: delegated-gateway-admin_delegated-gateway_svc_8444
  default:
    passthroughMode: Matched
    appendMatch:
    - type: IP
      value: %s
      port: 80
      protocol: tcp
`, config.CpNamespace, config.Mesh, esIP)

			// when
			err = kubernetes.Cluster.Install(framework.YamlK8s(meshPassthrough))

			// then
			Expect(err).ToNot(HaveOccurred())
			Eventually(func(g Gomega) {
				_, err := client.CollectEchoResponse(
					kubernetes.Cluster, "demo-client",
					fmt.Sprintf("http://%s/external-service", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
			}, "30s", "1s").Should(Succeed())

			Eventually(func(g Gomega) {
				resp, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					fmt.Sprintf("http://%s/another-external-service", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.ResponseCode).To(Equal(502))
			}, "30s", "1s").Should(Succeed())
		})
	}
}
