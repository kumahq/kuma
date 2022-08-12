package virtualoutbound

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func VirtualOutbound() {
	meshName := "virtual-outbound"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server1", meshName, WithArgs([]string{"echo", "--instance", "srv-1"}), WithServiceInstance("1"))).
			Install(TestServerUniversal("test-server2", meshName, WithArgs([]string{"echo", "--instance", "srv-2"}), WithServiceInstance("2"))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
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
		err := YamlUniversal(virtualOutboundInstance)(env.Cluster)
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
		err = YamlUniversal(virtualOutboundAll)(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(5 * time.Second)

		// Check we can reach the first instance
		stdout, stderr, err := env.Cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.1:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring(`"instance":"srv-1"`))

		// Check we can reach the second instance
		stdout, stderr, err = env.Cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.2:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring(`"instance":"srv-2"`))

		stdout, stderr, err = env.Cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(Or(ContainSubstring(`"instance":"srv-2"`), ContainSubstring(`"instance":"srv-1"`)))
	})
}
