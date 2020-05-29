// +build k8s,integration

package cmd_test

import (
	"testing"

	"github.com/Kong/kuma/tools/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegrationCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration CMD Suite")
}

var _ = Describe("Test K8s deployment with `kumactl install control-plane`", func() {

	It("Deploy on Single K8s cluster and verify the Kuma CP REST API is accessible", func(done Done) {
		t := framework.NewK8sTest(1, "", framework.Verbose)

		err := t.DeployKumaOnK8sClusterE(1)
		Expect(err).ToNot(HaveOccurred())

		err = t.VerifyKumaOnK8sClusterE()
		Expect(err).ToNot(HaveOccurred())

		_ = t.DeleteKumaOnK8sClusterE(1)
		_ = t.DeleteKumaNamespaceOnK8sClusterE(1)

		// completed
		close(done)
	}, 180)

	//It("Deploy on Two K8s clusters and verify the Kuma CP REST API is accessible", func() {
	//	t := framework.NewK8sTest(2, "", framework.Silent)
	//
	//	err := t.DeployKumaOnK8sClusterE(1)
	//	Expect(err).ToNot(HaveOccurred())
	//	err = t.DeployKumaOnK8sClusterE(2)
	//	Expect(err).ToNot(HaveOccurred())
	//
	//	err = t.VerifyKumaOnK8sClusterE(1)
	//	Expect(err).ToNot(HaveOccurred())
	//	err = t.VerifyKumaOnK8sClusterE(2)
	//	Expect(err).ToNot(HaveOccurred())
	//
	//	_ = t.DeleteKumaOnK8sClusterE(1)
	//	_ = t.DeleteKumaOnK8sClusterE(2)
	//	_ = t.DeleteKumaNamespaceOnK8sClusterE(1)
	//	_ = t.DeleteKumaNamespaceOnK8sClusterE(2)
	//}, 90)
})
