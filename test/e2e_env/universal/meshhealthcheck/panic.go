package meshhealthcheck

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
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
        http:
          path: "/"`, meshName)

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
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		for i := 1; i <= 6; i++ {
			dpName := fmt.Sprintf("dp-echo-%d", i)
			response := fmt.Sprintf("universal-%d", i)
			err = TestServerUniversal(dpName, meshName, WithArgs([]string{"echo", "--instance", response}))(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		}
		for i := 7; i <= 10; i++ {
			err := NewClusterSetup().Install(YamlUniversal(dp(i))).Setup(universal.Cluster)
			Expect(err).ToNot(HaveOccurred())
		}

		err = DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		for i := 7; i <= 10; i++ {
			Expect(universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", fmt.Sprintf("dp-echo-%d", i), "-m", meshName)).To(Succeed())
		}
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should switch to panic mode and dismiss all requests", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(503))
		}, "30s", "500ms").Should(Succeed())
	})
}
