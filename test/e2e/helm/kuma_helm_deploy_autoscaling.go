package helm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

func ControlPlaneAutoscalingWithHelmChart() {
	minReplicas := 3

	var cluster Cluster
	var deployOptsFuncs = KumaK8sDeployOpts

	BeforeEach(func() {
		c, err := NewK8sClusterWithTimeout(
			NewTestingT(),
			Kuma1,
			Silent,
			6*time.Second)
		Expect(err).ToNot(HaveOccurred())

		cluster = c.WithRetries(60)

		releaseName := fmt.Sprintf(
			"kuma-%s",
			strings.ToLower(random.UniqueId()),
		)
		deployOptsFuncs = append(deployOptsFuncs,
			WithInstallationMode(HelmInstallationMode),
			WithHelmReleaseName(releaseName),
			WithHelmOpt("controlPlane.autoscaling.enabled", "true"),
			WithHelmOpt("controlPlane.autoscaling.minReplicas", strconv.Itoa(minReplicas)),
			WithCNI())

		err = NewClusterSetup().
			Install(Kuma(core.Standalone, deployOptsFuncs...)).
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
}
