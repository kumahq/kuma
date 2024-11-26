package meshretry

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func HttpRetry() {
	meshName := "meshretry-http"
	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal"}))).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Eventually(func() error {
			return universal.Cluster.GetKumactlOptions().RunKumactl("delete", "meshretry", "--mesh", meshName, "mesh-retry-all-"+meshName)
		}).Should(Succeed())
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.GetKumactlOptions().RunKumactl("delete", "dataplane", "fake-echo-server", "-m", meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should retry on HTTP connection failure", func() {
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
		meshRetryPolicy := fmt.Sprintf(`
type: MeshRetry
mesh: "%s"
name: fake-meshretry-policy
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        http:
          numRetries: 5
`, meshName)

		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
<<<<<<< HEAD
		}).Should(Succeed())
		Consistently(func(g Gomega) {
			// --max-time 8 to wait for 8 seconds to beat the default 5s connect timeout
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(8),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
=======
		}, "30s", "100ms", MustPassRepeatedly(5)).Should(Succeed())
>>>>>>> fa008d158 (test(e2e): increase eventually time (#12099))

		By("Adding a faulty dataplane")
		Expect(universal.Cluster.Install(YamlUniversal(echoServerDataplane))).To(Succeed())

		// Increased the time to 30 seconds
		// reference: https://github.com/kumahq/kuma/issues/12098
		// The default `initial_fetch_timeout` is 15 seconds.
		// In case the race condition described in the issue occurs,
		// this provides enough time to validate whether the change has arrived.
		By("Check some errors happen")
		var errs []error
		for i := 0; i < 50; i++ {
			time.Sleep(time.Millisecond * 100)
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(8),
			)
<<<<<<< HEAD
			if err != nil {
				errs = append(errs, err)
			}
		}
		Expect(errs).ToNot(BeEmpty())
=======

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(500))
		}, "30s", "100ms").Should(Succeed())

		By("Apply a MeshRetry policy")
		Expect(universal.Cluster.Install(YamlUniversal(meshRetryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
	})

	It("should retry on HTTP connection failure with real MeshService", func() {
		meshFaultInjection := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton
spec:
  targetRef:
    kind: MeshService
    name: test-server
  from:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 500
              percentage: "50.0"
`, meshName)
		meshRetryPolicy := fmt.Sprintf(`
type: MeshRetry
mesh: "%s"
name: meshretry-policy
spec:
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        http:
          numRetries: 5
          retryOn:
            - "5xx"
`, meshName)

		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.universal.ms",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "100ms", MustPassRepeatedly(5)).Should(Succeed())

		By("Adding a MeshFaultInjection for test-server")
		Expect(universal.Cluster.Install(YamlUniversal(meshFaultInjection))).To(Succeed())

		By("Check some errors happen")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.universal.ms",
				client.NoFail(),
				client.OutputFormat(`{ "received": { "status": %{response_code} } }`),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(500))
		}, "30s", "100ms").Should(Succeed())

		By("Apply a MeshRetry policy")
		Expect(universal.Cluster.Install(YamlUniversal(meshRetryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.universal.ms",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
	})

	It("should retry on HTTP connection failure applied on MeshHTTPRoute", func() {
		meshFaultInjection := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton
spec:
  targetRef:
    kind: MeshService
    name: test-server
  from:
    - targetRef:
        kind: Mesh
      default:
        http:
          - abort:
              httpStatus: 500
              percentage: "50.0"
`, meshName)
		meshRetryPolicy := fmt.Sprintf(`
type: MeshRetry
mesh: "%s"
name: meshretry-policy
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: http-route-1
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          numRetries: 5
          retryOn:
            - "5xx"
`, meshName)
		meshHttpRoute := fmt.Sprintf(`
type: MeshHTTPRoute
mesh: %s
name: http-route-1
spec:
  targetRef:
    kind: MeshService
    name: demo-client
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      rules:
        - matches:
            - path:
                value: /
                type: PathPrefix
          default:
            backendRefs:
              - kind: MeshService
                name: test-server
                weight: 100`, meshName)

		Expect(universal.Cluster.Install(YamlUniversal(meshHttpRoute))).To(Succeed())

		By("Checking requests succeed")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "100ms", MustPassRepeatedly(5)).Should(Succeed())

		By("Adding a MeshFaultInjection for test-server")
		Expect(universal.Cluster.Install(YamlUniversal(meshFaultInjection))).To(Succeed())

		By("Check some errors happen")
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.NoFail(),
				client.OutputFormat(`{ "received": { "status": %{response_code} } }`),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(500))
		}, "30s", "100ms").Should(Succeed())
>>>>>>> fa008d158 (test(e2e): increase eventually time (#12099))

		By("Apply a MeshRetry policy")
		Expect(universal.Cluster.Install(YamlUniversal(meshRetryPolicy))).To(Succeed())

		By("Eventually all requests succeed consistently")
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.mesh",
				client.WithMaxTime(8),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "1m", "1s", MustPassRepeatedly(5)).Should(Succeed())
	})
}
