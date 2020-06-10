package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test Local and Global", func() {
	var clusters framework.Clusters

	BeforeEach(func() {
		var err error
		clusters, err = framework.NewK8sClusters(
			[]string{framework.Kuma1, framework.Kuma2},
			framework.Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.CreateNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = clusters.LabelNamespaceForSidecarInjection("kuma-test")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should deploy on Two K8s cluster and verify Kuma", func() {
		// given
		err := clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		// when
		err = clusters.VerifyKuma()

		//then
		Expect(err).ToNot(HaveOccurred())
	})

	It("Should deploy Local and Global on 2 clusters", func() {
		// given
		c1 := clusters.GetCluster(framework.Kuma1)
		c2 := clusters.GetCluster(framework.Kuma2)

		err := c1.DeployKuma("global")
		Expect(err).To(HaveOccurred())

		err = c2.DeployKuma("local")
		Expect(err).ToNot(HaveOccurred())

		// when
		err = clusters.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		// then
		logs1, err := c1.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs1).To(ContainSubstring("\"mode\":\"global\""))

		logs2, err := c2.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs2).To(ContainSubstring("\"mode\":\"local\""))
	})
})
