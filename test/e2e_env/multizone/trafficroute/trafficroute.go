package trafficroute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	. "github.com/kumahq/kuma/test/framework"
	. "github.com/kumahq/kuma/test/framework/client"
)

func meshMTLSOn(mesh, localityAware string) string {
	return fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  localityAwareLoadBalancing: %s
`, mesh, localityAware)
}

func TrafficRoute() {
	const meshName = "tr-test"

	BeforeAll(func() {
		// Global
		err := env.Global.Install(YamlUniversal(meshMTLSOn(meshName, "false")))
		E2EDeferCleanup(func() {
			out, _ := env.Global.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshes")
			println(out)
			Expect(env.Global.DeleteMesh(meshName)).To(Succeed())
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(WaitForMesh(meshName, env.Zones())).To(Succeed())

		// Universal Zone 1
		E2EDeferCleanup(func() {
			Expect(env.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		})
		err = NewClusterSetup().
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true), WithConcurrency(8))).
			Install(TestServerUniversal("dp-echo-1", meshName,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceVersion("v1"),
			)).
			Setup(env.UniZone1)
		Expect(err).ToNot(HaveOccurred())

		// Universal Zone 2
		E2EDeferCleanup(func() {
			Expect(env.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		})
		err = NewClusterSetup().
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
			Setup(env.UniZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(env.Global, meshName, mesh.TrafficRouteResourceTypeDescriptor)).To(Succeed())
	})

	It("should access all instances of the service", func() {
		const trafficRoute = `
type: TrafficRoute
name: three-way-route
mesh: tr-test
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
		Expect(YamlUniversal(trafficRoute)(env.Global)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(3),
				HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), Not(BeNil())),
				HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), Not(BeNil())),
				Not(HaveKeyWithValue(MatchRegexp(`.*echo-v3.*`), Not(BeNil()))),
				HaveKeyWithValue(MatchRegexp(`.*echo-v4.*`), Not(BeNil())),
			),
		)
	})

	It("should route 100 percent of the traffic to the different service", func() {
		const trafficRoute = `
type: TrafficRoute
name: route-echo-to-backend
mesh: tr-test
sources:
  - match:
      kuma.io/service: demo-client
destinations:
  - match:
      kuma.io/service: test-server
conf:
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: another-test-server
`
		Expect(YamlUniversal(trafficRoute)(env.Global)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKeyWithValue(Equal(`another-test-server`), Not(BeNil())),
			),
		)
	})

	It("should route split traffic between the versions with 20/80 ratio", func() {
		v1Weight := 80
		v2Weight := 20

		trafficRoute := fmt.Sprintf(`
type: TrafficRoute
name: route-20-80-split
mesh: tr-test
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
		Expect(YamlUniversal(trafficRoute)(env.Global)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh", WithNumberOfRequests(100))
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", v1Weight, 10)),
				HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", v2Weight, 10)),
			),
		)
	})

	Context("HTTP routing", func() {
		HaveOnlyResponseFrom := func(response string) types.GomegaMatcher {
			return And(
				HaveLen(1),
				HaveKeyWithValue(MatchRegexp(`.*`+response+`.*`), Not(BeNil())),
			)
		}

		It("should route matching by path", func() {
			const trafficRoute = `
type: TrafficRoute
name: route-by-path
mesh: tr-test
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
			Expect(YamlUniversal(trafficRoute)(env.Global)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh/version1")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh/version2")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh/version3")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v4"))
		})

		It("should same splits with a different weights", func() {
			const trafficRoute = `
type: TrafficRoute
name: two-splits
mesh: tr-test
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
			Expect(YamlUniversal(trafficRoute)(env.Global)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh/split", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 2, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 8, 1)),
				),
			)
		})
	}, Ordered)

	Context("locality aware loadbalancing", func() {
		BeforeEach(func() {
			const trafficRoute = `
type: TrafficRoute
name: route-all-tr-test
mesh: tr-test
sources:
  - match:
      kuma.io/service: '*'
destinations:
  - match:
      kuma.io/service: '*'
conf:
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: '*'
`
			Expect(YamlUniversal(trafficRoute)(env.Global)).To(Succeed())
		})

		It("should loadbalance all requests equally by default", func() {
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh/split", WithNumberOfRequests(40))
			}, "30s", "500ms").Should(
				And(
					HaveLen(4),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), Not(BeNil())),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), Not(BeNil())),
					HaveKeyWithValue(MatchRegexp(`.*echo-v3.*`), Not(BeNil())),
					HaveKeyWithValue(MatchRegexp(`.*echo-v4.*`), Not(BeNil())),
					// todo(jakubdyszkiewicz) uncomment when https://github.com/kumahq/kuma/issues/2563 is fixed
					// HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 10, 1)),
					// HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 10, 1)),
					// HaveKeyWithValue(MatchRegexp(`.*echo-v3.*`), BeNumerically("~", 10, 1)),
					// HaveKeyWithValue(MatchRegexp(`.*echo-v4.*`), BeNumerically("~", 10, 1)),
				),
			)
		})

		It("should keep the request in the zone when locality aware loadbalancing is enabled", func() {
			// given
			Expect(YamlUniversal(meshMTLSOn(meshName, "true"))(env.Global)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(env.UniZone1, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(
				And(
					HaveLen(1),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), Not(BeNil())),
				),
			)
		})
	}, Ordered)
}
