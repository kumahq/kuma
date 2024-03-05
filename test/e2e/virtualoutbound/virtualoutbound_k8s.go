package virtualoutbound

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func VirtualOutboundOnK8s() {
	var k8sCluster Cluster
	namespace := "virtual-outbound"

	BeforeEach(func() {
		k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

		err := NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithEnv("KUMA_DNS_SERVER_SERVICE_VIP_ENABLED", "false"),
			)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh("default"))).
			Install(testserver.Install(testserver.WithStatefulSet(true), testserver.WithReplicas(2), testserver.WithNamespace(namespace))).
			Setup(k8sCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DeleteNamespace(namespace)).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})

	It("doesn't support default vips", func() {
		virtualOutboundAll := `
apiVersion: kuma.io/v1alpha1
kind: VirtualOutbound
mesh: default
metadata:
  name: instance
spec:
  selectors:
  - match:
      kuma.io/service: "*"
  conf:
    host: "{{.svc}}.foo"
    port: "8080"
    parameters:
    - name: "svc"
      tagKey: "kuma.io/service"
`
		err := YamlK8s(virtualOutboundAll)(k8sCluster)
		Expect(err).ToNot(HaveOccurred())

		// Succeed with virtual-outbound
		Eventually(func(g Gomega) {
			res, err := client.CollectEchoResponse(k8sCluster, "demo-client", "test-server_virtual-outbound_svc_80.foo:8080",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(res.Instance).To(Or(Equal("test-server-1"), Equal("test-server-0")))
		}, "30s", "1s").Should(Succeed())

		// Fails with built in vip (it's disabled in conf)
		Consistently(func(g Gomega) {
			_, err := client.CollectEchoResponse(k8sCluster, "demo-client", "test-server_virtual-outbound_svc_80.mesh:80",
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).To(HaveOccurred())
		}, "2s", "250ms").Should(Succeed())
	})
}
