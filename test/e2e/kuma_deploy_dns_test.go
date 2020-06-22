package e2e_test

import (
	"strings"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/Kong/kuma/test/framework"
)

var _ = XDescribe("Test DNS", func() {

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

		err = clusters.InjectDNS()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := clusters.DeleteNamespace(TestNamespace)
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()
	})

	It("Should resolve with two apps", func() {
		// given
		c := clusters.GetCluster(Kuma1)

		// when
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

		// then
		out, err := k8s.RunKubectlAndGetOutputE(c.GetTesting(),
			c.GetKubectlOptions(TestNamespace),
			"exec", clientPod.GetName(),
			"-c", "client", "--", "getent", "hosts", "example-app")
		Expect(err).ToNot(HaveOccurred())
		svcIP := strings.Split(out, " ")[0]

		// and
		retry.DoWithRetry(c.GetTesting(), "resolve example-app.mesh",
			defaultRetries, defaultTimeout,
			func() (string, error) {
				out, err = k8s.RunKubectlAndGetOutputE(c.GetTesting(),
					c.GetKubectlOptions(TestNamespace),
					"exec", clientPod.GetName(),
					"-c", "client", "--", "getent", "hosts", "example-app.mesh")
				return out, err
			})

		virtualIP := strings.Split(out, " ")[0]

		// and
		Expect(virtualIP).To(ContainSubstring("240.0.0"))
		Expect(virtualIP).ToNot(Equal(svcIP))
	})
})
