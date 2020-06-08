package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test Local and Global", func() {
	It("Should deploy Local and Global on Two K8s cluster.", func() {
		clusters, err := framework.NewK8sClusters(
			[]string{framework.Kuma1, framework.Kuma2},
			framework.Silent)
		Expect(err).ToNot(HaveOccurred())

		c1 := clusters.GetCluster(framework.Kuma1)
		c2 := clusters.GetCluster(framework.Kuma2)

		err = c1.DeployKuma("global")
		Expect(err).ToNot(HaveOccurred())

		err = c2.DeployKuma("local")
		Expect(err).ToNot(HaveOccurred())

		err = c1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = c2.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		logs1, err := c1.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs1).To(ContainSubstring("\"mode\":\"global\""))

		logs2, err := c2.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		Expect(logs2).To(ContainSubstring("\"mode\":\"local\""))

		_ = clusters.DeleteKuma()

	})
})
