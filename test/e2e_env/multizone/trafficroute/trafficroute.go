package trafficroute

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func TrafficRoute() {
	const meshName = "tr-test"

	BeforeAll(func() {
		// Global
		Expect(multizone.Global.Install(MTLSMeshUniversal(meshName))).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(DemoClientUniversal(AppModeDemoClient, meshName, WithTransparentProxy(true), WithConcurrency(8))).
			Install(TestServerUniversal("dp-echo-1", meshName,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithServiceVersion("v1"),
			)).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())

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
			Setup(multizone.UniZone2)
		Expect(err).ToNot(HaveOccurred())
	})
	E2EAfterAll(func() {
		Expect(multizone.UniZone2.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, mesh.TrafficRouteResourceTypeDescriptor)).To(Succeed())
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
		Expect(YamlUniversal(trafficRoute)(multizone.Global)).To(Succeed())
		Expect(WaitForResource(mesh.TrafficRouteResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: "three-way-route"}, multizone.Zones()...)).Should(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			MatchAllKeys(Keys{
				"echo-v1": Not(BeNil()),
				"echo-v2": Not(BeNil()),
				"echo-v4": Not(BeNil()),
			}),
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
		Expect(YamlUniversal(trafficRoute)(multizone.Global)).To(Succeed())
		Expect(WaitForResource(mesh.TrafficRouteResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: "route-echo-to-backend"}, multizone.Zones()...)).Should(Succeed())

		Eventually(func() (map[string]int, error) {
			return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
		}, "30s", "500ms").Should(
			MatchAllKeys(Keys{
				"another-test-server": Not(BeNil()),
			}),
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
		Expect(YamlUniversal(trafficRoute)(multizone.Global)).To(Succeed())
		Expect(WaitForResource(mesh.TrafficRouteResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: "route-20-80-split"}, multizone.Zones()...)).Should(Succeed())

		Eventually(func(g Gomega) {
			res, err := client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh", client.WithNumberOfRequests(200))
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(res).To(MatchAllKeys(Keys{
				"echo-v1": BeNumerically("~", 2*v1Weight, 20),
				"echo-v2": BeNumerically("~", 2*v2Weight, 20),
			}))
		}, "1m", "5s").Should(Succeed())
	})

	Context("HTTP routing", func() {
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
			Expect(YamlUniversal(trafficRoute)(multizone.Global)).To(Succeed())
			Expect(WaitForResource(mesh.TrafficRouteResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: "route-by-path"}, multizone.Zones()...)).Should(Succeed())

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh/version1")
			}, "30s", "500ms").Should(MatchAllKeys(Keys{"echo-v1": Not(BeNil())}))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh/version2")
			}, "30s", "500ms").Should(MatchAllKeys(Keys{"echo-v2": Not(BeNil())}))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh/version3")
			}, "30s", "500ms").Should(MatchAllKeys(Keys{"echo-v3": Not(BeNil())}))
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(MatchAllKeys(Keys{"echo-v4": Not(BeNil())}))
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
			Expect(YamlUniversal(trafficRoute)(multizone.Global)).To(Succeed())
			Expect(WaitForResource(mesh.TrafficRouteResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: "two-splits"}, multizone.Zones()...)).Should(Succeed())

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh/split", client.WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				MatchAllKeys(Keys{
					"echo-v1": BeNumerically("~", 5, 1),
					"echo-v2": BeNumerically("~", 5, 1),
				}),
			)

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh", client.WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				MatchAllKeys(Keys{
					"echo-v1": BeNumerically("~", 2, 1),
					"echo-v2": BeNumerically("~", 8, 1),
				}),
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
			Expect(YamlUniversal(trafficRoute)(multizone.Global)).To(Succeed())
			Expect(WaitForResource(mesh.TrafficRouteResourceTypeDescriptor, model.ResourceKey{Mesh: meshName, Name: "route-all-tr-test"}, multizone.Zones()...)).Should(Succeed())
		})

		It("should loadbalance all requests equally by default", func() {
			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh/split", client.WithNumberOfRequests(40))
			}, "30s", "500ms").Should(
				MatchAllKeys(Keys{
					"echo-v1": Not(BeNil()),
					"echo-v2": Not(BeNil()),
					"echo-v3": Not(BeNil()),
					"echo-v4": Not(BeNil()),
					// todo(jakubdyszkiewicz) uncomment when https://github.com/kumahq/kuma/issues/2563 is fixed
					// HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), BeNumerically("~", 10, 1)),
					// HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), BeNumerically("~", 10, 1)),
					// HaveKeyWithValue(MatchRegexp(`.*echo-v3.*`), BeNumerically("~", 10, 1)),
					// HaveKeyWithValue(MatchRegexp(`.*echo-v4.*`), BeNumerically("~", 10, 1)),
				}),
			)
		})

		It("should keep the request in the zone when locality aware loadbalancing is enabled", func() {
			// given
			Expect(YamlUniversal(fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
routing:
  localityAwareLoadBalancing: true
`, meshName))(multizone.Global)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return client.CollectResponsesByInstance(multizone.UniZone1, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(
				MatchAllKeys(Keys{
					"echo-v1": Not(BeNil()),
				}),
			)
		})
	}, Ordered)
}
