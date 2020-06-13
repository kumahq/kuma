package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/pkg/config/core"

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

	It("Should deploy Local and Global on 2 clusters", func() {
		// given
		c1 := clusters.GetCluster(framework.Kuma1)
		c2 := clusters.GetCluster(framework.Kuma2)

		err := c1.DeployKuma(core.Global)
		Expect(err).ToNot(HaveOccurred())

		err = c2.DeployKuma(core.Local)
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c1.VerifyKuma()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		err = c2.VerifyKuma()
		// then
		Expect(err).To(HaveOccurred())

		// then
		logs1, err := c1.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs1).To(ContainSubstring("\"mode\":\"global\""))

		logs2, err := c2.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs2).To(ContainSubstring("\"mode\":\"local\""))
	})
})
