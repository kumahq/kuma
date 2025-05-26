package retry

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func Policy() {
	meshName := "retry"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(ResourceUniversal(samples.MeshDefaultBuilder().WithName(meshName).WithSkipCreatingInitialPolicies([]string{"*"}).Build())).
			Install(DemoClientUniversal(
				"demo-client-retry", meshName, WithTransparentProxy(true), WithServiceName("demo-client-retry"))).
			Install(TestServerUniversal(
				"test-server-retry", meshName, WithArgs([]string{"echo", "--instance", "universal"}), WithServiceName("test-server-retry")),
			).
			Install(TrafficRouteUniversal(meshName)).
			Install(TrafficPermissionUniversal(meshName)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should retry on HTTP connection failure", func() {
		fiPolicy := fmt.Sprintf(`
type: FaultInjection
mesh: "%s"
name: fi-retry
sources:
   - match:
       kuma.io/service: demo-client-retry
destinations:
   - match:
       kuma.io/service: test-server-retry
       kuma.io/protocol: http
conf:
   abort:
     httpStatus: 500
     percentage: 50`, meshName)
		retryPolicy := fmt.Sprintf(`
type: Retry
mesh: "%s"
name: fake-retry-policy
sources:
- match:
    kuma.io/service: demo-client-retry
destinations:
- match:
    kuma.io/service: test-server-retry
conf:
  http:
    numRetries: 5
`, meshName)

		By("Checking requests succeed")
		// then
		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				universal.Cluster,
				"demo-client-retry",
				"test-server-retry.mesh",
				client.WithNumberOfRequests(100),
			)
		}, "30s", "5s").Should(And(
			HaveLen(100),
			WithTransform(client.CountResponseCodes(200), BeNumerically("==", 100)),
		))

		By("Adding a fault injection")
		Expect(universal.Cluster.Install(YamlUniversal(fiPolicy))).To(Succeed())

		By("Check some errors happen")
		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				universal.Cluster,
				"demo-client-retry",
				"test-server-retry.mesh",
				client.WithNumberOfRequests(100),
			)
		}, "30s", "5s").Should(And(
			HaveLen(100),
			WithTransform(client.CountResponseCodes(500), BeNumerically("~", 50, 15)),
			WithTransform(client.CountResponseCodes(200), BeNumerically("~", 50, 15)),
		))

		By("Apply a retry policy")
		Expect(universal.Cluster.Install(YamlUniversal(retryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		// then
		Eventually(func() ([]client.FailureResponse, error) {
			return client.CollectResponsesAndFailures(
				universal.Cluster,
				"demo-client-retry",
				"test-server-retry.mesh",
				client.WithNumberOfRequests(100),
			)
		}, "30s", "5s").Should(And(
			HaveLen(100),
			ContainElements(HaveField("ResponseCode", 200)),
		))
	})
}
