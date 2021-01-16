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

var _ = Describe("Test Retry on Universal", func() {
	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		cluster = clusters.GetCluster(Kuma1)
		deployOptsFuncs = []DeployOptionsFunc{}

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())

		err = cluster.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		demoClientToken, err := cluster.GetKuma().GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		echoServerToken, err := cluster.GetKuma().GenerateDpToken("default", "echo-server_kuma-test_svc_8080")
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(DemoClientUniversal("default", demoClientToken)).
			Install(EchoServerUniversal("universal1", "default", echoServerToken)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	// AfterEach(func() {
	// 	if ShouldSkipCleanup() {
	// 		return
	// 	}
	// 	err := cluster.DeleteKuma(deployOptsFuncs...)
	// 	Expect(err).ToNot(HaveOccurred())
	//
	// 	err = cluster.DismissCluster()
	// 	Expect(err).ToNot(HaveOccurred())
	// })

	FIt("should retry on TCP connection failure", func() {
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
	})
})
