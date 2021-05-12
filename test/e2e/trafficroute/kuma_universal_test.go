package trafficroute_test

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

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

const trafficRoute = `
type: TrafficRoute
name: route-dc-to-echo-v3
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
        version: v3
`

var _ = Describe("Test Multizone Universal deployment", func() {

	const defaultMesh = "default"

	var global, remote_1, remote_2 Cluster
	var optsGlobal, optsRemote1, optsRemote2 []DeployOptionsFunc

	BeforeEach(func() {
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
			Install(YamlUniversal(trafficRoute)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken(defaultMesh, "echo-server_kuma-test_svc_8080")
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
			Install(EchoServerUniversal("dp-echo-1", defaultMesh, "echo-universal-1", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v1"))).
			Install(EchoServerUniversal("dp-echo-2", defaultMesh, "echo-universal-2", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v2"))).
			Install(EchoServerUniversal("dp-echo-3", defaultMesh, "echo-universal-3", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v3"))).
			Install(IngressUniversal(defaultMesh, ingressToken)).
			Setup(remote_2)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := remote_1.DeleteKuma(optsRemote1...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remote_2.DeleteKuma(optsRemote2...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should access all instances of the service", func() {
		instances := map[string]bool{}
		Eventually(func() bool {
			stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			if err != nil {
				return false
			}

			switch {
			case strings.Contains(stdout, "echo-universal-1"):
				instances["echo-universal-1"] = true
			case strings.Contains(stdout, "echo-universal-2"):
				instances["echo-universal-2"] = true
			case strings.Contains(stdout, "echo-universal-3"):
				instances["echo-universal-3"] = true
			}
			if len(instances) < 3 {
				fmt.Printf("%v/3\n", len(instances))
			}
			return len(instances) == 3
		}, "30s", "500ms").Should(BeTrue())
	})
})

var _ = Describe("Test Standalone Universal deployment", func() {

	const defaultMesh = "default"

	var universal Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters([]string{Kuma3}, Verbose)
		Expect(err).ToNot(HaveOccurred())

		universal = clusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(core.Standalone)).
			Install(YamlUniversal(trafficRoute)).
			Setup(universal)
		Expect(err).ToNot(HaveOccurred())
		err = universal.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := universal.GetKuma().GenerateDpToken(defaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(EchoServerUniversal("dp-echo-1", defaultMesh, "echo-universal-1", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v1"))).
			Install(EchoServerUniversal("dp-echo-2", defaultMesh, "echo-universal-2", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v2"))).
			Install(EchoServerUniversal("dp-echo-3", defaultMesh, "echo-universal-3", echoServerToken,
				WithTransparentProxy(true),
				WithServiceVersion("v3"))).
			Install(DemoClientUniversal(AppModeDemoClient, defaultMesh, demoClientToken, WithTransparentProxy(true))).
			Setup(universal)
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := universal.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = universal.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should access all instances of the service", func() {

		instances := map[string]bool{}
		Eventually(func() bool {
			stdout, _, err := universal.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			if err != nil {
				return false
			}

			switch {
			case strings.Contains(stdout, "echo-universal-1"):
				instances["echo-universal-1"] = true
			case strings.Contains(stdout, "echo-universal-2"):
				instances["echo-universal-2"] = true
			case strings.Contains(stdout, "echo-universal-3"):
				instances["echo-universal-3"] = true
			}
			if len(instances) < 3 {
				fmt.Printf("%v/3\n", len(instances))
			}
			return len(instances) == 3
		}, "30s", "500ms").Should(BeTrue())
	})
})
