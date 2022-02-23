package timeout

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func TimeoutPolicyOnUniversal() {
	var universalCluster Cluster

	faultInjection := `
type: FaultInjection
mesh: default
name: fi1
sources:
   - match:
       kuma.io/service: demo-client
destinations:
   - match:
       kuma.io/service: test-server
       kuma.io/protocol: http
conf:
   delay:
     percentage: 100
     value: 5s
`

	timeout := `
type: Timeout
mesh: default
name: echo-service-timeouts
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: test-server
conf:
  connectTimeout: 10s
  http:
    requestTimeout: 2s
`

	BeforeEach(func() {
		universalCluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(YamlUniversal(faultInjection)).
			Setup(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := universalCluster.GetKuma().GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := universalCluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = TestServerUniversal("test-server", "default", echoServerToken, WithArgs([]string{"echo", "--instance", "universal-1"}))(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))(universalCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(universalCluster.DeleteKuma()).To(Succeed())
		Expect(universalCluster.DismissCluster()).To(Succeed())
	})

	It("should reset the connection by timeout", func() {
		// check echo-server is up and running
		stdout, _, err := universalCluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		start := time.Now()
		_, _, err = universalCluster.Exec("", "", "demo-client",
			"curl", "-v", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		elapsed := time.Since(start)
		Expect(elapsed > 5*time.Second).To(BeTrue())

		// when apply Timeout policy
		err = NewClusterSetup().Install(YamlUniversal(timeout)).Setup(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		// then
		Eventually(func() bool {
			stdout, _, _ := universalCluster.Exec("", "", "demo-client",
				"curl", "-v", "test-server.mesh")
			return strings.Contains(stdout, "upstream request timeout")
		}, "30s", "500ms").Should(BeTrue())
	})
}
