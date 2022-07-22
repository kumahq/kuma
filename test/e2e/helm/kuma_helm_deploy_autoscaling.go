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

	var setup = func(withCni KumaDeploymentOption) {
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
				withCni,
			)).
			Setup(cluster)
		Expect(err).ToNot(HaveOccurred())
	}

	E2EAfterEach(func() {
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	DescribeTable(
		"Should scale to the minimum replicas",
		func(withCni KumaDeploymentOption) {
			setup(withCni)
			// given a kuma deployment with autoscaling and min replicas
			k8sCluster := cluster.(*K8sCluster)

			// when waiting for autoscaling
			err := k8sCluster.WaitApp(Config.KumaServiceName, Config.KumaNamespace, minReplicas)

			// then the min replicas should come up
			Expect(err).ToNot(HaveOccurred())
		},
		Entry("with default cni", WithCNI()),
		Entry("with new cni (experimental)", WithExperimentalCNI()),
	)
}
