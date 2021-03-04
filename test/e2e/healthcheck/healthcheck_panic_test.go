package healthcheck_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test HealthCheck panic threshold", func() {

	var universalCluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	healthCheck := `
type: HealthCheck
name: hc-1
mesh: default
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  interval: 10s
  timeout: 2s
  unhealthyThreshold: 3
  healthyThreshold: 1
  healthyPanicThreshold: 61
  failTrafficOnPanic: true
  tcp: {}`

	dp := func(idx int) string {
		return fmt.Sprintf(`
type: Dataplane
mesh: default
name: dp-echo-%d
networking:
  address: 192.168.0.%d
  inbound:
  - port: 8080
    servicePort: 80
    tags:
      kuma.io/service: echo-server_kuma-test_svc_8080
      kuma.io/protocol: http`, idx, idx)
	}

	BeforeEach(func() {
		universalCluster = NewUniversalCluster(NewTestingT(), Kuma1, Silent)
		deployOptsFuncs = []DeployOptionsFunc{}

		err := NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlUniversal(healthCheck)).
			Setup(universalCluster)
		Expect(err).ToNot(HaveOccurred())
		err = universalCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := universalCluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := universalCluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		for i := 1; i <= 6; i++ {
			dpName := fmt.Sprintf("dp-echo-%d", i)
			response := fmt.Sprintf("universal-%d", i)
			err = EchoServerUniversal(dpName, "default", response, echoServerToken)(universalCluster)
			Expect(err).ToNot(HaveOccurred())
		}
		for i := 7; i <= 10; i++ {
			err := NewClusterSetup().Install(YamlUniversal(dp(i))).Setup(universalCluster)
			Expect(err).ToNot(HaveOccurred())
		}

		err = DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))(universalCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		Expect(universalCluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		Expect(universalCluster.DismissCluster()).To(Succeed())
	})

	It("should switch to panic mode and dismiss all requests", func() {
		Eventually(func() bool {
			stdout, _, _ := universalCluster.Exec("", "", "demo-client",
				"curl", "-v", "echo-server_kuma-test_svc_8080.mesh")
			return strings.Contains(stdout, "no healthy upstream")
		}, "30s", "500ms").Should(BeTrue())
	})
})
