package e2e_test

import (
	"fmt"
	"path/filepath"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gruntwork-io/terratest/modules/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Kong/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Test K8s deployment with `kumactl install control-plane`", func() {

	It("Should deploy on Single K8s cluster and verify Kuma.", func(done Done) {
		clusters, err := framework.NewK8sClusters(
			[]string{framework.Kuma1},
			framework.Silent)
		Expect(err).ToNot(HaveOccurred())
		c := clusters.GetCluster(framework.Kuma1)

		err = c.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = c.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		logs, err := c.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(logs)

		_ = c.DeleteKuma()

		// completed
		close(done)
	}, 180)

	It("Should deploy on Two K8s cluster and verify Kuma.", func(done Done) {
		clusters, err := framework.NewK8sClusters(
			[]string{framework.Kuma1, framework.Kuma2},
			framework.Silent)
		Expect(err).ToNot(HaveOccurred())

		err = clusters.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = clusters.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		logs, err := clusters.GetKumaCPLogs()
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(logs)

		_ = clusters.DeleteKuma()

		// completed
		close(done)
	}, 180)

	It("Should check Kuma side-car injection", func(done Done) {
		clusters, err := framework.NewK8sClusters(
			[]string{framework.Kuma1},
			framework.Verbose)
		Expect(err).ToNot(HaveOccurred())
		c := clusters.GetCluster(framework.Kuma1)

		err = c.DeployKuma()
		Expect(err).ToNot(HaveOccurred())

		err = c.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		err = k8s.CreateNamespaceE(c.GetTesting(), c.GetKubectlOptions(), "kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = c.LabelNamespaceForSidecarInjection("kuma-test")
		Expect(err).ToNot(HaveOccurred())

		err = k8s.KubectlApplyE(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			filepath.Join("testdata", "example-app.yaml"))
		Expect(err).ToNot(HaveOccurred())

		k8s.WaitUntilServiceAvailable(c.GetTesting(),
			c.GetKubectlOptions("kuma-test"),
			"example-app", 10, 3*time.Second)

		k8s.WaitUntilNumPodsCreated(c.GetTesting(),
			c.GetKubectlOptions(),
			metav1.ListOptions{
				LabelSelector: "app=example-app",
			},
			1, 10, 3*time.Second)

		pods, err := k8s.ListPodsE(c.GetTesting(), c.GetKubectlOptions("kuma-test"),
			v1.ListOptions{
				LabelSelector: "app=example-app",
			})
		Expect(err).ToNot(HaveOccurred())
		Expect(len(pods)).To(Equal(1))
		Expect(func() bool {
			for _, c := range pods[0].Spec.Containers {
				if c.Name == "kuma-sidecar" {
					return true
				}
			}
			return false
		}()).To(Equal(true))

		_ = c.DeleteKuma()

		// completed
		close(done)
	}, 180)
})
