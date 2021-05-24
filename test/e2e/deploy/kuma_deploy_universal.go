package deploy

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/config/core"

	"github.com/go-errors/errors"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

func UniversalDeployment() {
	meshMTLSOn := func(mesh, localityAware string) string {
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

	meshMTLSOff := func(mesh string) string {
		return fmt.Sprintf(`
type: Mesh
name: %s
`, mesh)
	}

	const iterations = 100
	const defaultMesh = "default"
	const nonDefaultMesh = "non-default"

	var global, remote_1, remote_2 Cluster
	var optsGlobal, optsRemote1, optsRemote2 = KumaUniversalDeployOpts, KumaUniversalDeployOpts, KumaUniversalDeployOpts

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3, Kuma4, Kuma5},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma5)
		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Install(YamlUniversal(meshMTLSOn(nonDefaultMesh, "false"))).
			Install(YamlUniversal(meshMTLSOff(defaultMesh))).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken(nonDefaultMesh, "demo-client")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := globalCP.GenerateDpToken(defaultMesh, "ingress")
		Expect(err).ToNot(HaveOccurred())

		// TODO: right now these tests are deliberately run WithHDS(false)
		// even if HDS is enabled without any ServiceProbes it still affects
		// first 2-3 load balancer requests, it's fine but tests should be rewritten

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma3)
		optsRemote1 = append(optsRemote1,
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithHDS(false))

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote1...)).
			Install(EchoServerUniversal(AppModeEchoServer, nonDefaultMesh, "universal1", echoServerToken, WithTransparentProxy(true))).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(defaultMesh, ingressToken)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Cluster 2
		remote_2 = clusters.GetCluster(Kuma4)
		optsRemote2 = append(optsRemote2,
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithHDS(false))

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote2...)).
			Install(EchoServerUniversal(AppModeEchoServer, nonDefaultMesh, "universal2", echoServerToken, WithTransparentProxy(true))).
			Install(DemoClientUniversal(AppModeDemoClient, nonDefaultMesh, demoClientToken, WithTransparentProxy(true))).
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

	It("should access service locally and remotely", func() {
		retry.DoWithRetry(remote_1.GetTesting(), "curl local service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})

		retry.DoWithRetry(remote_2.GetTesting(), "curl remote service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err := remote_2.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})
	})

	It("should distribute requests cross zones", func() {
		// given services in zone1 and zone2 in a mesh with disabled Locality Aware Load Balancing

		// when executing requests from zone 1
		responses := 0
		for i := 0; i < iterations; i++ {
			stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			Expect(stdout).To(ContainSubstring("universal"))

			if strings.Contains(stdout, "universal1") {
				responses++
			}
		}

		// then some requests are routed to the same zone and some are not
		Expect(responses > iterations/8).To(BeTrue())
		Expect(responses < iterations*7/8).To(BeTrue())
	})

	It("should use locality aware load balancing", func() {
		// given services in zone1 and zone2 in a mesh with enabled Locality Aware Load Balancing
		err := YamlUniversal(meshMTLSOn(nonDefaultMesh, "true"))(global)
		Expect(err).ToNot(HaveOccurred())

		// when executing requests from zone 2
		responses := 0
		for i := 0; i < iterations; i++ {
			stdout, _, err := remote_2.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "echo-server_kuma-test_svc_8080.mesh")
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			Expect(stdout).To(ContainSubstring("universal"))

			if strings.Contains(stdout, "universal2") {
				responses++
			}
		}

		// then all the requests are routed to the same zone 2
		Expect(responses).To(Equal(iterations))
	})
}
