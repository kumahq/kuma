package globalkubernetes

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func KubernetesUniversalDeploymentWhenGlobalIsOnK8S() {
	var globalCluster, zoneCluster Cluster

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

		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(globalCluster)
		Expect(err).ToNot(HaveOccurred())
		globalCP := globalCluster.GetKuma()

		echoServerToken, err := globalCP.GenerateDpToken("default", "test-server")
		Expect(err).ToNot(HaveOccurred())
		demoClientToken, err := globalCP.GenerateDpToken("default", "demo-client")
		Expect(err).ToNot(HaveOccurred())

		// Zone
		zoneCluster = universalClusters.GetCluster(Kuma3)
		ingressTokenKuma3, err := globalCP.GenerateZoneIngressToken(Kuma3)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Install(TestServerUniversal("test-server", "default", echoServerToken, WithArgs([]string{"echo", "--instance", "universal-1"}))).
			Install(DemoClientUniversal(AppModeDemoClient, "default", demoClientToken, WithTransparentProxy(true))).
			Install(IngressUniversal(ingressTokenKuma3)).
			Setup(zoneCluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		err := globalCluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = globalCluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = zoneCluster.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zoneCluster.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("communication in between apps in zone works", func() {
		stdout, _, err := zoneCluster.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "--fail", "test-server.mesh")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
}
