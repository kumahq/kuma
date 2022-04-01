package universal

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func HealthCheckPanicThreshold() {
	var universalCluster Cluster

	healthCheck := `
type: HealthCheck
name: hc-1
mesh: default
sources:
- match:
    kuma.io/service: '*'
destinations:
- match:
    kuma.io/service: test-server
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
      kuma.io/service: test-server
      kuma.io/protocol: http`, idx, idx)
	}

	BeforeEach(func() {
		universalCluster = NewUniversalCluster(NewTestingT(), Kuma3, Silent)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(YamlUniversal(healthCheck)).
			Setup(universalCluster)
		Expect(err).ToNot(HaveOccurred())

		for i := 1; i <= 6; i++ {
			dpName := fmt.Sprintf("dp-echo-%d", i)
			response := fmt.Sprintf("universal-%d", i)
			err = TestServerUniversal(dpName, "default", WithArgs([]string{"echo", "--instance", response}))(universalCluster)
			Expect(err).ToNot(HaveOccurred())
		}
		for i := 7; i <= 10; i++ {
			err := NewClusterSetup().Install(YamlUniversal(dp(i))).Setup(universalCluster)
			Expect(err).ToNot(HaveOccurred())
		}

		err = DemoClientUniversal(AppModeDemoClient, "default", WithTransparentProxy(true))(universalCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(universalCluster.DismissCluster()).To(Succeed())
	})

	It("should switch to panic mode and dismiss all requests", func() {
		Eventually(func() bool {
			stdout, _, _ := universalCluster.Exec("", "", "demo-client",
				"curl", "-v", "test-server.mesh")
			return strings.Contains(stdout, "no healthy upstream")
		}, "30s", "500ms").Should(BeTrue())
	})
}
