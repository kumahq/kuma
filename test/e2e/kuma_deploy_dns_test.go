package e2e_test

import (
	"path/filepath"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Kong/kuma/test/framework"
)

var _ = Describe("Test DNS", func() {

	It("Should inject .kuma resolver", func() {
		clusters, err := framework.NewK8sClusters(
			[]string{framework.Kuma1},
			framework.Verbose)
		Expect(err).ToNot(HaveOccurred())
		c := clusters.GetCluster(framework.Kuma1)

		err = c.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = c.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = c.InjectDNS()
		Expect(err).ToNot(HaveOccurred())

		err = k8s.CreateNamespaceE(c.GetTesting(), c.GetKubectlOptions(), "kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = c.LabelNamespaceForSidecarInjection("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = k8s.KubectlApplyE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			filepath.Join("testdata", "example-app-svc.yaml"))
		Expect(err).ToNot(HaveOccurred())

		k8s.WaitUntilServiceAvailable(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			"example-app", defaultRetries, defaultTimeout)

		err = k8s.KubectlApplyE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			filepath.Join("testdata", "example-app.yaml"))
		Expect(err).ToNot(HaveOccurred())

		k8s.WaitUntilNumPodsCreated(c.GetTesting(),
			c.GetKubectlOptions(),
			metav1.ListOptions{
				LabelSelector: "app=example-app",
			},
			1, defaultRetries, defaultTimeout)

		err = k8s.KubectlApplyE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			filepath.Join("testdata", "example-client-svc.yaml"))
		Expect(err).ToNot(HaveOccurred())

		k8s.WaitUntilServiceAvailable(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			"example-client", defaultRetries, defaultTimeout)

		err = k8s.KubectlApplyE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			filepath.Join("testdata", "example-client.yaml"))
		Expect(err).ToNot(HaveOccurred())

		k8s.WaitUntilNumPodsCreated(c.GetTesting(),
			c.GetKubectlOptions(),
			metav1.ListOptions{
				LabelSelector: "app=example-client",
			},
			1, defaultRetries, defaultTimeout)

		clientPods := k8s.ListPods(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			metav1.ListOptions{
				LabelSelector: "app=example-client",
			})
		Expect(len(clientPods)).To(Equal(1))

		clientPod := clientPods[0]

		k8s.WaitUntilPodAvailable(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			clientPod.GetName(),
			defaultRetries, defaultTimeout)

		out, err := k8s.RunKubectlAndGetOutputE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			"exec", clientPod.GetName(), "--", "nslookup", "example-app.kuma")
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(ContainSubstring("240.0.0"))

		err = k8s.DeleteNamespaceE(c.GetTesting(), c.GetKubectlOptions(), "kuma-test")
		Expect(err).ToNot(HaveOccurred())

		_ = clusters.DeleteKuma()

	})
})
