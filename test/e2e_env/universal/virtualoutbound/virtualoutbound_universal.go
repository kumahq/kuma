package virtualoutbound

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func VirtualOutbound() {
	meshName := "virtual-outbound"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server1", meshName, WithArgs([]string{"echo", "--instance", "srv-1"}), WithServiceInstance("1"))).
			Install(TestServerUniversal("test-server2", meshName, WithArgs([]string{"echo", "--instance", "srv-2"}), WithServiceInstance("2"))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should add hostnames for individual instances", func() {
		virtualOutboundInstance := `
type: VirtualOutbound
mesh: virtual-outbound
name: instances 
selectors:
- match:
    kuma.io/service: "*"
    instance: "*"
conf:
    host: "{{.svc}}.{{.instance}}"
    port: "8080"
    parameters:
    - name: "svc"
      tagKey: "kuma.io/service"
    - name: "instance"
`
		err := YamlUniversal(virtualOutboundInstance)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		virtualOutboundAll := `
type: VirtualOutbound
mesh: virtual-outbound
name: all 
selectors:
- match:
    kuma.io/service: "*"
conf:
    host: "{{.svc}}"
    port: "8080"
    parameters:
    - name: "svc"
      tagKey: "kuma.io/service"
`
		err = YamlUniversal(virtualOutboundAll)(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(5 * time.Second)

		// Check we can reach the first instance
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.1:8080",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("srv-1"))
		}).Should(Succeed())

		// Check we can reach the second instance
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.2:8080",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("srv-2"))
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server:8080",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Or(Equal("srv-2"), Equal("srv-1")))
		}).Should(Succeed())
	})
}
