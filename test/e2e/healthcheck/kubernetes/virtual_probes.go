package kubernetes

import (
	"fmt"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func VirtualProbes() {
	var k8sCluster Cluster
	var optsKubernetes = KumaK8sDeployOpts

	E2EBeforeSuite(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1}, Silent)
		Expect(err).ToNot(HaveOccurred())

		k8sCluster = k8sClusters.GetCluster(Kuma1)

		Expect(Kuma(config_core.Standalone, optsKubernetes...)(k8sCluster)).To(Succeed())
		Expect(NamespaceWithSidecarInjection(TestNamespace)(k8sCluster)).To(Succeed())
		Expect(k8sCluster.VerifyKuma()).To(Succeed())
	})

	E2EAfterSuite(func() {
		Expect(k8sCluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(k8sCluster.DeleteKuma(optsKubernetes...)).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})

	PollPodsReady := func(name string, namespace string) error {
		pods, err := k8s.ListPodsE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions(namespace),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app=%s", name)})
		if err != nil {
			return err
		}
		for _, p := range pods {
			err := WaitUntilPodReadyE(k8sCluster.GetTesting(), k8sCluster.GetKubectlOptions(namespace), p.GetName(), 0, 0)
			if err != nil {
				return err
			}
		}
		return nil
	}

	It("should deploy test-server with probes", func() {
		Expect(testserver.Install()(k8sCluster)).To(Succeed())

		// The testserver install func also does this, but we
		// repeat is here to make the deployment test criteria
		// clearer and to make the test robust to framework changes.
		opts := testserver.DefaultDeploymentOpts()

		// Sample pod readiness to ensure they stay ready to at least 10sec.
		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			Expect(PollPodsReady(opts.Namespace, opts.Name)).To(Succeed())
		}
	})
}
