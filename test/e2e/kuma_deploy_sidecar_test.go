package e2e_test

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test K8s deployment with `kumactl install control-plane`", func() {

	var clusters framework.Clusters

	BeforeEach(func() {
		var err error
		clusters, err = framework.NewK8sClusters(
			[]string{framework.Kuma1},
			framework.Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.CreateNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = clusters.LabelNamespaceForSidecarInjection("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		_, err = clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = clusters.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should check Kuma side-car injection", func() {
		// given
		c := clusters.GetCluster(framework.Kuma1)

		// when
		err := c.DeployApp("kuma-test", "example-app")
		Expect(err).ToNot(HaveOccurred())

		appPods, err := k8s.ListPodsE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
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
