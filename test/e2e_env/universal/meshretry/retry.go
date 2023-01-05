package retry

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func Policy() {
	meshName := "meshretry"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal"}))).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Eventually(func() error {
			return env.Cluster.GetKumactlOptions().RunKumactl("delete", "meshretry", "--mesh", meshName, "meshretry-all-"+meshName)
		}).Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should retry on TCP connection failure", func() {
		echoServerDataplane := fmt.Sprintf(`
type: Dataplane
mesh: "%s"
name: fake-echo-server
networking:
  address:  241.0.0.1
  inbound:
  - port: 7777
    servicePort: 7777
    tags:
      kuma.io/service: test-server
      kuma.io/protocol: http
`, meshName)
		meshretryPolicy := fmt.Sprintf(`
type: MeshRetry
mesh: "%s"
name: fake-meshretry-policy
sources:
- match:
    kuma.io/service: demo-client
destinations:
- match:
    kuma.io/service: test-server
conf:
  http:
    numRetries: 5
`, meshName)

		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())
		Consistently(func(g Gomega) {
			// -m 8 to wait for 8 seconds to beat the default 5s connect timeout
			stdout, stderr, err := env.Cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(BeEmpty())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		})

		By("Adding a faulty dataplane")
		Expect(env.Cluster.Install(YamlUniversal(echoServerDataplane))).To(Succeed())

		By("Check some errors happen")
		var errs []error
		for i := 0; i < 50; i++ {
			time.Sleep(time.Millisecond * 100)
			_, _, err := env.Cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "test-server.mesh")

			if err != nil {
				errs = append(errs, err)
			}
		}
		Expect(errs).ToNot(BeEmpty())

		By("Apply a meshretry policy")
		Expect(env.Cluster.Install(YamlUniversal(meshretryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		Eventually(func(g Gomega) {
			stdout, _, err := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "-m", "8", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		}).Should(Succeed())
		Consistently(func(g Gomega) {
			// -m 8 to wait for 8 seconds to beat the default 5s connect timeout
			stdout, stderr, err := env.Cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "test-server.mesh")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stderr).To(BeEmpty())
			g.Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
		})
	})
}
