package meshfaultinjection

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/v3/test/framework"
	"github.com/kumahq/kuma/v3/test/framework/client"
	"github.com/kumahq/kuma/v3/test/framework/envs/universal"
)

func Policy() {
	meshName := "mesh-fault-injection"
	zoneName := universal.Cluster.ZoneName()
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
	meshIdentity := fmt.Sprintf(`
type: MeshIdentity
name: identity
mesh: "%s"
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, meshName)
	meshTrafficPermission := fmt.Sprintf(`
type: MeshTrafficPermission
name: allow-mesh
mesh: "%s"
spec:
  rules:
  - default:
      allow:
      - spiffeID:
          type: Prefix
          value: spiffe://%s.%s.mesh.local
`, meshName, meshName, zoneName)
	faultInjection := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton-402
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/service: test-server
  rules:
    - matches:
        - spiffeID:
            type: Exact
            value: spiffe://%s.%s.mesh.local/workload/demo-client-blocked
      default:
        http:
          - abort:
              httpStatus: 402
              percentage: "100.0"
    - matches:
        - spiffeID:
            type: Exact
            value: spiffe://%s.%s.mesh.local/workload/demo-client-timeout
      default:
        http:
          - delay:
              value: 5s
              percentage: "100.0"
`, meshName, meshName, zoneName, meshName, zoneName)
	faultInjectionAllSources := fmt.Sprintf(`
type: MeshFaultInjection
mesh: "%s"
name: mesh-fault-injecton-all
spec:
  targetRef:
    kind: Dataplane
    labels:
      kuma.io/service: test-service-block-all-sources
  rules:
    - default:
        http:
          - abort:
              httpStatus: 421
              percentage: 100
`, meshName)
	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(YamlUniversal(meshIdentity)).
			Install(YamlUniversal(meshTrafficPermission)).
			Install(YamlUniversal(faultInjection)).
			Install(YamlUniversal(faultInjectionAllSources)).
			Install(YamlUniversal(timeout)).
			Install(TestServerUniversal(
				"test-server", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"}),
				WithLabels(map[string]string{"kuma.io/service": "test-server"}),
			)).
			Install(TestServerUniversal(
				"test-server-block-all-sources", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"}),
				WithServiceName("test-service-block-all-sources"),
				WithLabels(map[string]string{"kuma.io/service": "test-service-block-all-sources"}),
			)).
			Install(DemoClientUniversal("demo-client", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-blocked", meshName, WithTransparentProxy(true))).
			Install(DemoClientUniversal("demo-client-timeout", meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
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

	DescribeTable(
		"should be affected by fault and return",
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
			address:              "test-server.svc.mesh.local",
			expectedResponseCode: 402,
		}),
		Entry("421 when requests from any client are blocked", testCase{
			client:               "demo-client",
			address:              "test-service-block-all-sources.svc.mesh.local",
			expectedResponseCode: 421,
		}),
		Entry("421 when requests from the demo-client-blocked", testCase{
			client:               "demo-client-timeout",
			address:              "test-service-block-all-sources.svc.mesh.local",
			expectedResponseCode: 421,
		}),
		Entry("421 when requests from the demo-client-blocked", testCase{
			client:               "demo-client-blocked",
			address:              "test-service-block-all-sources.svc.mesh.local",
			expectedResponseCode: 421,
		}),
	)

	It("should not be affected by any fault", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(
				universal.Cluster, "demo-client", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	It("should delay responses for demo-client-timeout", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectFailure(
				universal.Cluster, "demo-client-timeout", "test-server.svc.mesh.local",
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "30s", "1s").Should(Succeed())
	})
}
