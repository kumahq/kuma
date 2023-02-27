package cni

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	util_k8s "github.com/kumahq/kuma/pkg/plugins/runtime/k8s/util"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AppDeploymentWithCniAndNoTaintController() {
	var cluster Cluster
	var k8sCluster *K8sCluster

	BeforeEach(func() {
		k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)
		cluster = k8sCluster.
			WithTimeout(6 * time.Second).
			WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)

		err := NewClusterSetup().
			Install(Kuma(core.Standalone,
				WithInstallationMode(HelmInstallationMode),
				WithHelmReleaseName(releaseName),
				WithHelmOpt("cni.delayStartupSeconds", "40000"),
				WithHelmOpt("experimental.cni", "false"),
				WithCNIV1(),
			)).
			Setup(cluster)
		// here we could patch the "command" of the CNI, kubectl patch ...
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It(
		"is susceptible to the race condition",
		func() {
			// given a non-healthy CNI

			// when test server is deployed without working CNI
			err := NewClusterSetup().
				Install(NamespaceWithSidecarInjection(TestNamespace)).
				Install(testserver.Install(testserver.WithoutWaitingToBeReady())).
				Setup(cluster)

			// then
			Expect(err).ShouldNot(HaveOccurred())
			podName, err := PodNameOfApp(k8sCluster, "test-server", TestNamespace)
			Expect(err).ToNot(HaveOccurred())

			// and DP received config
			Eventually(func(g Gomega) {
				received, err := DataplaneReceivedConfig(k8sCluster, "default", fmt.Sprintf("%s.%s", podName, TestNamespace))
				g.Expect(err).ToNot(HaveOccurred())
				g.Expect(received).To(BeTrue())
			}, "30s", "1s").Should(Succeed())

			// and test-server container in the pod is unhealthy (probe fail without iptables rules applied)
			Consistently(func(g Gomega) {
				pod, err := k8s.GetPodE(cluster.GetTesting(), cluster.GetKubectlOptions(TestNamespace), podName)

				g.Expect(err).ToNot(HaveOccurred())
				status := util_k8s.FindContainerStatus(pod, "test-server")
				g.Expect(status).ToNot(BeNil())
				g.Expect(status.Ready).To(BeFalse())
			}, "10s", "1s").Should(Succeed())
		},
	)
}
