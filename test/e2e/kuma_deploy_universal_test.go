package e2e_test

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/config/core"

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

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2, Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken(AppModeEchoServer)
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken(AppModeDemoClient)
		Expect(err).ToNot(HaveOccurred())

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma(core.Remote, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(EchoServerUniversal(echoServerToken)).
			Install(DemoClientUniversal(demoClientToken)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Cluster 2
		remote_2 = clusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(core.Remote, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(DemoClientUniversal(demoClientToken)).
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
		err := remote_1.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = global.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())

		err = remote_1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should deploy two apps", func() {
		stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4001")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		retry.DoWithRetry(remote_2.GetTesting(), "curl remote service",
			DefaultRetries, DefaultTimeout,
			func() (string, error) {
				stdout, _, err = remote_2.ExecWithRetries("", "", "demo-client",
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
