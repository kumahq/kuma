package trafficroute

import (
	"fmt"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func TrafficRoute() {
	meshName := "trafficroute"
	var esHttpHostPort string

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

		esHttpHostPort = net.JoinHostPort(universal.Cluster.GetApp("route-es-http").GetContainerName(), "80")
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
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
		Expect(DeleteMeshResources(universal.Cluster, meshName, mesh.ExternalServiceResourceTypeDescriptor)).To(Succeed())
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
        version: v4
`
		Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(3),
				HaveKey(Equal(`echo-v1`)),
				HaveKey(Equal(`echo-v2`)),
				HaveKey(Equal(`echo-v4`)),
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

	It("should route split traffic between the versions with 20/80 ratio", func() {
		v1Weight := 8
		v2Weight := 2

		trafficRoute := fmt.Sprintf(`
type: TrafficRoute
name: route-20-80-split
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
    - weight: %d
      destination:
        kuma.io/service: test-server
        version: v1
    - weight: %d
      destination:
        kuma.io/service: test-server
        version: v2
`, v1Weight, v2Weight)

		Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(10))
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKeyWithValue(Equal(`echo-v1`), BeNumerically("~", v1Weight, 1)),
				HaveKeyWithValue(Equal(`echo-v2`), BeNumerically("~", v2Weight, 1)),
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
  split:
    - weight: 50
      destination:
        kuma.io/service: test-server
        version: v1
    - weight: 50
      destination:
        kuma.io/service: route-es-http
`))).To(Succeed())
		Expect(universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: ExternalService
name: route-es-http-1
mesh: trafficroute
networking:
  address: %s
tags:
  kuma.io/service: route-es-http
  kuma.io/protocol: http
`, esHttpHostPort)))).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKey(Equal(`echo-v1`)),
				HaveKey(Equal(`route-es-http`)),
			),
		)
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

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/version1")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/version2")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/version3")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v4"))
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

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithHeader("x-version", "v1"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithHeader("x-version", "v2"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithHeader("x-version", "v3"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v4"))
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
  - weight: 50
    destination:
      kuma.io/service: test-server
      version: v3
  - weight: 50
    destination:
      kuma.io/service: test-server
      version: v4
`
			Expect(YamlUniversal(trafficRoute)(universal.Cluster)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/split", client.WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v3.*`), BeNumerically("~", 5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v4.*`), BeNumerically("~", 5, 1)),
				),
			)
		})

		It("should same splits with a different weights", func() {
			const trafficRoute = `
type: TrafficRoute
name: same-splits
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
      version: v1
  - weight: 80
    destination:
      kuma.io/service: test-server
      version: v2
`
			Expect(universal.Cluster.Install(YamlUniversal(trafficRoute))).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh/split", client.WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 2, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 8, 1)),
				),
			)
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

			Eventually(func() error {
				resp, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/test-rewrite-prefix")
				if err != nil {
					return err
				}
				if resp.Received.Path != "/new-rewrite-prefix" {
					return errors.Errorf("expected %s, got %s", "/new-rewrite-prefix", resp.Received.Path)
				}
				return nil
			}, "30s", "500ms").Should(Succeed())

			Eventually(func() error {
				resp, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/test-regex")
				if err != nil {
					return err
				}
				if resp.Received.Path != "/regex-test" {
					return errors.Errorf("expected %s, got %s", "/regex-test", resp.Received.Path)
				}
				return nil
			}, "30s", "500ms").Should(Succeed())
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

			Eventually(func() error {
				resp, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/modified-host")
				if err != nil {
					return err
				}
				host := resp.Received.Headers["Host"]
				if len(host) < 1 || host[0] != "modified-host" {
					return errors.Errorf("expected %s, got %s", "modified-host", host)
				}
				return nil
			}, "30s", "500ms").Should(Succeed())

			Eventually(func() error {
				resp, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/from-path")
				if err != nil {
					return err
				}
				host := resp.Received.Headers["Host"]
				if len(host) < 1 || host[0] != "path" {
					return errors.Errorf("expected %s, got %s", "path", host)
				}
				return nil
			}, "30s", "500ms").Should(Succeed())
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

			Eventually(func() error {
				resp, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/modified-headers",
					client.WithHeader("header-to-remove", "abc"),
					client.WithHeader("x-multiple-values", "abc"),
				)
				if err != nil {
					return err
				}
				header := resp.Received.Headers["X-Custom-Header"]
				if len(header) < 1 || header[0] != "xyz" {
					return errors.Errorf("expected %s, got %s", "xyz", header)
				}
				if len(resp.Received.Headers["Header-To-Remove"]) > 0 {
					return errors.New("expected 'Header-To-Remove' to not be present")
				}
				header = resp.Received.Headers["X-Multiple-Values"]
				if len(header) < 2 || header[0] != "abc" || header[1] != "xyz" {
					return errors.Errorf("expected %s, got %s", "abc,xyz", header)
				}
				return nil
			}, "30s", "500ms").Should(Succeed())

			// "add" should replace existing headers
			Eventually(func() error {
				resp, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.mesh/modified-headers", client.WithHeader("x-custom-header", "abc"))
				if err != nil {
					return err
				}
				header := resp.Received.Headers["X-Custom-Header"]
				if len(header) < 1 || header[0] != "xyz" {
					return errors.Errorf("expected %s, got %s", "xyz", header)
				}
				return nil
			}, "30s", "500ms").Should(Succeed())
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
  - match:
      path:
        prefix: /
    split:
    - weight: 50
      destination:
        kuma.io/service: test-server
        version: v1
    - weight: 50
      destination:
        kuma.io/service: route-es-http
  destination:
    kuma.io/service: test-server
`)(universal.Cluster)).To(Succeed())
			Expect(universal.Cluster.Install(YamlUniversal(fmt.Sprintf(`
type: ExternalService
name: route-es-http-1
mesh: trafficroute
networking:
  address: %s
tags:
  kuma.io/service: route-es-http
  kuma.io/protocol: http
`, esHttpHostPort)))).To(Succeed())

<<<<<<< HEAD
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKey(Equal(`echo-v1`)),
					HaveKey(Equal(`route-es-http`)),
				),
			)
=======
			// Check and retry until the config got propagated to the client
			Eventually(func() ([]server_types.EchoResponse, error) {
				return client.CollectResponses(universal.Cluster, "demo-client", "test-server.mesh/i-am-here")
			}, "1m", "500ms").MustPassRepeatedly(5).Should(HaveEach(HaveField("Received.Headers", HaveKeyWithValue("X-I-Am-Here", []string{"route-internal-external"}))))

			Expect(client.CollectResponsesByInstance(universal.Cluster, "demo-client", "test-server.mesh", client.WithNumberOfRequests(100))).
				Should(And(
					HaveKeyWithValue("echo-v1", BeNumerically("~", 50, 25)),
					HaveKeyWithValue("route-es-http", BeNumerically("~", 50, 25)),
				))
>>>>>>> 32f87a14f (test(lb): loosen up assertion lb weights (#11720))
		})
	})
}
