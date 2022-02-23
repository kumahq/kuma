package helm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ControlPlaneAutoscalingWithHelmChart() {
	minReplicas := 3

	var cluster Cluster

	BeforeEach(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent).
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
				WithHelmOpt("controlPlane.autoscaling.enabled", "true"),
				WithHelmOpt("controlPlane.autoscaling.minReplicas", strconv.Itoa(minReplicas)),
				WithCNI(),
			)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if ShouldSkipCleanup() {
			return
		}
		// tear down Kuma
		Expect(cluster.DeleteKuma()).To(Succeed())
		// tear down cluster
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("Should scale to the minimum replicas", func() {
		// given a kuma deployment with autoscaling and min replicas
		k8sCluster := cluster.(*K8sCluster)

		// when waiting for autoscaling
		err := k8sCluster.WaitApp(Config.KumaServiceName, Config.KumaNamespace, minReplicas)

		// then the min replicas should come up
		Expect(err).ToNot(HaveOccurred())
	})
}
