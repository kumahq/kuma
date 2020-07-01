package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config/mode"

	. "github.com/Kong/kuma/test/framework"
)

var _ = FDescribe("Test Universal deployment", func() {

	var global, remote_1, remote_2 Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2, Kuma3},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		global = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma(mode.Global)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		remote_1 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(EchoServerUniversal()).
			Install(DemoClientUniversal()).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		remote_2 = clusters.GetCluster(Kuma3)

		err = NewClusterSetup().
			Install(Kuma(mode.Remote)).
			Install(EchoServerUniversal()).
			Install(DemoClientUniversal()).
			Setup(remote_2)
		Expect(err).ToNot(HaveOccurred())
		err = remote_2.VerifyKuma()
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
		//Expect(true).To(BeTrue())
		_, stderr, err := remote_1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))

		_, stderr, err = remote_2.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
})
