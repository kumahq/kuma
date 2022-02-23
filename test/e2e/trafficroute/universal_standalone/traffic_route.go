package universal_standalone

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/client"
)

const defaultMesh = "default"

var universal Cluster

var _ = E2EBeforeSuite(func() {
	clusters, err := NewUniversalClusters([]string{Kuma3}, Verbose)
	Expect(err).ToNot(HaveOccurred())

	universal = clusters.GetCluster(Kuma3)

	Expect(NewClusterSetup().
		Install(Kuma(core.Standalone)).
		Setup(universal)).To(Succeed())

	testServerToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "test-server")
	Expect(err).ToNot(HaveOccurred())
	anotherTestServerToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "another-test-server")
	Expect(err).ToNot(HaveOccurred())
	demoClientToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "demo-client")
	Expect(err).ToNot(HaveOccurred())

	Expect(NewClusterSetup().
		Install(TestServerUniversal("dp-echo-1", defaultMesh, testServerToken,
			WithArgs([]string{"echo", "--instance", "echo-v1"}),
			WithServiceVersion("v1"),
		)).
		Install(TestServerUniversal("dp-echo-2", defaultMesh, testServerToken,
			WithArgs([]string{"echo", "--instance", "echo-v2"}),
			WithServiceVersion("v2"),
		)).
		Install(TestServerUniversal("dp-echo-3", defaultMesh, testServerToken,
			WithArgs([]string{"echo", "--instance", "echo-v3"}),
			WithServiceVersion("v3"),
		)).
		Install(TestServerUniversal("dp-echo-4", defaultMesh, testServerToken,
			WithArgs([]string{"echo", "--instance", "echo-v4"}),
			WithServiceVersion("v4"),
		)).
		Install(TestServerUniversal("dp-another-test", defaultMesh, anotherTestServerToken,
			WithArgs([]string{"echo", "--instance", "another-test-server"}),
			WithServiceName("another-test-server"),
		)).
		Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, demoClientToken, WithTransparentProxy(true))).
		Setup(universal)).To(Succeed())

	E2EDeferCleanup(universal.DismissCluster)
})

func KumaStandalone() {
	E2EAfterEach(func() {
		// remove all TrafficRoutes
		items, err := universal.GetKumactlOptions().KumactlList("traffic-routes", "default")
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			if item == "route-all-default" {
				continue
			}
			err := universal.GetKumactlOptions().KumactlDelete("traffic-route", item, "default")
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("should access all instances of the service", func() {
		const trafficRoute = `
type: TrafficRoute
name: three-way-route
mesh: default
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
		Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh")
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
mesh: default
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
		Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh")
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
mesh: default
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
		Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh", WithNumberOfRequests(10))
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKeyWithValue(Equal(`echo-v1`), BeNumerically("~", v1Weight, 1)),
				HaveKeyWithValue(Equal(`echo-v2`), BeNumerically("~", v2Weight, 1)),
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
mesh: default
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
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh/version1")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh/version2")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh/version3")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v4"))
		})

		It("should route matching by header", func() {
			const trafficRoute = `
type: TrafficRoute
name: route-by-header
mesh: default
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
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh", WithHeader("x-version", "v1"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh", WithHeader("x-version", "v2"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh", WithHeader("x-version", "v3"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v4"))
		})

		It("should split by header and split by default", func() {
			const trafficRoute = `
type: TrafficRoute
name: two-splits
mesh: default
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
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh/split", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh", WithNumberOfRequests(10))
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
mesh: default
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
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh/split", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(universal, "demo-client", "test-server.mesh", WithNumberOfRequests(10))
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
mesh: default
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
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() error {
				resp, err := CollectResponse(universal, "demo-client", "test-server.mesh/test-rewrite-prefix")
				if err != nil {
					return err
				}
				if resp.Received.Path != "/new-rewrite-prefix" {
					return errors.Errorf("expected %s, got %s", "/new-rewrite-prefix", resp.Received.Path)
				}
				return nil
			}, "30s", "500ms").Should(Succeed())

			Eventually(func() error {
				resp, err := CollectResponse(universal, "demo-client", "test-server.mesh/test-regex")
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
mesh: default
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
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() error {
				resp, err := CollectResponse(universal, "demo-client", "test-server.mesh/modified-host")
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
				resp, err := CollectResponse(universal, "demo-client", "test-server.mesh/from-path")
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
mesh: default
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
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() error {
				resp, err := CollectResponse(universal, "demo-client", "test-server.mesh/modified-headers",
					WithHeader("header-to-remove", "abc"),
					WithHeader("x-multiple-values", "abc"),
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
				resp, err := CollectResponse(universal, "demo-client", "test-server.mesh/modified-headers", WithHeader("x-custom-header", "abc"))
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
	})
}
