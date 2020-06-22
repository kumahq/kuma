package e2e_test

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/Kong/kuma/test/framework"
)

var _ = XDescribe("Test K8s deployment with `kumactl install control-plane`", func() {

	var clusters Clusters

	BeforeEach(func() {
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.CreateNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.LabelNamespaceForSidecarInjection(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		_, err = clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = clusters.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should check Kuma side-car injection", func() {
		// given
		c := clusters.GetCluster(Kuma1)

		// when
		err := c.DeployApp(TestNamespace, "example-app")
		Expect(err).ToNot(HaveOccurred())

		appPods, err := k8s.ListPodsE(c.GetTesting(),
			c.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: "app=example-app",
			})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(appPods)).To(Equal(1))

		appPod := appPods[0]

		// then
		Expect(func() bool {
			for _, c := range appPod.Spec.Containers {
				if c.Name == "kuma-sidecar" {
					return true
				}
			}
			return false
		}()).To(Equal(true))
	})
})
