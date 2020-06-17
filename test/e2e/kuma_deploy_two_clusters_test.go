package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/test/framework"
)

var _ = XDescribe("Test Local and Global", func() {
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
		// when
		err := clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		// then
		err = clusters.VerifyKuma()
		//then
		Expect(err).ToNot(HaveOccurred())
	})
})
