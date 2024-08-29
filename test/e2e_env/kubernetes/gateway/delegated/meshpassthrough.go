package delegated

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshpassthrough_api "github.com/kumahq/kuma/pkg/plugins/policies/meshpassthrough/api/v1alpha1"
	meshproxypatch_api "github.com/kumahq/kuma/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	"github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func MeshPassthrough(config *Config) func() {
	GinkgoHelper()

	return func() {
		AfterEach(func() {
			Expect(framework.DeleteMeshResources(kubernetes.Cluster, config.Mesh, meshpassthrough_api.MeshPassthroughResourceTypeDescriptor)).To(Succeed())
			Expect(framework.DeleteMeshResources(kubernetes.Cluster, config.Mesh, meshproxypatch_api.MeshProxyPatchResourceTypeDescriptor)).To(Succeed())
		})

		// we cannot test in a stable way first doing a successful request, because once connection is estabilished we don't break it after applying the policy (similar case to egress) the connection exists and works.
		It("should control traffic to domains and IPs", func() {
			esIP, err := kubernetes.Cluster.GetClusterIP("external-service", config.NamespaceOutsideMesh)
			Expect(err).ToNot(HaveOccurred())

			// given
			// we want the connection to close quickly, so after configuration change requests will start to fail.
			meshProxyPatch := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshProxyPatch
metadata:
  name: delegated-idle-connection
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
    appendModifications:
      - networkFilter:
          operation: Patch
          match:
            name: envoy.filters.network.tcp_proxy
            listenerName: outbound:passthrough:ipv4
          jsonPatches: # optional and mutually exclusive with "value": list of modifications in JSON Patch notation
            - op: add
              path: /idleTimeout
              value: 0.5s
`, config.CpNamespace, config.Mesh)
			err = kubernetes.Cluster.Install(framework.YamlK8s(meshProxyPatch))
			Expect(err).ToNot(HaveOccurred())

			// and
			meshPassthrough := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshPassthrough
metadata:
  name: disable-passthrough-delegated
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
			Expect(err).ToNot(HaveOccurred())

			// then
			Eventually(func(g Gomega) {
				resp, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					fmt.Sprintf("http://%s/external-service", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.ResponseCode).To(Equal(502))
			}, "60s", "1s", MustPassRepeatedly(3)).Should(Succeed())

			meshPassthrough = fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1 
kind: MeshPassthrough
metadata:
  name: allow-specified-delegated
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
			}, "60s", "1s", MustPassRepeatedly(3)).Should(Succeed())

			Consistently(func(g Gomega) {
				resp, err := client.CollectFailure(
					kubernetes.Cluster, "demo-client",
					fmt.Sprintf("http://%s/another-external-service", config.KicIP),
					client.FromKubernetesPod(config.NamespaceOutsideMesh, "demo-client"),
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(resp.ResponseCode).To(Equal(502))
			}, "30s", "1s", MustPassRepeatedly(3)).Should(Succeed())
		})
	}
}
