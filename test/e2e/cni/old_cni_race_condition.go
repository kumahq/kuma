package cni

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AppDeploymentWithCniAndNoTaintController() {
	defaultMesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
`

	var cluster Cluster
	var k8sCluster *K8sCluster

	var setup = func() {
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
				WithSkipDefaultMesh(true), // it's common case for HELM deployments that Mesh is also managed by HELM therefore it's not created by default
				WithHelmOpt("cni.delayStartupSeconds", "40"),
				WithHelmOpt("experimental.cni", "false"),
				WithCNI(),
			)).
			Install(YamlK8s(defaultMesh)).
			Setup(cluster)
		// here we could patch the "command" of the CNI, kubectl patch ...
		Expect(err).ToNot(HaveOccurred())
	}

	E2EAfterEach(func() {
		Expect(cluster.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
		Expect(k8sCluster.DeleteNode("k3d-third-node-0")).To(Succeed())
	})

	It(
		"is susceptible to the race condition",
		func() {
			setup()

			// k3s1 v1.19.16 hangs if the name is the same in the previous test
			err := k8sCluster.CreateNode("third-node", "second=true")
			Expect(err).ToNot(HaveOccurred())

			err = k8sCluster.LoadImages("kuma-dp", "kuma-universal")
			Expect(err).ToNot(HaveOccurred())

			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(TestNamespace)).
				Install(testserver.Install(func(opts *testserver.DeploymentOpts) {
					opts.NodeSelector = map[string]string{
						"second": "true",
					}
				})).
				Setup(cluster)

			// test-server probe will fail without iptables rules applied
			Expect(err).Should(HaveOccurred())
			_, errorIsOfTypeMaxRetriesExceeded := err.(retry.MaxRetriesExceeded)
			Expect(errorIsOfTypeMaxRetriesExceeded).To(Equal(true))
		},
	)
}
