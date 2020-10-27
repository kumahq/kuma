package e2e_test

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

var _ = Describe("Test Universal deployment", func() {

	meshDefaulMtlsOn := `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`
	trafficPermissionAll := `
type: TrafficPermission
name: traffic-permission-all
mesh: default
sources:
- match:
   kuma.io/service: "*"
destinations:
- match:
   kuma.io/service: "*"
`
	var global, remote_1, remote_2 Cluster
	var optsGlobal, optsRemote1, optsRemote2 []DeployOptionsFunc
	var echoServerToken string

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2, Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1)
		optsGlobal = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		echoServerToken, err = globalCP.GenerateDpToken("echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken("demo-client")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := globalCP.GenerateDpToken("ingress")
		Expect(err).ToNot(HaveOccurred())

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma2)
		optsRemote1 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote1...)).
			Install(DemoClientUniversal(demoClientToken)).
			Install(IngressUniversal(ingressToken)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Cluster 2
		remote_2 = clusters.GetCluster(Kuma3)
		optsRemote2 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote2...)).
			Install(DemoClientUniversal(demoClientToken)).
			Install(IngressUniversal(ingressToken)).
			Setup(remote_2)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		remote_1CP := remote_1.GetKuma()
		remote_2CP := remote_2.GetKuma()

		err = global.GetKumactlOptions().KumactlApplyFromString(
			fmt.Sprintf(ZoneTemplateUniversal, Kuma2, remote_1CP.GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())

		err = global.GetKumactlOptions().KumactlApplyFromString(
			fmt.Sprintf(ZoneTemplateUniversal, Kuma3, remote_2CP.GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(meshDefaulMtlsOn)(global)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(trafficPermissionAll)(global)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := remote_1.DeleteKuma(optsRemote1...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.DeleteKuma(optsRemote2...)
		Expect(err).ToNot(HaveOccurred())
		err = global.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())

		err = remote_1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should access service locally and remotely", func() {
		err := EchoServerUniversal("universal1", echoServerToken)(remote_1)
		Expect(err).ToNot(HaveOccurred())

		retry.DoWithRetry(remote_1.GetTesting(), "curl local service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail", "localhost:4001")
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
					"curl", "-v", "-m", "3", "--fail", "localhost:4001")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})
	})

	const iterations = 50

	It("Should distribute requests cross zones", func() {
		err := EchoServerUniversal("universal1", echoServerToken)(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = EchoServerUniversal("universal2", echoServerToken)(remote_2)
		Expect(err).ToNot(HaveOccurred())

		responses := 0
		for i := 0; i < iterations; i++ {
			stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "localhost:4001")
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			Expect(stdout).To(ContainSubstring("universal"))

			if strings.Contains(stdout, "universal1") {
				responses++
			}
		}

		Expect(responses > 10).To(BeTrue())
		Expect(responses < iterations-10).To(BeTrue())
	})

	It("Should use locality aware load balancing", func() {
		err := EchoServerUniversal("universal1", echoServerToken, WithLocality(true))(remote_1)
		Expect(err).ToNot(HaveOccurred())

		err = EchoServerUniversal("universal2", echoServerToken, WithLocality(true))(remote_2)
		Expect(err).ToNot(HaveOccurred())

		responses := 0
		for i := 0; i < iterations; i++ {
			stdout, _, err := remote_2.ExecWithRetries("", "", "demo-client",
				"curl", "-v", "-m", "3", "--fail", "localhost:4001")
			Expect(err).ToNot(HaveOccurred())
			Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
			Expect(stdout).To(ContainSubstring("universal"))

			if strings.Contains(stdout, "universal2") {
				responses++
			}
		}
		Expect(responses).To(Equal(iterations))
	})
})
