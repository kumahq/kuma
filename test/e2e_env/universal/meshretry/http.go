package meshretry

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshfault_api "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func HttpRetry() {
	meshName := "meshretry-http"

	uniServiceYAML := fmt.Sprintf(`
type: MeshService
name: test-server
mesh: %s
labels:
  kuma.io/origin: zone
  kuma.io/env: universal
spec:
  selector:
    dataplaneTags:
      kuma.io/service: test-server
  ports:
  - port: 80
    targetPort: 80
    appProtocol: http
`, meshName)

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerUniversal("test-server", meshName, WithArgs([]string{"echo", "--instance", "universal"}))).
			Install(YamlUniversal(uniServiceYAML)).
			Install(YamlUniversal(`
type: HostnameGenerator
name: uni-ms
spec:
  template: '{{ .DisplayName }}.universal.ms'
  selector:
    meshService:
      matchLabels:
        kuma.io/origin: zone
        kuma.io/env: universal
`)).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		// Delete the default meshretry policy
		Eventually(func() error {
			return universal.Cluster.GetKumactlOptions().RunKumactl("delete", "meshretry", "--mesh", meshName, "mesh-retry-all-"+meshName)
		}).Should(Succeed())
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(universal.Cluster, meshName,
			meshretry_api.MeshRetryResourceTypeDescriptor,
			meshfault_api.MeshFaultInjectionResourceTypeDescriptor,
			meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor,
		)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	It("should retry on HTTP connection failure", func() {
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
    kind: MeshService
    name: demo-client
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
				universal.Cluster, "demo-client", "test-server.mesh",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "10s", "100ms", MustPassRepeatedly(5)).Should(Succeed())

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
		}, "10s", "100ms").Should(Succeed())

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
  targetRef:
    kind: Mesh
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
		}, "10s", "100ms", MustPassRepeatedly(5)).Should(Succeed())

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
		}, "10s", "100ms").Should(Succeed())

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
		}, "10s", "100ms", MustPassRepeatedly(5)).Should(Succeed())

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
		}, "10s", "100ms").Should(Succeed())

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
}
