package e2e_test

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Kubernetes/Universal deployment when Global is on K8S", func() {

	var globalCluster, remoteCluster Cluster
	var optsGlobal, optsRemote []DeployOptionsFunc

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		globalCluster = k8sClusters.GetCluster(Kuma1)
		optsGlobal = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())
		err = globalCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
		globalCP := globalCluster.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())
		ingressToken, err := globalCP.GenerateDpToken("default", "ingress")
		Expect(err).ToNot(HaveOccurred())

		// Remote
		remoteCluster = universalClusters.GetCluster(Kuma3)
		optsRemote = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote...)).
			Install(EchoServerUniversal(AppModeEchoServer, "default", "universal", echoServerToken)).
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken)).
			Install(IngressUniversal("default", ingressToken)).
			Setup(remoteCluster)
		Expect(err).ToNot(HaveOccurred())
		err = remoteCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := globalCluster.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())
		err = globalCluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remoteCluster.DeleteKuma(optsRemote...)
		Expect(err).ToNot(HaveOccurred())
		err = remoteCluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("communication in between apps in remote zone works", func() {
		stdout, _, err := remoteCluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "localhost:4001")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		retry.DoWithRetry(remoteCluster.GetTesting(), "curl remote service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err = remoteCluster.ExecWithRetries("", "", "demo-client",
					"curl", "-v", "-m", "3", "localhost:4001")
				if err != nil {
					return "should retry", err
				}
				if strings.Contains(stdout, "HTTP/1.1 200 OK") {
					return "Accessing service successful", nil
				}
				return "should retry", errors.Errorf("should retry")
			})
	})
})
