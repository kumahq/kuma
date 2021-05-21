package universal_multizone

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
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

		echoServerToken, err := globalCP.GenerateDpToken(defaultMesh, "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		backendToken, err := globalCP.GenerateDpToken(defaultMesh, "backend")
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
			Install(EchoServerUniversal("dp-echo-1", defaultMesh, "echo-v1", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v1"),
			)).
			Install(EchoServerUniversal("dp-echo-2", defaultMesh, "echo-v2", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v2"),
			)).
			Install(EchoServerUniversal("dp-echo-3", defaultMesh, "echo-v3", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v3"),
			)).
			Install(EchoServerUniversal("dp-echo-4", defaultMesh, "echo-v4", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v4"),
			)).
			Install(EchoServerUniversal("dp-backend-1", defaultMesh, "backend-v1", backendToken,
				WithServiceName("backend"),
				WithServiceVersion("v1"),
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

	collectResponses := func(url string, expected, notExpected []string) {
		instances := map[string]bool{}
		Eventually(func() bool {
			stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", url)
			if err != nil {
				return false
			}

			for _, ne := range notExpected {
				Expect(stdout).ToNot(ContainSubstring(ne))
			}

			for _, e := range expected {
				if strings.Contains(stdout, e) {
					instances[e] = true
				}
			}

			return len(instances) == len(expected)
		}, "30s", "500ms").Should(BeTrue())
	}

	within := func(n float64, l, r int) bool {
		return float64(l) <= n && n <= float64(r)
	}

	waitRatio := func(url string, v1Weight, v2Weight int, echo1, echo2 string) error {
		minAttempts := 100
		maxAttempts := 1000
		errorPercent := 3

		Eventually(func() bool {
			_, _, err := remote_1.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", url)
			return err == nil
		}, "30s", "500ms").Should(BeTrue())

		echo1Resp := float64(0)
		echo2Resp := float64(0)
		for i := 0; i < maxAttempts; i++ {
			stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
				"curl", "-m", "3", "--fail", url)
			Expect(err).ToNot(HaveOccurred())

			if strings.Contains(stdout, echo1) {
				echo1Resp++
			}
			if strings.Contains(stdout, echo2) {
				echo2Resp++
			}

			if i < minAttempts {
				continue
			}

			currentPercent1 := echo1Resp / float64(i+1) * 100
			currentPercent2 := echo2Resp / float64(i+1) * 100

			if within(currentPercent1, v1Weight-errorPercent, v1Weight+errorPercent) && within(currentPercent2, v2Weight-errorPercent, v2Weight+errorPercent) {
				return nil
			}
		}
		return errors.Errorf("resulted split values %v and %v aren't within original values %v and %v",
			echo1Resp/float64(maxAttempts)*100, echo2Resp/float64(maxAttempts)*100, v1Weight, v2Weight)
	}

	It("should access all instances of the service", func() {
		const trafficRoute = `
type: TrafficRoute
name: route-dc-to-echo
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
		Expect(NewClusterSetup().Install(YamlUniversal(trafficRoute)).Setup(global)).To(Succeed())

		collectResponses("echo-server_kuma-test_svc_8080.mesh", []string{"echo-v1", "echo-v2", "echo-v4"}, []string{"echo-v3"})
	})

	It("should route 100 percent of the traffic to the different service", func() {
		const trafficRoute = `
type: TrafficRoute
name: route-dc-to-echo
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
		Expect(NewClusterSetup().Install(YamlUniversal(trafficRoute)).Setup(global)).To(Succeed())

		collectResponses("backend.mesh", []string{"backend-v1"}, nil)
		collectResponses("backend.mesh", []string{"backend-v1"}, []string{"echo-v1", "echo-v2", "echo-v3", "echo-v4"})
	})

	It("should route split traffic between the versions with 20/80 ratio", func() {
		v1Weight := 80
		v2Weight := 20

		trafficRoute := fmt.Sprintf(`
type: TrafficRoute
name: route-dc-to-echo
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
		Expect(NewClusterSetup().Install(YamlUniversal(trafficRoute)).Setup(global)).To(Succeed())

		err := waitRatio("echo-server_kuma-test_svc_8080.mesh", v1Weight, v2Weight, "echo-v1", "echo-v2")
		Expect(err).ToNot(HaveOccurred())
	})
}
