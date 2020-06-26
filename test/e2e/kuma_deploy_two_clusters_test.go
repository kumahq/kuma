package e2e_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/Kong/kuma/test/framework"
)

var _ = XDescribe("Test Remote and Global", func() {
	var clusters Clusters

	BeforeEach(func() {
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1, Kuma2},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.CreateNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.LabelNamespaceForSidecarInjection(TestNamespace)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should deploy on Two K8s cluster and verify Kuma", func() {
		// when
		_, err := clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		// then
		err = clusters.VerifyKuma()
		//then
		Expect(err).ToNot(HaveOccurred())
	})
})
