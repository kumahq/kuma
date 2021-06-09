package universal_multizone

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/e2e/trafficroute/testutil"
	. "github.com/kumahq/kuma/test/framework"
)

func KumaMultizone() {
	var meshMTLSOn = func(mesh, localityAware string) string {
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

	const defaultMesh = "default"

	var global, remote_1, remote_2 Cluster
	var optsGlobal, optsRemote1, optsRemote2 []DeployOptionsFunc

	E2EBeforeSuite(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3, Kuma4, Kuma5},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma5)
		optsGlobal = []DeployOptionsFunc{}
		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Install(YamlUniversal(meshMTLSOn(defaultMesh, "false"))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		testServerToken, err := globalCP.GenerateDpToken(defaultMesh, "test-server")
		Expect(err).ToNot(HaveOccurred())
		anotherTestServerToken, err := globalCP.GenerateDpToken(defaultMesh, "another-test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken(defaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := globalCP.GenerateDpToken(defaultMesh, "ingress")
		Expect(err).ToNot(HaveOccurred())

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma3)
		optsRemote1 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote1...)).
			Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(defaultMesh, ingressToken)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Cluster 2
		remote_2 = clusters.GetCluster(Kuma4)
		optsRemote2 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote2...)).
			Install(TestServerUniversal("dp-echo-1", defaultMesh, testServerToken,
				WithArgs([]string{"echo", "--instance", "echo-v1"}),
				WithProtocol("http"),
				WithServiceVersion("v1"),
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("dp-echo-2", defaultMesh, testServerToken,
				WithArgs([]string{"echo", "--instance", "echo-v2"}),
				WithProtocol("http"),
				WithServiceVersion("v2"),
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("dp-echo-3", defaultMesh, testServerToken,
				WithArgs([]string{"echo", "--instance", "echo-v3"}),
				WithProtocol("http"),
				WithServiceVersion("v3"),
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("dp-echo-4", defaultMesh, testServerToken,
				WithArgs([]string{"echo", "--instance", "echo-v4"}),
				WithProtocol("http"),
				WithServiceVersion("v4"),
				WithTransparentProxy(true),
			)).
			Install(TestServerUniversal("dp-another-test", defaultMesh, anotherTestServerToken,
				WithArgs([]string{"echo", "--instance", "another-test-server"}),
				WithProtocol("http"),
				WithServiceName("another-test-server"),
				WithTransparentProxy(true),
			)).
			Install(IngressUniversal(defaultMesh, ingressToken)).
			Setup(remote_2)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		// remove all TrafficRoutes
		items, err := global.GetKumactlOptions().KumactlList("traffic-routes", "default")
		Expect(err).ToNot(HaveOccurred())
		for _, item := range items {
			if item == "route-all-default" {
				continue
			}
			err := global.GetKumactlOptions().KumactlDelete("traffic-route", item, "default")
			Expect(err).ToNot(HaveOccurred())
		}
	})

	E2EAfterSuite(func() {
		Expect(remote_1.DeleteKuma(optsRemote1...)).To(Succeed())
		Expect(remote_1.DismissCluster()).To(Succeed())

		Expect(remote_2.DeleteKuma(optsRemote2...)).To(Succeed())
		Expect(remote_2.DismissCluster()).To(Succeed())

		Expect(global.DeleteKuma(optsGlobal...)).To(Succeed())
		Expect(global.DismissCluster()).To(Succeed())
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
		Expect(YamlUniversal(trafficRoute)(global)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh")
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
  destination:
    kuma.io/service: another-test-server
`
		Expect(YamlUniversal(trafficRoute)(global)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh")
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
		Expect(YamlUniversal(trafficRoute)(global)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh", WithNumberOfRequests(100))
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), ApproximatelyEqual(v1Weight, 10)),
				HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), ApproximatelyEqual(v2Weight, 10)),
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
			Expect(YamlUniversal(trafficRoute)(global)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh/version1")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh/version2")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh/version3")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v4"))
		})

		It("should same splits with a different weights", func() {
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
  - weight: 20
    destination:
      kuma.io/service: test-server
      version: v1
  - weight: 80
    destination:
      kuma.io/service: test-server
      version: v2
`
			Expect(YamlUniversal(trafficRoute)(global)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh/split", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), ApproximatelyEqual(5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), ApproximatelyEqual(5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return CollectResponsesByInstance(remote_1, "demo-client", "test-server.mesh", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), ApproximatelyEqual(2, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), ApproximatelyEqual(8, 1)),
				),
			)
		})
	})
}
