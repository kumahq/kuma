package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config/mode"

	. "github.com/Kong/kuma/test/framework"
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
   service: "*"
destinations:
- match:
   service: "*"
`
	var global, remote_1, remote_2 Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2, Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(EchoServerUniversal()).
			Install(DemoClientUniversal()).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Cluster 2
		remote_2 = clusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(DemoClientUniversal()).
			Setup(remote_2)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(mode.Global)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()
		remote_1CP := remote_1.GetKuma()
		remote_2CP := remote_2.GetKuma()
		err = globalCP.AddCluster(remote_1CP.GetName(),
			globalCP.GetKDSServerAddress(), remote_1CP.GetKDSServerAddress(), remote_1CP.GetIngressAddress())
		Expect(err).ToNot(HaveOccurred())
		err = globalCP.AddCluster(remote_2CP.GetName(),
			globalCP.GetKDSServerAddress(), remote_2CP.GetKDSServerAddress(), remote_2CP.GetIngressAddress())
		Expect(err).ToNot(HaveOccurred())

		err = global.RestartKuma()
		Expect(err).ToNot(HaveOccurred())

		// remove these once Zones are added dynamically
		err = YamlUniversal(meshDefaulMtlsOn)(global)
		Expect(err).ToNot(HaveOccurred())

		err = YamlUniversal(trafficPermissionAll)(global)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = global.DeleteKuma()
		_ = remote_1.DeleteKuma()
		_ = remote_2.DeleteKuma()
		_ = global.DismissCluster()
		_ = remote_1.DismissCluster()
		_ = remote_2.DismissCluster()
	})

	It("Should deploy two apps", func() {
		stdout, _, err := remote_1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))

		stdout, _, err = remote_2.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
})
