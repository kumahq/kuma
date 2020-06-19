package e2e_test

import (
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test App deployment", func() {

	var clusters Clusters

	BeforeEach(func() {
		var err error
		clusters, err = NewK8sClusters(
			[]string{Kuma1},
			Verbose)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.CreateNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = clusters.LabelNamespaceForSidecarInjection("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		_, err = clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = clusters.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = clusters.InjectDNS()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should deploy two apps", func() {
		// setup
		c := clusters.GetCluster(Kuma1)

		// given
		err := c.DeployApp(TestNamespace, "example-app")
		Expect(err).ToNot(HaveOccurred())

		err = c.DeployApp(TestNamespace, "example-client")
		Expect(err).ToNot(HaveOccurred())

		clientPods := k8s.ListPods(c.GetTesting(),
			c.GetKubectlOptions(TestNamespace),
			metav1.ListOptions{
				LabelSelector: "app=example-client",
			})
		Expect(len(clientPods)).To(Equal(1))

		clientPod := clientPods[0]

		k8s.WaitUntilPodAvailable(c.GetTesting(),
			c.GetKubectlOptions(TestNamespace),
			clientPod.GetName(),
			defaultRetries, defaultTimeout)

		// when
		out, err := k8s.RunKubectlAndGetOutputE(c.GetTesting(),
			c.GetKubectlOptions(TestNamespace),
			"exec", clientPod.GetName(), "--", "/usr/bin/curl", "example-app")
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(out).To(ContainSubstring("Thank you for using nginx."))

		retry.DoWithRetry(c.GetTesting(), "resolve example-app.mesh",
			defaultRetries, defaultTimeout,
			func() (string, error) {
				out, err = k8s.RunKubectlAndGetOutputE(c.GetTesting(),
					c.GetKubectlOptions("kuma-test"),
					"exec", clientPod.GetName(),
					"-c", "client", "--", "getent", "hosts", "example-app.mesh")
				return out, err
			})

		// when
		out, err = k8s.RunKubectlAndGetOutputE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			"exec", clientPod.GetName(), "--", "/usr/bin/curl", "example-app.mesh")
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(out).To(ContainSubstring("Thank you for using nginx."))
	})
})
