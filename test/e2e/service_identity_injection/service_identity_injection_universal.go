package service_identity_injection

import (
	"strings"

	"github.com/go-errors/errors"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ServiceIdentityInjectionUniversal() {
	meshMTLSOn := func() string {
		return `
type: Mesh
name: default
mtls:
  enabledBackend: ca-1
  backends:
  - name: ca-1
    type: builtin
`
	}

	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma3)
		deployOptsFuncs = KumaUniversalDeployOpts

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(YamlUniversal(meshMTLSOn())).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken)).
			Install(EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := cluster.DeleteKuma(deployOptsFuncs...)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("server should receive X-Kuma-Forwarded-Client-* headers", func() {
		retry.DoWithRetry(cluster.GetTesting(), "curl local service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err := cluster.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "--fail", "localhost:4001")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})

		_, stderr, err := cluster.Exec("", "", "demo-client", "curl", "-v", "-m", "8", "--fail", "localhost:4001")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(BeEmpty())

		appStdout := cluster.(*UniversalCluster).GetApp(AppModeEchoServer).GetMainApp().Out()
		Expect(appStdout).To(ContainSubstring("X-Kuma-Forwarded-Client-Cert: spiffe://default/demo-client"))
		Expect(appStdout).To(ContainSubstring("X-Kuma-Forwarded-Client-Service: demo-client"))
		Expect(appStdout).To(ContainSubstring("X-Kuma-Forwarded-Client-Zone: us-east-1"))
	})
}
