package cni

import (
	"fmt"
	"strings"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func AppDeploymentWithCniAndTaintController() {
	defaultMesh := `
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: default
`

	var cluster Cluster
	var k8sCluster *K8sCluster
	nodeName := fmt.Sprintf(
		"second-%s",
		strings.ToLower(random.UniqueId()),
	)

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
				WithExperimentalCNI(),
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
		Expect(k8sCluster.DeleteNode("k3d-" + nodeName + "-0")).To(Succeed())
	})

	It(
		"prevents race condition following a slow CNI start",
		func() {
			setup()

			err := k8sCluster.CreateNode(nodeName, "second=true")
			Expect(err).ToNot(HaveOccurred())

			err = k8sCluster.LoadImages("kuma-dp", "kuma-cni", "kuma-universal")
			Expect(err).ToNot(HaveOccurred())

			err = NewClusterSetup().
				Install(NamespaceWithSidecarInjection(TestNamespace)).
				Install(testserver.Install(func(opts *testserver.DeploymentOpts) {
					opts.NodeSelector = map[string]string{
						"second": "true",
					}
				})).
				Install(DemoClientK8sWithAffinity("default", TestNamespace)).
				Setup(cluster)
			Expect(err).ToNot(HaveOccurred())

			// assert pods demo-client and testserver are available on the node
			clientPodName, err := PodNameOfApp(cluster, "demo-client", TestNamespace)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() (string, error) {
				_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server")
				return stderr, err
			}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

			Eventually(func() (string, error) {
				_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server_kuma-test_svc_80.mesh")
				return stderr, err
			}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))

			Eventually(func() (string, error) { // should access a service with . instead of _
				_, stderr, err := cluster.ExecWithRetries(TestNamespace, clientPodName, "demo-client",
					"curl", "-v", "-m", "3", "--fail", "test-server.kuma-test.svc.80.mesh")
				return stderr, err
			}, "10s", "1s").Should(ContainSubstring("HTTP/1.1 200 OK"))
		},
	)
}
