package virtualoutbound

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func VirtualOutboundOnUniversal() {
	var cluster Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		_ = cluster.GetKumactlOptions().RunKumactl("delete", "virtual-outbound", "instances")
		_ = cluster.GetKumactlOptions().RunKumactl("delete", "virtual-outbound", "all")

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server1", "default", echoServerToken, WithArgs([]string{"echo", "--instance", "srv-1"}), WithServiceInstance("1"))).
			Install(TestServerUniversal("test-server2", "default", echoServerToken, WithArgs([]string{"echo", "--instance", "srv-2"}), WithServiceInstance("2"))).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := cluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should add hostnames for individual instances", func() {
		virtualOutboundInstance := `
type: VirtualOutbound
mesh: default
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
		err := YamlUniversal(virtualOutboundInstance)(cluster)
		Expect(err).ToNot(HaveOccurred())

		virtualOutboundAll := `
type: VirtualOutbound
mesh: default
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
		err = YamlUniversal(virtualOutboundAll)(cluster)
		Expect(err).ToNot(HaveOccurred())

		time.Sleep(5 * time.Second)

		// Check we can reach the first instance
		stdout, stderr, err := cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.1:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring(`"instance":"srv-1"`))

		// Check we can reach the second instance
		stdout, stderr, err = cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.2:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(ContainSubstring(`"instance":"srv-2"`))

		stdout, stderr, err = cluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server:8080")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(BeEmpty())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		Expect(stdout).To(Or(ContainSubstring(`"instance":"srv-2"`), ContainSubstring(`"instance":"srv-1"`)))
	})
}
