package meshhealthcheck

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/test/e2e_env/universal/env"
	. "github.com/kumahq/kuma/test/framework"
)

func MeshHealthCheckPanicThreshold() {
	const meshName = "mhc-panic"

	healthCheck := fmt.Sprintf(`
type: MeshHealthCheck
mesh: %s
name: hc-1
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        interval: 10s
        timeout: 2s
        unhealthyThreshold: 3
        healthyThreshold: 1
        healthyPanicThreshold: 61
        failTrafficOnPanic: true
        tcp: {}`, meshName)

	dp := func(idx int) string {
		return fmt.Sprintf(`
type: Dataplane
mesh: %s
name: dp-echo-%d
networking:
  address: 192.168.0.%d
  inbound:
  - port: 8080
    servicePort: 80
    tags:
      kuma.io/service: test-server
      kuma.io/protocol: http`, meshName, idx, idx)
	}

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(healthCheck)).
			Setup(env.Cluster)
		Expect(err).ToNot(HaveOccurred())

		for i := 1; i <= 6; i++ {
			dpName := fmt.Sprintf("dp-echo-%d", i)
			response := fmt.Sprintf("universal-%d", i)
			err = TestServerUniversal(dpName, meshName, WithArgs([]string{"echo", "--instance", response}))(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		}
		for i := 7; i <= 10; i++ {
			err := NewClusterSetup().Install(YamlUniversal(dp(i))).Setup(env.Cluster)
			Expect(err).ToNot(HaveOccurred())
		}

		err = DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))(env.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(env.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(env.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should switch to panic mode and dismiss all requests", func() {
		Eventually(func() bool {
			stdout, _, _ := env.Cluster.Exec("", "", "demo-client",
				"curl", "-v", "test-server.mesh")
			return strings.Contains(stdout, "no healthy upstream")
		}, "30s", "500ms").Should(BeTrue())
	})
}
