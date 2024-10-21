package trafficroute

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
	server_types "github.com/kumahq/kuma/test/server/types"
)

func TrafficRoute() {
	meshName := "trafficroute"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MeshUniversal(meshName)).
			Install(TrafficRouteUniversal(meshName)).
			Install(TestServerUniversal("dp-echo-1", meshName,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceVersion("v1"),
			)).
			Install(TestServerUniversal("dp-echo-2", meshName,
				WithArgs([]string{"echo", "--instance", "echo-v2"}),
				WithServiceVersion("v2"),
			)).
			Install(TestServerUniversal("dp-echo-3", meshName,
				WithArgs([]string{"echo", "--instance", "echo-v3"}),
				WithServiceVersion("v3"),
			)).
			Install(TestServerUniversal("dp-echo-4", meshName,
				WithArgs([]string{"echo", "--instance", "echo-v4"}),
				WithServiceVersion("v4"),
			)).
			Install(TestServerUniversal("dp-another-test", meshName,
				WithArgs([]string{"echo", "--instance", "another-test-server"}),
				WithServiceName("another-test-server"),
			)).
			Install(TrafficPermissionUniversal(meshName)).
			Install(TestServerExternalServiceUniversal("route-es-http", 80, false)).
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true))).
			Setup(universal.Cluster)).To(Succeed())

		err := universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: ExternalService
name: route-es-http-1
mesh: trafficroute
networking:
  address: %s
tags:
  kuma.io/service: route-es-http
  kuma.io/protocol: http
`, net.JoinHostPort(universal.Cluster.GetApp("route-es-http").GetContainerName(), "80"))))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteApp("route-es-http")).To(Succeed())
	})

	E2EAfterEach(func() {
		// remove all TrafficRoutes
		items, err := universal.Cluster.GetKumactlOptions().KumactlList("traffic-routes", meshName)
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			if item == "route-all-trafficroute" {
				continue
			}
			err := universal.Cluster.GetKumactlOptions().KumactlDelete("traffic-route", item, meshName)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("should access all instances of the service", func() {
		const trafficRoute = `
type: TrafficRoute
name: three-way-route
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  loadBalancer:
    roundRobin: {}
  split:
    - weight: 1
      destination:
        kuma.io/service: test-server
        version: v1
    - weight: 1
      destination:
        kuma.io/service: test-server
        version: v2
    - weight: 1
      destination:
        kuma.io/service: test-server
        version: v3
`
		Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(3),
				HaveKey(Equal(`echo-v1`)),
				HaveKey(Equal(`echo-v2`)),
				HaveKey(Equal(`echo-v3`)),
			),
		)
	})

	It("should route 100 percent of the traffic to the different service", func() {
		const trafficRoute = `
type: TrafficRoute
name: route-echo-to-backend
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  loadBalancer:
    roundRobin: {}
  split:
    - weight: 100
      destination:
        kuma.io/service: another-test-server
`
		Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKey(Equal("another-test-server")),
			),
		)
	})

	It("should split traffic between internal and external services", func() {
		Expect(universal.Cluster.Install(YamlUniversal(`
type: TrafficRoute
name: route-internal-external
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match: # this match is here to be able to check if the config was set
      path:
        prefix: /i-am-here
    modify:
      requestHeaders:
        add:
        - name: X-I-Am-Here
          value: 'route-internal-external'
    destination:
      kuma.io/service: test-server
  split:
    - weight: 50
      destination:
        kuma.io/service: test-server
        version: v1
    - weight: 50
      destination:
        kuma.io/service: route-es-http
`))).To(Succeed())
		// Check and retry until the config got propagated to the client
		Eventually(func() ([]server_types.EchoResponse, error) {
			return client.CollectResponses(universal.Cluster, "demo-client", "test-server.mesh/i-am-here")
		}, "1m", "500ms").MustPassRepeatedly(5).Should(HaveEach(HaveField("Received.Headers", HaveKeyWithValue("X-I-Am-Here", []string{"route-internal-external"}))))

		Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(50))).
			Should(And(
				HaveLen(2),
				HaveKey(Equal(`echo-v1`)),
				HaveKey(Equal(`route-es-http`)),
			))
	})

	Context("HTTP routing", func() {
		HaveOnlyResponseFrom := func(response string) types.GomegaMatcher {
			return And(
				HaveLen(1),
				HaveKey(Equal(response)),
			)
		}

		It("should route matching by path", func() {
			const trafficRoute = `
type: TrafficRoute
name: route-by-path
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match: # this match is here to be able to check if the config was set
      path:
        prefix: /i-am-here
    modify:
      requestHeaders:
        add:
        - name: X-I-Am-Here
          value: 'route-by-path'
    destination:
      kuma.io/service: test-server
  - match:
      path:
        prefix: /version1
    destination:
      kuma.io/service: test-server
      version: v1
  - match:
      path:
        exact: /version2
    destination:
      kuma.io/service: test-server
      version: v2
  - match:
      path:
        regex: "^/version3$"
    destination:
      kuma.io/service: test-server
      version: v3
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: test-server
    version: v4
`
			Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

			// Check and retry until the config got propagated to the client
			Eventually(func() ([]server_types.EchoResponse, error) {
				return client.CollectResponses(universal.Cluster, "demo-client", "test-server.mesh/i-am-here")
			}, "1m", "500ms").MustPassRepeatedly(5).Should(HaveEach(HaveField("Received.Headers", HaveKeyWithValue("X-I-Am-Here", []string{"route-by-path"}))))

			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/version1", client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v1"))
			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/version2", client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v2"))
			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/version3", client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v3"))
			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v4"))
		})

		It("should route matching by header", func() {
			const trafficRoute = `
type: TrafficRoute
name: route-by-header
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match: # this match is here to be able to check if the config was set
      path:
        prefix: /i-am-here
    modify:
      requestHeaders:
        add:
        - name: X-I-Am-Here
          value: 'route-by-header'
    destination:
      kuma.io/service: test-server
  - match:
      headers:
        x-version:
          prefix: v1
    destination:
      kuma.io/service: test-server
      version: v1
  - match:
      headers:
        x-version:
          exact: v2
    destination:
      kuma.io/service: test-server
      version: v2
  - match:
      headers:
        x-version:
          regex: "^v3$"
    destination:
      kuma.io/service: test-server
      version: v3
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: test-server
    version: v4
`
			Expect(YamlUniversal(trafficRoute)(universal.Cluster)).To(Succeed())
			// Check and retry until the config got propagated to the client
			Eventually(func() ([]server_types.EchoResponse, error) {
				return client.CollectResponses(universal.Cluster, "demo-client", "test-server.mesh/i-am-here")
			}, "1m", "500ms").MustPassRepeatedly(5).Should(HaveEach(HaveField("Received.Headers", HaveKeyWithValue("X-I-Am-Here", []string{"route-by-header"}))))

			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithHeader("x-version", "v1"), client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v1"))
			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithHeader("x-version", "v2"), client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v2"))
			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithHeader("x-version", "v3"), client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v3"))
			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithHeader("x-version", "v4"), client.WithNumberOfRequests(10))).
				Should(HaveOnlyResponseFrom("echo-v4"))
		})

		It("should split by header and split by default", func() {
			const trafficRoute = `
type: TrafficRoute
name: two-splits
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match: # this match is here to be able to check if the config was set
      path:
        prefix: /i-am-here
    modify:
      requestHeaders:
        add:
        - name: X-I-Am-Here
          value: 'two-splits'
    destination:
      kuma.io/service: test-server
  - match:
      path:
        prefix: /split
    split:
    - weight: 50
      destination:
        kuma.io/service: test-server
        version: v1
    - weight: 50
      destination:
        kuma.io/service: test-server
        version: v2
  loadBalancer:
    roundRobin: {}
  split:
  - weight: 20
    destination:
      kuma.io/service: test-server
      version: v3
  - weight: 80
    destination:
      kuma.io/service: test-server
      version: v4
`
			Expect(YamlUniversal(trafficRoute)(universal.Cluster)).To(Succeed())
			// Check and retry until the config got propagated to the client
			Eventually(func() ([]server_types.EchoResponse, error) {
				return client.CollectResponses(universal.Cluster, "demo-client", "test-server.mesh/i-am-here")
			}, "1m", "500ms").MustPassRepeatedly(5).Should(HaveEach(HaveField("Received.Headers", HaveKeyWithValue("X-I-Am-Here", []string{"two-splits"}))))

			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/split", client.WithNumberOfRequests(50))).
				Should(And(
					HaveLen(2),
					HaveKey(`echo-v1`),
					HaveKey(`echo-v2`),
				))

			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(100))).
				Should(And(
					HaveLen(2),
					HaveKey(`echo-v3`),
					HaveKeyWithValue(`echo-v4`, BeNumerically("~", 80, 20)),
				))
		})

		It("should modify path", func() {
			const trafficRoute = `
type: TrafficRoute
name: modify-path
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match:
      path:
        prefix: "/test-rewrite-prefix"
    modify:
      path:
        rewritePrefix: "/new-rewrite-prefix"
    destination:
      kuma.io/service: test-server
  - match:
      path:
        prefix: "/test-regex"
    modify:
      path:
        regex:
          pattern: "^/(.*)-(.*)$"
          substitution: "/\\2-\\1"
    destination:
      kuma.io/service: test-server
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: test-server
`
			Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/test-rewrite-prefix")).
					Should(HaveField("Received.Path", "/new-rewrite-prefix"))
			}, "30s", "500ms").Should(Succeed())

			Expect(client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/test-regex")).
				Should(HaveField("Received.Path", "/regex-test"))
		})

		It("should modify host", func() {
			const trafficRoute = `
type: TrafficRoute
name: modify-host
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match:
      path:
        prefix: "/modified-host"
    modify:
      host:
        value: "modified-host"
    destination:
      kuma.io/service: test-server
  - match:
      path:
        prefix: "/from-path"
    modify:
      host:
        fromPath:
          pattern: "^/from-(.*)$"
          substitution: "\\1"
    destination:
      kuma.io/service: test-server
    destination:
      kuma.io/service: test-server
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: test-server
`
			Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/modified-host")).
					Should(HaveField("Received.Headers", HaveKeyWithValue("Host", []string{"modified-host"})))
			}, "30s", "500ms").Should(Succeed())

			Expect(client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/from-path")).
				Should(HaveField("Received.Headers", HaveKeyWithValue("Host", []string{"path"})))
		})

		It("should modify headers", func() {
			const trafficRoute = `
type: TrafficRoute
name: modify-headers
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match:
      path:
        prefix: "/modified-headers"
    modify:
      requestHeaders:
        add:
        - name: x-custom-header
          value: xyz
        - name: x-multiple-values
          value: xyz
          append: true
        remove:
        - name: header-to-remove
    destination:
      kuma.io/service: test-server
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: test-server
`
			Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/modified-headers",
					client.WithHeader("header-to-remove", "abc"),
					client.WithHeader("x-multiple-values", "abc"),
				)).Should(
					HaveField("Received.Headers", And(
						HaveKeyWithValue("X-Custom-Header", []string{"xyz"}),
						Not(HaveKey("Header-To-Remove")),
						HaveKeyWithValue("X-Multiple-Values", []string{"abc", "xyz"}),
					)),
				)
			}, "30s", "500ms").Should(Succeed())

			// "add" should replace existing headers
			Expect(client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/modified-headers", client.WithHeader("x-custom-header", "abc"))).
				Should(HaveField("Received.Headers", HaveKeyWithValue("X-Custom-Header", []string{"xyz"})))
		})

		It("should split traffic between internal and external destinations", func() {
			Expect(YamlUniversal(`
type: TrafficRoute
name: route-internal-external
mesh: trafficroute
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  http:
  - match: # this match is here to be able to check if the config was set
      path:
        prefix: /i-am-here
    modify:
      requestHeaders:
        add:
        - name: X-I-Am-Here
          value: 'route-internal-external'
    destination:
      kuma.io/service: test-server
  - match:
      path:
        prefix: /
    split:
    - weight: 1
      destination:
        kuma.io/service: test-server
        version: v1
    - weight: 1
      destination:
        kuma.io/service: route-es-http
  destination:
    kuma.io/service: test-server
`)(universal.Cluster)).To(Succeed())

			// Check and retry until the config got propagated to the client
			Eventually(func() ([]server_types.EchoResponse, error) {
				return client.CollectResponses(universal.Cluster, "demo-client", "test-server.mesh/i-am-here")
			}, "1m", "500ms").MustPassRepeatedly(5).Should(HaveEach(HaveField("Received.Headers", HaveKeyWithValue("X-I-Am-Here", []string{"route-internal-external"}))))

			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(100))).
				Should(And(
					HaveKeyWithValue("echo-v1", BeNumerically("~", 50, 25)),
					HaveKeyWithValue("route-es-http", BeNumerically("~", 50, 25)),
				))
		})
	})
}
