package e2e_test

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

var _ = Describe("Test Control Plane autoscaling with Helm chart", func() {
	minReplicas := 3

	var cluster Cluster
	var deployOptsFuncs []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewK8sClusters(
			[]string{Kuma1},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		cluster = clusters.GetCluster(Kuma1)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)
		deployOptsFuncs = []DeployOptionsFunc{
			WithInstallationMode(HelmInstallationMode),
			WithHelmReleaseName(releaseName),
			WithHelmOpt("controlPlane.autoscaling.enabled", "true"),
			WithHelmOpt("controlPlane.autoscaling.minReplicas", strconv.Itoa(minReplicas)),
			WithCNI(),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
			Install(KumaDNS()).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		// tear down Kuma
		Expect(cluster.DeleteKuma(deployOptsFuncs...)).To(Succeed())
		// tear down cluster
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("Should scale to the minimum replicas", func() {
		// given a kuma deployment with autoscaling and min replicas
		k8sCluster := cluster.(*K8sCluster)

		// when waiting for autoscaling
		err := k8sCluster.WaitApp(KumaServiceName, KumaNamespace, minReplicas)

		// then the min replicas should come up
		Expect(err).ToNot(HaveOccurred())
	})
})
