package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Universal deployment", func() {

	var c1 Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		c1 = clusters.GetCluster(Kuma1)

		err = NewClusterSetup().
			Install(Kuma()).
			Install(EchoServerUniversal()).
			Install(DemoClientUniversal()).
			Setup(c1)
		Expect(err).ToNot(HaveOccurred())
		err = c1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		_ = c1.DeleteKuma()
		_ = c1.DismissCluster()
	})

	It("Should deploy two apps", func() {
		_, stderr, err := c1.ExecWithRetries("", "", "demo-client",
			"curl", "-v", "-m", "3", "localhost:4000")
		Expect(err).ToNot(HaveOccurred())
		Expect(stderr).To(ContainSubstring("HTTP/1.1 200 OK"))
	})
})
