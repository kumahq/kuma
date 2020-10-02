package e2e

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Kubernetes/Universal deployment when Global is on K8S", func() {

	var globalCluster, remoteCluster Cluster

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters(
			[]string{Kuma1},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		universalClusters, err := NewUniversalClusters(
			[]string{Kuma3},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		// Global
		globalCluster = k8sClusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())
		err = globalCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
		globalCP := globalCluster.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken(AppModeEchoServer)
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken(AppModeDemoClient)
		Expect(err).ToNot(HaveOccurred())

		// Remote
		remoteCluster = universalClusters.GetCluster(Kuma3)
		err = NewClusterSetup().
			Install(Kuma(core.Remote, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(EchoServerUniversal(echoServerToken)).
			Install(DemoClientUniversal(demoClientToken)).
			Setup(remoteCluster)
		Expect(err).ToNot(HaveOccurred())
		err = remoteCluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// connect Remote with Global
		err = k8s.KubectlApplyFromStringE(globalCluster.GetTesting(), globalCluster.GetKubectlOptions(),
			fmt.Sprintf(ZoneTemplateK8s,
				Kuma3,
				remoteCluster.GetKuma().GetIngressAddress()))
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := globalCluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = globalCluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = remoteCluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = remoteCluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("communication in between apps in remote zone works", func() {
		stdout, _, err := remoteCluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4001")
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
