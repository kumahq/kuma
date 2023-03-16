package meshfaultinjection

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "mesh-fault-injection"
	timeout := fmt.Sprintf(`
type: MeshTimeout
mesh: "%s"
name: mesh-timeout
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        http:
          requestTimeout: 3s
`, meshName)
	faultInjection := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton-402
spec:
  targetRef:
    kind: MeshService
    name: test-server
  from:
    - targetRef:
        kind: MeshService
        name: demo-client-blocked
      default:
        http:
          - abort:
              httpStatus: 402
              percentage: "100.0"
    - targetRef:
        kind: MeshService
        name: demo-client-timeout
      default:
        http:
          - delay:
              value: 5s
              percentage: "100.0"
`, meshName)
	faultInjectionAllSources := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton-all
spec:
  targetRef:
    kind: MeshService
    name: test-service-block-all-sources
  from:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 421
              percentage: 100
`, meshName)
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(faultInjection)).
			Install(YamlUniversal(faultInjectionAllSources)).
			Install(YamlUniversal(timeout)).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(TestServerUniversal("test-server-block-all-sources", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"}),
				WithServiceName("test-service-block-all-sources"),
			)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-blocked", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-timeout", meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})
	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	type testCase struct {
		client               string
		address              string
		expectedResponseCode int
	}

	DescribeTable("should be affected by fault and return",
		func(given testCase) {
			Eventually(func(g Gomega) {
				response, err := client.CollectFailure(
					universal.Cluster, given.client, given.address,
				)
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(response.ResponseCode).To(Equal(given.expectedResponseCode))
			}, "30s", "1s").Should(Succeed())
		},
		Entry("402 when requests from the demo-client-blocked", testCase{
			client:               "demo-client-blocked",
			address:              "test-server.mesh",
			expectedResponseCode: 402,
		}),
		Entry("421 when requests from any client are blocked", testCase{
			client:               "demo-client",
			address:              "test-service-block-all-sources.mesh",
			expectedResponseCode: 421,
		}),
		Entry("421 when requests from the demo-client-blocked", testCase{
			client:               "demo-client-timeout",
			address:              "test-service-block-all-sources.mesh",
			expectedResponseCode: 421,
		}),
		Entry("421 when requests from the demo-client-blocked", testCase{
			client:               "demo-client-blocked",
			address:              "test-service-block-all-sources.mesh",
			expectedResponseCode: 421,
		}),
	)

	It("should not be affected by any fault", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	It("should delay responses for demo-client-timeout", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client-timeout", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "30s", "1s").Should(Succeed())
	})
}
