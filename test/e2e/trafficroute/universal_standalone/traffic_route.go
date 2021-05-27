package universal_standalone

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/kumahq/kuma/test/e2e/trafficroute/testutil"
	. "github.com/kumahq/kuma/test/framework"
)

func KumaStandalone() {
	const defaultMesh = "default"

	var universal Cluster

	E2EBeforeSuite(func() {
		clusters, err := NewUniversalClusters([]string{Kuma3}, Verbose)
		Expect(err).ToNot(HaveOccurred())

		universal = clusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Setup(universal)
		Expect(err).ToNot(HaveOccurred())
		err = universal.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		backendToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "backend")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(EchoServerUniversal("dp-echo-1", defaultMesh, "echo-v1", echoServerToken,
				WithTransparentProxy(true),
				WithProtocol("http"),
				WithServiceVersion("v1"),
			)).
			Install(EchoServerUniversal("dp-echo-2", defaultMesh, "echo-v2", echoServerToken,
				WithTransparentProxy(true),
				WithProtocol("http"),
				WithServiceVersion("v2"),
			)).
			Install(EchoServerUniversal("dp-echo-3", defaultMesh, "echo-v3", echoServerToken,
				WithTransparentProxy(true),
				WithProtocol("http"),
				WithServiceVersion("v3"),
			)).
			Install(EchoServerUniversal("dp-echo-4", defaultMesh, "echo-v4", echoServerToken,
				WithTransparentProxy(true),
				WithProtocol("http"),
				WithServiceVersion("v4"),
			)).
			Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(EchoServerUniversal("dp-backend-1", defaultMesh, "backend-v1", backendToken,
				WithServiceName("backend"),
				WithServiceVersion("v1"),
				WithTransparentProxy(true),
			)).
			Setup(universal)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterSuite(func() {
		Expect(universal.DeleteKuma()).To(Succeed())
		Expect(universal.DismissCluster()).To(Succeed())
	})

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
      kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  loadBalancer:
    roundRobin: {}
  split:
    - weight: 1
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v1
    - weight: 1
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v2
    - weight: 1
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v4
`
		Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh")
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
      kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  loadBalancer:
    roundRobin: {}
  split:
    - weight: 100
      destination:
        kuma.io/service: backend
`
		Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh")
		}, "30s", "500ms").Should(
			And(
				HaveLen(1),
				HaveKeyWithValue(MatchRegexp(`.*backend-v1*`), Not(BeNil())),
				Not(HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), Not(BeNil()))),
				Not(HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), Not(BeNil()))),
				Not(HaveKeyWithValue(MatchRegexp(`.*echo-v3.*`), Not(BeNil()))),
				Not(HaveKeyWithValue(MatchRegexp(`.*echo-v4.*`), Not(BeNil()))),
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
      kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  loadBalancer:
    roundRobin: {}
  split:
    - weight: %d
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v1
    - weight: %d
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v2
`, v1Weight, v2Weight)
		Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

		Eventually(func() (map[string]int, error) {
			return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh", WithNumberOfRequests(10))
		}, "30s", "500ms").Should(
			And(
				HaveLen(2),
				HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), ApproximatelyEqual(v1Weight, 1)),
				HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), ApproximatelyEqual(v2Weight, 1)),
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
      kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  http:
  - match:
      path:
        prefix: /version1
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v1
  - match:
      path:
        exact: /version2
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v2
  - match:
      path:
        regex: "^/version3$"
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v3
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: echo-server_kuma-test_svc_8080
    version: v4
`
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh/version1")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh/version2")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh/version3")
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh")
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
      kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  http:
  - match:
      headers:
        x-version:
          prefix: v1
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v1
  - match:
      headers:
        x-version:
          exact: v2
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v2
  - match:
      headers:
        x-version:
          regex: "^v3$"
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v3
  loadBalancer:
    roundRobin: {}
  destination:
    kuma.io/service: echo-server_kuma-test_svc_8080
    version: v4
`
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh", WithHeader("x-version", "v1"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v1"))
			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh", WithHeader("x-version", "v2"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v2"))
			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh", WithHeader("x-version", "v3"))
			}, "30s", "500ms").Should(HaveOnlyResponseFrom("echo-v3"))
			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh")
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
      kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  http:
  - match:
      path:
        prefix: /split
    split:
    - weight: 50
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v1
    - weight: 50
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v2
  loadBalancer:
    roundRobin: {}
  split:
  - weight: 50
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v3
  - weight: 50
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v4
`
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh/split", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), ApproximatelyEqual(5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), ApproximatelyEqual(5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v3.*`), ApproximatelyEqual(5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v4.*`), ApproximatelyEqual(5, 1)),
				),
			)
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
      kuma.io/service: echo-server_kuma-test_svc_8080
conf:
  http:
  - match:
      path:
        prefix: /split
    split:
    - weight: 50
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v1
    - weight: 50
      destination:
        kuma.io/service: echo-server_kuma-test_svc_8080
        version: v2
  loadBalancer:
    roundRobin: {}
  split:
  - weight: 20
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v1
  - weight: 80
    destination:
      kuma.io/service: echo-server_kuma-test_svc_8080
      version: v2
`
			Expect(YamlUniversal(trafficRoute)(universal)).To(Succeed())

			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh/split", WithNumberOfRequests(10))
			}, "30s", "500ms").Should(
				And(
					HaveLen(2),
					HaveKeyWithValue(MatchRegexp(`.*echo-v1.*`), ApproximatelyEqual(5, 1)),
					HaveKeyWithValue(MatchRegexp(`.*echo-v2.*`), ApproximatelyEqual(5, 1)),
				),
			)

			Eventually(func() (map[string]int, error) {
				return CollectResponses(universal, "demo-client", "echo-server_kuma-test_svc_8080.mesh", WithNumberOfRequests(10))
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
